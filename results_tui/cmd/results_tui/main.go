package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"results_tui/internal/results"
	"results_tui/internal/tui"
)

func main() {
	root, err := results.FindRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve results root: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(root, results.Scan(root))
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run results tui: %v\n", err)
		os.Exit(1)
	}
}
