package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Brownei/aitop/process"
	"github.com/Brownei/aitop/providers"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	provider, err := providers.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	procs, err := process.GetProcesses()
	if err != nil {
		log.Fatal(err)
	}

	model := process.NewModel(procs, provider)
	model.SortProcesses()
	model.Update(model.Table)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
