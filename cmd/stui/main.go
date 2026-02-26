// Package main is the CLI entrypoint for STUI.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SonarSoftwareInc/stui/internal/tui"
)

// main initializes and runs the Bubble Tea TUI program.
func main() {
	p := tea.NewProgram(tui.NewAppModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
