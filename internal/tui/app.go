// Package tui implements Bubble Tea models for the STUI interface.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles used across the TUI.
var (
	// BannerStyle renders the application title in bold purple.
	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	// HelpStyle renders help text in muted gray.
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// AppModel is the root model for the STUI application.
type AppModel struct {
	// quitting indicates the user has requested to quit.
	quitting bool
	// width and height of the terminal.
	width  int
	height int
}

// NewAppModel creates a new root application model.
func NewAppModel() AppModel {
	return AppModel{}
}

// Init implements tea.Model.
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View implements tea.Model.
func (m AppModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	banner := BannerStyle.Render("STUI - Sonar Terminal User Interface")
	help := HelpStyle.Render("Press q to quit.")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n", banner, help)
}

// Quitting returns whether the model is in a quitting state.
func (m AppModel) Quitting() bool {
	return m.quitting
}

// Width returns the current terminal width.
func (m AppModel) Width() int {
	return m.width
}

// Height returns the current terminal height.
func (m AppModel) Height() int {
	return m.height
}
