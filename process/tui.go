package process

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57"))

	chatStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	chatLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	// Stats bar styles
	statsBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			MarginBottom(1)

	cpuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

	memStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	uptimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("190")).
			Bold(true)
)

// AIProvider is an interface for different LLM providers
type AIProvider interface {
	Name() string
	Analyze(ctx context.Context, process ProcessInfo) (string, error)
	ValidateAPIKey(ctx context.Context) error
}

type ProviderConfig struct {
	Provider string // "anthropic", "openai", "gemini"
	APIKey   string
}

type AIAnalysisMsg struct {
	PID      int32
	Analysis string
}

type ThrottlePromptMsg struct {
	PID int32
}

type Model struct {
	Processes          []ProcessInfo
	SystemStats        SystemStats
	Table              table.Model
	TextInput          textinput.Model
	SortBy             string
	quitting           bool
	width              int
	height             int
	analysisInProgress bool
	selectedProcessPID int32
	provider           AIProvider
	throttlePrompt     *ThrottlePromptMsg // nil or set to show prompt
	chatActive         bool
}

type TickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// NewModel creates a new TUI model with table and chatbox
func NewModel(procs []ProcessInfo, provider AIProvider) Model {
	// Setup table columns
	columns := []table.Column{
		{Title: "PID", Width: 10},
		{Title: "Name", Width: 25},
		{Title: "CPU%", Width: 12},
		{Title: "Memory%", Width: 12},
	}

	// Create initial rows from processes
	var rows []table.Row
	for _, p := range procs {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.PID),
			truncate(p.Name, 25),
			fmt.Sprintf("%.2f", p.CPU),
			fmt.Sprintf("%.2f", p.Memory),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(20),
		// Enable keyboard navigation
		table.WithFocused(true),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = selectedStyle
	t.SetStyles(s)

	// Setup text input for chat
	ti := textinput.New()
	ti.Placeholder = "Type a message and press Enter to send..."
	ti.CharLimit = 156
	ti.Width = 60

	// Get initial system stats
	stats, _ := getSystemStats()

	return Model{
		Processes:   procs,
		SystemStats: stats,
		Table:       t,
		TextInput:   ti,
		SortBy:      "cpu",
		chatActive:  false,
		provider:    provider,
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.Table.SetHeight(msg.Height - 10)
		m.TextInput.Width = msg.Width - 20

	case tea.KeyMsg:
		// Handle chat mode
		if m.chatActive {
			switch msg.String() {
			case "esc":
				m.chatActive = false
				m.TextInput.Blur()
				return m, nil
			case "enter":
				// Send the message
				msg := m.TextInput.Value()
				if msg != "" {
					// Here you would handle the message
					// For now, just clear the input
					m.TextInput.SetValue("")
					m.TextInput.Placeholder = fmt.Sprintf("Sent: %s", msg)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.TextInput, cmd = m.TextInput.Update(msg)
				return m, cmd
			}
		}

		// Normal mode
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "c":
			m.SortBy = "cpu"
			m.SortProcesses()
			m.updateTableRows()
			return m, nil

		case "m":
			m.SortBy = "memory"
			m.SortProcesses()
			m.updateTableRows()
			return m, nil

		case "K":
			// Kill selected process
			if len(m.Processes) > 0 {
				selectedRow := m.Table.Cursor()
				if selectedRow < len(m.Processes) {
					selectedPID := m.Processes[selectedRow].PID
					killProcess(selectedPID)
					// Refresh processes
					procs, _ := GetProcesses()
					m.Processes = procs
					m.SortProcesses()
					m.updateTableRows()
				}
			}
			return m, nil

		case "t":
			// Activate chat mode (only 't', not Enter to avoid conflict with table selection)
			m.chatActive = true
			m.TextInput.Focus()
			cmds = append(cmds, textinput.Blink)
			return m, nil
		}

	case TickMsg:
		// Update process list every tick
		procs, err := GetProcesses()
		if err == nil {
			m.Processes = procs
			m.SortProcesses()

			stats, _ := getSystemStats()
			m.SystemStats = stats

			for _, p := range m.Processes {
				if p.isHighCPU() && p.AIAnalysis == "" && !m.analysisInProgress {
					m.analysisInProgress = true
					m.selectedProcessPID = p.PID

					return m, func() tea.Msg {
						ctx := context.Background()
						analysis, err := analyzeProcessBehavior(ctx, m.provider, p)
						if err != nil {
							analysis = "Analysis failed: " + err.Error()
						}
						return AIAnalysisMsg{PID: p.PID, Analysis: analysis}
					}
				}
			}

			m.updateTableRows()
		}
		cmds = append(cmds, tickCmd())

	case AIAnalysisMsg:
		// Find the process and update its analysis
		for i := range m.Processes {
			if m.Processes[i].PID == msg.PID {
				m.Processes[i].AIAnalysis = msg.Analysis

				// If it's stuck, show throttle prompt
				if m.Processes[i].CPU > 50 && len(msg.Analysis) > 0 {
					m.throttlePrompt = &ThrottlePromptMsg{PID: msg.PID}
				}
				break
			}
		}
		m.analysisInProgress = false
		return m, tickCmd()

	case ThrottlePromptMsg:
		// User pressed 'y' to throttle
		killProcess(msg.PID) // or throttleProcess(msg.PID, 10)
		m.throttlePrompt = nil
		return m, tickCmd()
	}

	// Update table
	var tableCmd tea.Cmd
	m.Table, tableCmd = m.Table.Update(msg)
	cmds = append(cmds, tableCmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateTableRows() {
	var rows []table.Row
	for _, p := range m.Processes {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.PID),
			truncate(p.Name, 25),
			fmt.Sprintf("%.2f", p.CPU),
			fmt.Sprintf("%.2f", p.Memory),
		})
	}
	m.Table.SetRows(rows)
}

// formatUptime converts seconds to human-readable format
func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// renderStatsBar creates the system stats bar
func (m Model) renderStatsBar() string {
	cpu := cpuStyle.Render(fmt.Sprintf("CPU: %.1f%%", m.SystemStats.CPUPercent))
	mem := memStyle.Render(fmt.Sprintf("MEM: %.1f%%", m.SystemStats.MemPercent))
	uptime := uptimeStyle.Render(fmt.Sprintf("UP: %s", formatUptime(m.SystemStats.Uptime)))

	stats := fmt.Sprintf(" %s  │  %s  │  %s ", cpu, mem, uptime)
	return statsBarStyle.Render(stats)
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var output string

	// System Stats Bar
	output += m.renderStatsBar() + "\n"

	// Header
	output += headerStyle.Render("🖥️  gotop - Go Process Monitor") + "\n"
	output += fmt.Sprintf("Sort by: %s | [K]ill [c]pu [m]emory | [t]ype message | [q]uit\n\n", m.SortBy)

	// Table
	output += m.Table.View() + "\n"

	// Chatbox
	chatLabel := chatLabelStyle.Render("💬 Chat")
	if m.chatActive {
		chatLabel = chatLabelStyle.Render("💬 Chat (ESC to cancel)")
	}
	output += chatStyle.Render(chatLabel + "\n" + m.TextInput.View())

	return output
}

func (m *Model) SortProcesses() {
	sort.Slice(m.Processes, func(i, j int) bool {
		if m.SortBy == "memory" {
			return m.Processes[i].Memory > m.Processes[j].Memory
		}
		// Default: sort by CPU
		return m.Processes[i].CPU > m.Processes[j].CPU
	})
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}

func analyzeProcessBehavior(ctx context.Context, provider AIProvider, p ProcessInfo) (string, error) {
	return provider.Analyze(ctx, p)
}
