package main

import (
	"fmt"
	"os/exec"
)

// Throttle a process by lowering its priority
func throttleProcess(pid int32, niceValue int) error {
	cmd := exec.Command("renice", "-n", fmt.Sprintf("%d", niceValue), "-p", fmt.Sprintf("%d", pid))
	return cmd.Run()
}

// Restore normal priority
func UnthrottleProcess(pid int32) error {
	return throttleProcess(pid, 0)
}
