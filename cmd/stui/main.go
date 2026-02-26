// Package main is the CLI entrypoint for STUI.
package main

import (
	"fmt"
	"os"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/tui"
)

// main initializes and runs the Bubble Tea TUI program.
// If the user requests a privilege escalation relaunch from the
// preflight screen, the process replaces itself with the detected
// escalation command (sudo or doas).
func main() {
	p := tea.NewProgram(tui.NewAppModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check if the TUI requested a privilege escalation relaunch.
	if m, ok := finalModel.(tui.AppModel); ok && m.ElevateRelaunch() {
		relaunchElevated(m)
	}
}

// relaunchElevated replaces the current process with a privilege-
// escalated invocation using the detected method (sudo or doas).
func relaunchElevated(m tui.AppModel) {
	esc := m.Escalation()
	if esc == nil {
		fmt.Println("Error: no privilege escalation command available")
		os.Exit(1)
	}

	selfPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error: could not determine executable path: %v\n", err)
		os.Exit(1)
	}

	// Build args: ["sudo|doas", "/path/to/stui", ...original args...]
	args := append([]string{esc.Name, selfPath}, os.Args[1:]...)

	fmt.Printf("Relaunching with %s...\n", esc.Name)
	// Replace the current process with the escalation command.
	if err := syscall.Exec(esc.Path, args, os.Environ()); err != nil {
		fmt.Printf("Error: failed to exec %s: %v\n", esc.Name, err)
		os.Exit(1)
	}
}
