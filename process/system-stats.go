package process

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemStats struct {
	CPUPercent float64
	MemPercent float64
	Uptime     uint64
}

func getSystemStats() (SystemStats, error) {
	cpuPercent, _ := cpu.Percent(0, false)
	memStats, _ := mem.VirtualMemory()
	uptime, _ := host.Uptime()

	return SystemStats{
		CPUPercent: cpuPercent[0],
		MemPercent: memStats.UsedPercent,
		Uptime:     uptime,
	}, nil
}
