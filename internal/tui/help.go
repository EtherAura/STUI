// help.go implements the help screen.
// This screen displays documentation links, version information,
// and keyboard shortcut reference for the STUI application.
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// HelpModel displays help information and documentation links.
type HelpModel struct {
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewHelpModel creates a new help screen model.
func NewHelpModel() HelpModel {
	return HelpModel{}
}

// Init implements tea.Model. No initial command needed.
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles navigation keys.
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
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

// View implements tea.Model. Renders the help screen.
func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Help"))
	b.WriteString("\n\n")

	b.WriteString(BannerStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render("  enter    Select / confirm"))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render("  esc      Go back"))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render("  q        Quit"))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render("  ctrl+c   Force quit"))
	b.WriteString("\n\n")

	b.WriteString(BannerStyle.Render("About"))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render("  STUI is a community-built terminal interface for"))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render("  installing and managing Sonar applications."))
	b.WriteString("\n\n")
	b.WriteString(DimStyle.Render("  https://github.com/EtherAura/stui"))
	b.WriteString("\n\n")

	b.WriteString(HelpStyle.Render("Press esc to go back"))

	return AppStyle.Render(b.String())
}
