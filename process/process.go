package process

import (
	"github.com/shirou/gopsutil/v4/process"
)

type ProcessInfo struct {
	PID        int32
	Name       string
	CPU        float64
	Memory     float32
	CPUHistory []float64 // Last 10 readings
	Status     string    // "normal", "suspicious", "stuck"
	AIAnalysis string    // AI's assessment
	Nice       int
}

func GetProcesses() ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procs []ProcessInfo

	for _, p := range processes {
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()

		procs = append(procs, ProcessInfo{
			PID:    p.Pid,
			Name:   name,
			CPU:    cpuPercent,
			Memory: memPercent,
		})
	}

	return procs, nil
}

func killProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

// Add to tracking
func (p *ProcessInfo) recordCPU(cpuPercent float64) {
	p.CPUHistory = append(p.CPUHistory, cpuPercent)
	// Keep only last 10 readings
	if len(p.CPUHistory) > 10 {
		p.CPUHistory = p.CPUHistory[1:]
	}
}

// Check if CPU is abnormally high
func (p ProcessInfo) isHighCPU() bool {
	return p.CPU > 9.0 // More than 20%
}
