// Package main is the CLI entrypoint for STUI.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/tui"
)

// main initializes and runs the Bubble Tea TUI program.
// If the user requests a sudo relaunch from the preflight screen,
// the process replaces itself with a sudo-wrapped invocation.
func main() {
	p := tea.NewProgram(tui.NewAppModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check if the TUI requested a sudo relaunch.
	if m, ok := finalModel.(tui.AppModel); ok && m.SudoRelaunch() {
		relaunchWithSudo()
	}
}

// relaunchWithSudo replaces the current process with a sudo-wrapped
// invocation of the same binary and arguments.
func relaunchWithSudo() {
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		fmt.Println("Error: sudo not found on PATH")
		os.Exit(1)
	}

	selfPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error: could not determine executable path: %v\n", err)
		os.Exit(1)
	}

	// Build args: ["sudo", "/path/to/stui", ...original args...]
	args := append([]string{"sudo", selfPath}, os.Args[1:]...)

	fmt.Println("Relaunching with sudo...")
	// Replace the current process with sudo.
	if err := syscall.Exec(sudoPath, args, os.Environ()); err != nil {
		fmt.Printf("Error: failed to exec sudo: %v\n", err)
		os.Exit(1)
	}
}
