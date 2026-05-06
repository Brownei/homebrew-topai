package process

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Brownei/aitop/providers"
	"github.com/Brownei/aitop/utils"
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

	// Modal styles
	modalStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("255")).
			Padding(2, 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Width(80)

	modalTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true).
			MarginBottom(1).
			Align(lipgloss.Center)

	userMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true).
				MarginBottom(1)

	responseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			MarginLeft(2).
			MarginBottom(1)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginBottom(1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginTop(1).
			Align(lipgloss.Center)

	// Overlay background (dimmed)
	overlayStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("0"))
)

type AIAnalysisMsg struct {
	PID      int32
	Analysis string
}

type ThrottlePromptMsg struct {
	PID int32
}

// Chat response message for async AI calls
type ChatResponseMsg struct {
	Response string
	Error    error
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
	provider           providers.AIProvider
	throttlePrompt     *ThrottlePromptMsg // nil or set to show prompt

	// Chat state
	chatActive   bool
	chatLoading  bool
	chatResponse string
	chatError    string
	lastSentMsg  string
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
func NewModel(procs []ProcessInfo, provider providers.AIProvider) Model {
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
	ti.Placeholder = "Ask me anything about your system..."
	ti.CharLimit = 500
	ti.Width = 70

	// Get initial system stats
	stats, _ := getSystemStats()

	return Model{
		Processes:   procs,
		SystemStats: stats,
		Table:       t,
		TextInput:   ti,
		SortBy:      "cpu",
		chatActive:  false,
		chatLoading: false,
		provider:    provider,
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()
	configDir, err := os.UserConfigDir()
	var configPath string
	if err == nil {
		configPath = filepath.Join(configDir, "topia", "config.json")
		os.MkdirAll(filepath.Dir(configPath), 0700)
	}

	defer ctx.Done()

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.Table.SetHeight(msg.Height - 10)
		// Update modal width based on screen size
		modalWidth := min(80, msg.Width-4)
		m.TextInput.Width = modalWidth - 8
		modalStyle = modalStyle.Width(modalWidth)

	case tea.KeyMsg:
		// Handle chat mode
		if m.chatActive {
			switch msg.String() {
			case "esc":
				m.chatActive = false
				m.TextInput.Blur()
				m.chatLoading = false
				return m, nil
			case "enter":
				// Send the message
				msgText := m.TextInput.Value()
				if msgText != "" && !m.chatLoading {
					m.lastSentMsg = msgText
					m.chatLoading = true
					m.chatResponse = ""
					m.chatError = ""
					m.TextInput.SetValue("")
					m.TextInput.Blur()

					// Send async request
					return m, func() tea.Msg {
						ctx := context.Background()
						response, err := m.provider.Analyze(ctx, msgText)
						return ChatResponseMsg{Response: response, Error: err}
					}
				}
				return m, nil
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			default:
				if !m.chatLoading {
					var cmd tea.Cmd
					m.TextInput, cmd = m.TextInput.Update(msg)
					return m, cmd
				}
				return m, nil
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
			// Toggle chat modal
			m.chatActive = !m.chatActive
			if m.chatActive {
				m.TextInput.Focus()
				cmds = append(cmds, textinput.Blink)
			} else {
				m.TextInput.Blur()
			}
			return m, nil

		case "p":
			// Switch provider
			m.throttlePrompt = nil
			newProvider, err := providers.PromptForProvider(configPath)
			if err == nil {
				m.provider = newProvider
				// Clear old analyses
				for i := range m.Processes {
					m.Processes[i].AIAnalysis = ""
				}
			}
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

	case ChatResponseMsg:
		m.chatLoading = false
		if msg.Error != nil {
			m.chatError = msg.Error.Error()
			m.chatResponse = ""
		} else {
			m.chatResponse = msg.Response
			m.chatError = ""
		}
		// Re-focus input for next message
		m.TextInput.Focus()
		cmds = append(cmds, textinput.Blink)
		return m, nil
	}

	// Update table (only when not in chat mode)
	if !m.chatActive {
		var tableCmd tea.Cmd
		m.Table, tableCmd = m.Table.Update(msg)
		cmds = append(cmds, tableCmd)
	}

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

// wrapText wraps text to fit within a given width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Word wrap
		words := strings.Fields(line)
		currentLine := ""
		for _, word := range words {
			if len(currentLine)+len(word)+1 > width {
				result.WriteString(strings.TrimSpace(currentLine))
				result.WriteString("\n")
				currentLine = word + " "
			} else {
				currentLine += word + " "
			}
		}
		if currentLine != "" {
			result.WriteString(strings.TrimSpace(currentLine))
			result.WriteString("\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// renderChatModal creates the chat modal overlay
func (m Model) renderChatModal() string {
	var content strings.Builder

	// Title
	title := modalTitleStyle.Render("💬 Ask AI")
	content.WriteString(title + "\n\n")

	// Show user message if sent
	if m.lastSentMsg != "" {
		userLabel := userMessageStyle.Render("You:")
		content.WriteString(userLabel + "\n")

		// Wrap user message
		wrappedUser := wrapText(m.lastSentMsg, m.TextInput.Width)
		content.WriteString(responseStyle.Render(wrappedUser) + "\n\n")
	}

	// Show loading, error, or response
	if m.chatLoading {
		loading := loadingStyle.Render("⏳ AI is thinking...")
		content.WriteString(loading + "\n\n")
	} else if m.chatError != "" {
		errMsg := errorStyle.Render("❌ Error: " + m.chatError)
		content.WriteString(errMsg + "\n\n")
	} else if m.chatResponse != "" {
		aiLabel := userMessageStyle.Render("AI:")
		content.WriteString(aiLabel + "\n")

		// Wrap response to fit modal
		wrappedResponse := wrapText(m.chatResponse, m.TextInput.Width)
		content.WriteString(responseStyle.Render(wrappedResponse) + "\n\n")
	}

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.TextInput.Width))
	content.WriteString(separator + "\n\n")

	// Input field (only show when not loading)
	if !m.chatLoading {
		content.WriteString(m.TextInput.View() + "\n")
	}

	// Hint
	hint := hintStyle.Render("Enter to send • ESC to close • Ctrl+C to quit")
	content.WriteString(hint)

	// Apply modal styling
	modalContent := modalStyle.Render(content.String())

	// Center the modal on screen
	if m.width > 0 && m.height > 0 {
		modalWidth := lipgloss.Width(modalContent)
		modalHeight := lipgloss.Height(modalContent)

		// Horizontal centering
		leftPadding := (m.width - modalWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}

		// Vertical centering
		topPadding := (m.height - modalHeight) / 2
		if topPadding < 0 {
			topPadding = 0
		}

		return lipgloss.NewStyle().
			Padding(topPadding, leftPadding).
			Render(modalContent)
	}

	return modalContent
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
	output += headerStyle.Render("🖥️  Topai - Go Process Monitor") + "\n"
	output += fmt.Sprintf("Sort by: %s | [K]ill [c]pu [m]emory | [t]oggle chat | [q]uit\n\n", m.SortBy)

	// Table
	output += m.Table.View() + "\n"

	// Render chat modal as overlay if active
	if m.chatActive {
		modalView := m.renderChatModal()

		// Use lipgloss to overlay modal on main view
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

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

func analyzeProcessBehavior(ctx context.Context, provider providers.AIProvider, p ProcessInfo) (string, error) {
	prompt := utils.GetGlobalPrompt(p.Name, p.CPU, float64(p.Memory), p.PID)

	return provider.Analyze(ctx, prompt)
}
