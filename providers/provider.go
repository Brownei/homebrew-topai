package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Brownei/aitop/process"
	tea "github.com/charmbracelet/bubbletea"
)

func LoadConfig() (process.AIProvider, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "aitop", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0700)

	// Try to load existing config
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg process.ProviderConfig
		if err := json.Unmarshal(data, &cfg); err == nil && cfg.Provider != "" && cfg.APIKey != "" {
			return initializeProvider(cfg.Provider, cfg.APIKey)
		}
	}

	// First run - prompt user to set up
	return promptForProvider(configPath)
}

func initializeProvider(providerName, apiKey string) (process.AIProvider, error) {
	ctx := context.Background()
	var provider process.AIProvider

	switch providerName {
	case "anthropic":
		provider = NewAnthropicProvider(apiKey)
	case "openai":
		provider = NewOpenAIProvider(apiKey)
	case "gemini":
		provider = NewGeminiProvider(apiKey)
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	// Validate the API key
	if err := provider.ValidateAPIKey(ctx); err != nil {
		return nil, fmt.Errorf("invalid API key for %s: %v", providerName, err)
	}

	return provider, nil
}

func promptForProvider(configPath string) (process.AIProvider, error) {
	// Use the new TUI-based selection
	model := NewSelectionModel()

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run provider selection: %w", err)
	}

	m := finalModel.(SelectionModel)

	// Check if user quit without selecting
	if !m.IsValidated() || m.GetResult() == nil {
		return nil, fmt.Errorf("no provider selected")
	}

	// Get the selected provider details
	providerID, apiKey := m.GetSelectedProvider()

	// Save config
	cfg := process.ProviderConfig{
		Provider: providerID,
		APIKey:   apiKey,
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(configPath, data, 0600)

	fmt.Printf("\n✓ Provider set to: %s\n", m.GetResult().Name())
	fmt.Printf("✓ Config saved to: %s\n\n", configPath)

	return m.GetResult(), nil
}
