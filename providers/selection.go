package providers

import (
	"fmt"
	"strings"

	"github.com/Brownei/aitop/process"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the provider selection UI
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(true).
				Padding(0, 1)

	unselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginLeft(4)

	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Underline(true).
			MarginLeft(4)

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			Width(70)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginLeft(2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true).
			MarginLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1).
			MarginLeft(2)
)

// ProviderInfo holds display information for each provider
type ProviderInfo struct {
	ID          string
	Name        string
	Icon        string
	Description string
	Model       string
	URL         string
	Color       lipgloss.Color
}

var providersList = []ProviderInfo{
	{
		ID:          "anthropic",
		Name:        "Anthropic Claude",
		Icon:        "◬",
		Description: "Fast, efficient analysis with Claude Haiku",
		Model:       "Claude Haiku 4.5",
		URL:         "https://console.anthropic.com",
		Color:       lipgloss.Color("208"), // Orange/brown
	},
	{
		ID:          "openai",
		Name:        "OpenAI GPT-4",
		Icon:        "◯",
		Description: "Powerful general-purpose analysis",
		Model:       "GPT-4",
		URL:         "https://platform.openai.com/api-keys",
		Color:       lipgloss.Color("10"), // Green
	},
	{
		ID:          "gemini",
		Name:        "Google Gemini",
		Icon:        "♊",
		Description: "Google's latest AI with fast responses",
		Model:       "Gemini 1.5 Flash",
		URL:         "https://aistudio.google.com/app/apikey",
		Color:       lipgloss.Color("39"), // Blue
	},
}

// SelectionState tracks which part of the UI we're in
type SelectionState int

const (
	StateSelecting SelectionState = iota
	StateEnteringKey
	StateValidating
	StateError
	StateSuccess
)

// SelectionModel is the Bubble Tea model for provider selection
type SelectionModel struct {
	state       SelectionState
	cursor      int
	providers   []ProviderInfo
	apiKeyInput textinput.Model
	selected    *ProviderInfo
	err         error
	validated   bool
	result      process.AIProvider
	quitting    bool
}

// NewSelectionModel creates a new provider selection model
func NewSelectionModel() SelectionModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your API key..."
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 256
	ti.Width = 50

	return SelectionModel{
		state:       StateSelecting,
		cursor:      0,
		providers:   providersList,
		apiKeyInput: ti,
	}
}

func (m SelectionModel) Init() tea.Cmd {
	return nil
}

func (m SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case StateSelecting:
			return m.handleSelectingKeys(msg)
		case StateEnteringKey:
			return m.handleEnteringKeyKeys(msg)
		case StateError:
			// Any key dismisses error and goes back to entering key
			m.state = StateEnteringKey
			m.err = nil
			return m, nil
		case StateSuccess:
			// Any key exits on success
			if msg.String() == "enter" || msg.String() == "q" || msg.String() == "esc" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

	case validationResult:
		return handleValidationResult(m, msg)
	}

	// Update text input
	if m.state == StateEnteringKey {
		var cmd tea.Cmd
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m SelectionModel) handleSelectingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.providers)-1 {
			m.cursor++
		}
		return m, nil

	case "enter", " ":
		m.selected = &m.providers[m.cursor]
		m.state = StateEnteringKey
		m.apiKeyInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m SelectionModel) handleEnteringKeyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Go back to provider selection
		m.state = StateSelecting
		m.selected = nil
		m.apiKeyInput.Blur()
		m.apiKeyInput.SetValue("")
		return m, nil

	case "enter":
		apiKey := m.apiKeyInput.Value()
		if apiKey == "" {
			return m, nil
		}

		// Start validation
		m.state = StateValidating
		return m, m.validateAPIKey(apiKey)

	default:
		var cmd tea.Cmd
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
		return m, cmd
	}
}

type validationResult struct {
	provider process.AIProvider
	err      error
}

func (m SelectionModel) validateAPIKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		provider, err := initializeProvider(m.selected.ID, apiKey)
		return validationResult{provider: provider, err: err}
	}
}

// Handle the validation result
func handleValidationResult(m SelectionModel, result validationResult) (tea.Model, tea.Cmd) {
	if result.err != nil {
		m.state = StateError
		m.err = result.err
		return m, nil
	}

	m.state = StateSuccess
	m.validated = true
	m.result = result.provider
	return m, nil
}

// View renders the UI
func (m SelectionModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🧠 aitop - AI Provider Selection") + "\n")
	b.WriteString(subtitleStyle.Render("Select an AI provider to analyze your processes") + "\n\n")

	switch m.state {
	case StateSelecting:
		b.WriteString(m.renderProviderList())
	case StateEnteringKey:
		b.WriteString(m.renderAPIKeyInput())
	case StateValidating:
		b.WriteString(m.renderValidating())
	case StateError:
		b.WriteString(m.renderError())
	case StateSuccess:
		b.WriteString(m.renderSuccess())
	}

	return boxStyle.Render(b.String())
}

func (m SelectionModel) renderProviderList() string {
	var b strings.Builder

	for i, p := range m.providers {
		isSelected := i == m.cursor

		// Icon and name
		iconStyle := lipgloss.NewStyle().Foreground(p.Color)
		icon := iconStyle.Render(p.Icon)

		var name string
		if isSelected {
			name = selectedItemStyle.Render("▸ " + icon + "  " + p.Name)
		} else {
			name = unselectedItemStyle.Render("  " + icon + "  " + p.Name)
		}

		b.WriteString(name + "\n")

		// Description and URL (only for selected)
		if isSelected {
			b.WriteString(descriptionStyle.Render(p.Description+" → "+p.Model) + "\n")
			b.WriteString(urlStyle.Render("Get key: "+p.URL) + "\n")
			b.WriteString("\n")
		}
	}

	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • q quit"))
	return b.String()
}

func (m SelectionModel) renderAPIKeyInput() string {
	var b strings.Builder

	// Selected provider info
	p := m.selected
	iconStyle := lipgloss.NewStyle().Foreground(p.Color)
	b.WriteString(fmt.Sprintf("Selected: %s %s\n\n", iconStyle.Render(p.Icon), p.Name))

	// API key input
	b.WriteString(labelStyle.Render("Enter your API key:") + "\n")
	b.WriteString(m.apiKeyInput.View() + "\n\n")

	b.WriteString(helpStyle.Render("esc go back • enter submit • input is hidden"))
	return b.String()
}

func (m SelectionModel) renderValidating() string {
	return labelStyle.Render("✓ Validating API key...") + "\n\n" +
		helpStyle.Render("Please wait while we verify your credentials")
}

func (m SelectionModel) renderError() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render("✗ Validation failed") + "\n\n")
	b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	b.WriteString(helpStyle.Render("Press any key to try again"))

	return b.String()
}

func (m SelectionModel) renderSuccess() string {
	var b strings.Builder

	b.WriteString(successStyle.Render("✓ API key validated successfully!") + "\n\n")
	b.WriteString(labelStyle.Render(fmt.Sprintf("Provider: %s", m.result.Name())) + "\n\n")
	b.WriteString(helpStyle.Render("Press enter to continue"))

	return b.String()
}

// GetResult returns the selected provider (call after program exits with success)
func (m SelectionModel) GetResult() process.AIProvider {
	return m.result
}

// GetSelectedProvider returns the selected provider ID and API key
func (m SelectionModel) GetSelectedProvider() (string, string) {
	if m.selected == nil {
		return "", ""
	}
	return m.selected.ID, m.apiKeyInput.Value()
}

// IsValidated returns whether the API key was validated
func (m SelectionModel) IsValidated() bool {
	return m.validated
}
