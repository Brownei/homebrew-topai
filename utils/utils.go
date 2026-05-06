package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetGlobalPrompt(name string, cpu, memory float64, pid int32) string {
	return fmt.Sprintf(`Analyze this process behavior:
- Name: %s
- PID: %d
- Current CPU: %.2f%%
- Memory: %.2f%%
- CPU History (last 10 seconds): %v

Is this process "stuck" (hung, frozen, deadlocked) or just "busy" (legitimately using resources)?

Respond in ONE sentence:
- If STUCK: "STUCK: [brief reason why it seems frozen]"
- If BUSY: "BUSY: [brief reason it's legitimately busy]"
- If NORMAL: "NORMAL: [why it's fine]"`,
		name, pid, cpu, memory, cpu)
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "Unable to get config directory", err
	}

	configPath := filepath.Join(configDir, "topia", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0700)

	return configPath, nil
}
