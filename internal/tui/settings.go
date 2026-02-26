// settings.go implements the settings screen placeholder.
// This screen will eventually allow the user to configure STUI
// preferences such as default Sonar URL, theme, and log verbosity.
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SettingsModel displays the settings screen.
type SettingsModel struct {
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewSettingsModel creates a new settings screen model.
func NewSettingsModel() SettingsModel {
	return SettingsModel{}
}

// Init implements tea.Model. No initial command needed.
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles navigation keys.
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View implements tea.Model. Renders the settings screen.
func (m SettingsModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Settings"))
	b.WriteString("\n\n")
	b.WriteString(DimStyle.Render("No settings available yet."))
	b.WriteString("\n")
	b.WriteString(DimStyle.Render("Future options: default Sonar URL, theme, log verbosity."))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("Press esc to go back"))

	return AppStyle.Render(b.String())
}
