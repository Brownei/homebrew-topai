package process

import (
	"github.com/shirou/gopsutil/v4/process"
)

type ProcessInfo struct {
	PID    int32
	Name   string
	CPU    float64
	Memory float32
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
