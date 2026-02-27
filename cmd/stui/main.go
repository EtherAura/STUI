// Package main is the CLI entrypoint for STUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
	"github.com/EtherAura/stui/internal/tui"
)

// main initializes and runs the Bubble Tea TUI program.
// If the user requests a privilege escalation relaunch from the
// preflight screen, the process replaces itself with the detected
// escalation command (sudo or doas). The --resume-app flag allows
// the relaunched process to return directly to the installer
// the user was viewing.
func main() {
	resumeApp := flag.String("resume-app", "", "app ID to resume after elevated relaunch")
	flag.Parse()

	// Validate --resume-app against the installer registry to prevent
	// untrusted strings from reaching syscall.Exec arguments.
	if *resumeApp != "" {
		reg := installer.NewRegistry()
		if _, ok := reg[*resumeApp]; !ok {
			fmt.Printf("Error: unknown app ID %q\n", *resumeApp)
			os.Exit(1)
		}
	}

	var model tui.AppModel
	if *resumeApp != "" {
		model = tui.NewAppModelWithResume(*resumeApp)
	} else {
		model = tui.NewAppModel()
	}

	p := tea.NewProgram(model)
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

	// Build args: ["sudo|doas", "/path/to/stui", "--resume-app=<id>"].
	// We construct a clean arg list rather than forwarding os.Args
	// to avoid duplicating the --resume-app flag.
	args := []string{esc.Name, selfPath}
	if appID := m.ResumeAppID(); appID != "" {
		args = append(args, "--resume-app="+appID)
	}

	// Clear the terminal so the relaunched process starts clean.
	fmt.Print("\033[H\033[2J")
	fmt.Printf("Relaunching with %s...\n", esc.Name)
	// Replace the current process with the escalation command.
	// Use a sanitized environment to avoid leaking sensitive vars.
	if err := syscall.Exec(esc.Path, args, sanitizeEnv(os.Environ())); err != nil {
		fmt.Printf("Error: failed to exec %s: %v\n", esc.Name, err)
		os.Exit(1)
	}
}

// envAllowlist lists environment variable prefixes and exact names
// that are safe to forward to the elevated child process.
var envAllowlist = []string{
	"PATH=",
	"TERM=",
	"HOME=",
	"LANG=",
	"LC_",
	"DISPLAY=",
	"XDG_RUNTIME_DIR=",
	"USER=",
	"LOGNAME=",
	"SHELL=",
	"COLORTERM=",
}

// sanitizeEnv filters an environment variable list to only include
// entries matching the allowlist. This prevents sensitive variables
// (API keys, cloud credentials, etc.) from leaking into a root process.
func sanitizeEnv(environ []string) []string {
	var filtered []string
	for _, entry := range environ {
		for _, prefix := range envAllowlist {
			if strings.HasPrefix(entry, prefix) {
				filtered = append(filtered, entry)
				break
			}
		}
	}
	return filtered
}
