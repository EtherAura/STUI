// Package tui implements Bubble Tea models for the STUI interface.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// Screen identifies which TUI screen is currently active.
type Screen int

// Screen constants enumerate all screens in the application flow.
const (
	// ScreenMenu is the main application selection menu.
	ScreenMenu Screen = iota
	// ScreenDetail shows app details and a confirm/back prompt.
	ScreenDetail
	// ScreenPreflight runs and displays preflight checks.
	ScreenPreflight
	// ScreenConfig is the interactive configuration wizard.
	ScreenConfig
	// ScreenInstall shows real-time installation progress.
	ScreenInstall
	// ScreenVerify displays post-install verification results.
	ScreenVerify
)

// AppModel is the root model for the STUI application.
// It owns the installer registry, tracks the active screen,
// and delegates to the appropriate sub-model.
type AppModel struct {
	// registry holds the available installer constructors.
	registry installer.Registry
	// screen is the currently active screen.
	screen Screen
	// menu is the main application selection menu model.
	menu MenuModel
	// detail is the app detail/confirm screen model.
	detail DetailModel
	// preflight is the preflight check screen model.
	preflight PreflightModel
	// quitting indicates the user has requested to quit.
	quitting bool
	// width and height of the terminal.
	width  int
	height int
}

// NewAppModel creates a new root application model with the
// default installer registry and menu pre-loaded.
func NewAppModel() AppModel {
	reg := installer.NewRegistry()
	return AppModel{
		registry: reg,
		screen:   ScreenMenu,
		menu:     NewMenuModel(reg),
	}
}

// Init implements tea.Model. Initializes the active sub-model.
func (m AppModel) Init() tea.Cmd {
	return m.menu.Init()
}

// Update implements tea.Model. Dispatches messages to the active
// screen's sub-model and handles global keys and screen transitions.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case AppSelectedMsg:
		// Transition from menu to detail screen.
		m.detail = NewDetailModel(m.registry, msg.AppID)
		m.screen = ScreenDetail
		return m, nil
	case BackToMenuMsg:
		// Return to the main menu.
		m.screen = ScreenMenu
		return m, nil
	case StartPreflightMsg:
		// Transition from detail to preflight screen.
		m.preflight = NewPreflightModel(m.registry, msg.AppID)
		m.screen = ScreenPreflight
		return m, m.preflight.Init()
	case StartConfigMsg:
		// TODO: transition to config wizard screen.
		return m, nil
	}

	// Delegate to the active screen's sub-model.
	var cmd tea.Cmd
	switch m.screen {
	case ScreenMenu:
		m.menu, cmd = m.menu.Update(msg)
	case ScreenDetail:
		m.detail, cmd = m.detail.Update(msg)
	case ScreenPreflight:
		m.preflight, cmd = m.preflight.Update(msg)
	}

	return m, cmd
}

// View implements tea.Model. Renders the active screen.
func (m AppModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.screen {
	case ScreenMenu:
		return m.menu.View()
	case ScreenDetail:
		return m.detail.View()
	case ScreenPreflight:
		return m.preflight.View()
	default:
		return ""
	}
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

// Screen returns the currently active screen.
func (m AppModel) Screen() Screen {
	return m.screen
}
