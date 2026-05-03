package process

import (
	"fmt"
	"sort"
	"time"

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

	chatStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	chatLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)
)

type Model struct {
	Processes  []ProcessInfo
	Table      table.Model
	TextInput  textinput.Model
	SortBy     string
	quitting   bool
	width      int
	height     int
	chatActive bool
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
func NewModel(procs []ProcessInfo) Model {
	// Setup table columns
	columns := []table.Column{
		{Title: "PID", Width: 10},
		{Title: "Name", Width: 25},
		{Title: "CPU%", Width: 12},
		{Title: "Memory%", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(20),
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

	return Model{
		Processes:  procs,
		Table:      t,
		TextInput:  ti,
		SortBy:     "cpu",
		chatActive: false,
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.Table.SetHeight(msg.Height - 8)
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

		case "up", "k":
			m.Table, _ = m.Table.Update(msg)

		case "down", "j":
			m.Table, _ = m.Table.Update(msg)

		case "c":
			m.SortBy = "cpu"
			m.SortProcesses()
			m.updateTableRows()

		case "m":
			m.SortBy = "memory"
			m.SortProcesses()
			m.updateTableRows()

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

		case "t", "enter":
			// Activate chat mode
			m.chatActive = true
			m.TextInput.Focus()
			cmds = append(cmds, textinput.Blink)
		}

	case TickMsg:
		// Update process list every tick
		procs, err := GetProcesses()
		if err == nil {
			m.Processes = procs
			m.SortProcesses()
			m.updateTableRows()
		}
		cmds = append(cmds, tickCmd())
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
			utils.Truncate(p.Name, 25),
			fmt.Sprintf("%.2f", p.CPU),
			fmt.Sprintf("%.2f", p.Memory),
		})
	}
	m.Table.SetRows(rows)
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var output string

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
