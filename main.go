package main

import (
	"log"

	"github.com/Brownei/aitop/process"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	procs, err := process.GetProcesses()
	if err != nil {
		log.Fatal(err)
	}

	model := process.NewModel(procs)
	model.SortProcesses()
	model.Update(model.Table)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
