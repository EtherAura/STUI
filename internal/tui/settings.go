// settings.go implements the settings screen for STUI preferences.
// Currently supports toggling password field visibility in the
// config wizard. Future options: default Sonar URL, theme, log
// verbosity.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// settingItem describes a single boolean toggle in the settings screen.
type settingItem struct {
	// label is the user-visible name for this setting.
	label string
	// description provides additional context below the label.
	description string
}

// settingsItems is the ordered list of toggle settings.
// The index here corresponds to the field order in SettingsModel.
var settingsItems = []settingItem{
	{
		label:       "Show Passwords",
		description: "Display password and secret fields as plaintext in the config wizard",
	},
}

// SettingsModel displays the settings screen and tracks user preferences.
type SettingsModel struct {
	// showPasswords controls whether secret fields are displayed
	// as plaintext (true) or masked (false) in the config wizard.
	showPasswords bool
	// cursor tracks which setting is currently highlighted.
	cursor int
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewSettingsModel creates a new settings screen model with defaults.
// Passwords are hidden by default.
func NewSettingsModel() SettingsModel {
	return SettingsModel{}
}

// ShowPasswords returns whether secret fields should be displayed
// as plaintext in the config wizard.
func (m SettingsModel) ShowPasswords() bool {
	return m.showPasswords
}

// Init implements tea.Model. No initial command needed.
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles navigation and toggle keys.
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(settingsItems)-1 {
				m.cursor++
			}
		case " ", "enter":
			m = m.toggleCurrent()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// toggleCurrent flips the boolean value of the currently selected setting.
func (m SettingsModel) toggleCurrent() SettingsModel {
	if m.cursor == 0 {
		m.showPasswords = !m.showPasswords
	}
	return m
}

// boolAt returns the current boolean value for the setting at index i.
func (m SettingsModel) boolAt(i int) bool {
	if i == 0 {
		return m.showPasswords
	}
	return false
}

// View implements tea.Model. Renders the settings with toggle checkboxes.
func (m SettingsModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	for i, item := range settingsItems {
		// Render checkbox indicator.
		checkbox := "[ ]"
		if m.boolAt(i) {
			checkbox = "[x]"
		}

		// Highlight the cursor line.
		line := fmt.Sprintf("%s %s", checkbox, item.label)
		if i == m.cursor {
			b.WriteString(BannerStyle.Render(line))
		} else {
			b.WriteString(BodyStyle.Render(line))
		}
		b.WriteString("\n")
		b.WriteString("    " + DimStyle.Render(item.description))
		b.WriteString("\n\n")
	}

	b.WriteString(HelpStyle.Render("space/enter: toggle • ↑/↓: navigate • esc: back"))

	return AppStyle.Render(b.String())
}

// Cursor returns the index of the currently highlighted setting.
func (m SettingsModel) Cursor() int {
	return m.cursor
}
