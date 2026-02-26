// detail.go implements the application detail/confirm screen.
// After selecting an app from the main menu, this screen shows
// the app name, description, and ordered install steps, then
// asks the user to confirm or go back.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// StartPreflightMsg signals the app to transition from the detail
// screen to the preflight check screen.
type StartPreflightMsg struct {
	// AppID is the registry key of the selected application.
	AppID string
}

// BackToMenuMsg signals the app to return to the main menu.
type BackToMenuMsg struct{}

// DetailModel displays information about a selected application
// and prompts the user to confirm or go back.
type DetailModel struct {
	// appID is the registry key of the selected application.
	appID string
	// installer is the instantiated installer for the selected app.
	installer installer.Installer
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewDetailModel creates a detail screen for the given app ID,
// instantiating its installer from the registry.
func NewDetailModel(reg installer.Registry, appID string) DetailModel {
	ctor, ok := reg[appID]
	if !ok {
		return DetailModel{appID: appID}
	}
	return DetailModel{
		appID:     appID,
		installer: ctor(),
	}
}

// Init implements tea.Model. No initial command needed.
func (m DetailModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles enter to confirm and
// esc/backspace to return to the menu.
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, func() tea.Msg {
				return StartPreflightMsg{AppID: m.appID}
			}
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

// View implements tea.Model. Renders the app details and instructions.
func (m DetailModel) View() string {
	if m.installer == nil {
		return AppStyle.Render(
			ErrorStyle.Render("Error: unknown application ") + m.appID + "\n\n" +
				HelpStyle.Render("Press esc to go back."),
		)
	}

	var b strings.Builder

	// Title section.
	b.WriteString(TitleStyle.Render(m.installer.Name()))
	b.WriteString("\n")
	b.WriteString(BodyStyle.Render(m.installer.Description()))
	b.WriteString("\n\n")

	// Installation steps section.
	steps := m.installer.Steps()
	b.WriteString(BannerStyle.Render("Installation Steps"))
	b.WriteString("\n")
	for i, step := range steps {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  %d. %s", i+1, step.Name)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Action prompt.
	b.WriteString(SuccessStyle.Render("Press enter to begin installation"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press esc to go back"))

	return AppStyle.Render(b.String())
}

// AppID returns the registry key of the displayed application.
func (m DetailModel) AppID() string {
	return m.appID
}
