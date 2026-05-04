package utils

import (
	"fmt"

	"github.com/Brownei/aitop/process"
)

func GetGlobalPrompt(p process.ProcessInfo) string {
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
		p.Name, p.PID, p.CPU, p.Memory, p.CPU)
}
