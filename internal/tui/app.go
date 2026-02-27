// Package tui implements Bubble Tea models for the STUI interface.
package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// Screen identifies which TUI screen is currently active.
type Screen int

// Screen constants enumerate all screens in the application flow.
const (
	// ScreenMenu is the top-level category menu.
	ScreenMenu Screen = iota
	// ScreenInstallers is the app selection sub-menu under Installers.
	ScreenInstallers
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
	// ScreenSettings is the settings configuration screen.
	ScreenSettings
	// ScreenHelp shows help information and keyboard shortcuts.
	ScreenHelp
)

// AppModel is the root model for the STUI application.
// It owns the installer registry, tracks the active screen,
// and delegates to the appropriate sub-model.
type AppModel struct {
	// registry holds the available installer constructors.
	registry installer.Registry
	// screen is the currently active screen.
	screen Screen
	// menu is the top-level category menu model.
	menu MenuModel
	// installerList is the installer selection sub-menu model.
	installerList InstallerListModel
	// detail is the app detail/confirm screen model.
	detail DetailModel
	// preflight is the preflight check screen model.
	preflight PreflightModel
	// config is the configuration wizard screen model.
	config ConfigModel
	// progress is the installation progress screen model.
	progress ProgressModel
	// verify is the post-install verification screen model.
	verify VerifyModel
	// settings is the settings screen model.
	settings SettingsModel
	// help is the help screen model.
	help HelpModel
	// cancel cancels the context for the currently running async
	// operation (preflight, install, verify). Nil when idle.
	cancel context.CancelFunc
	// quitting indicates the user has requested to quit.
	quitting bool
	// elevateRelaunch indicates the app should relaunch with elevated privileges.
	elevateRelaunch bool
	// escalation holds the detected privilege escalation method for relaunch.
	escalation *installer.EscalationMethod
	// resumeAppID is the app to resume at after an elevated relaunch.
	// Empty when starting normally.
	resumeAppID string
	// width and height of the terminal.
	width  int
	height int
}

// NewAppModel creates a new root application model with the
// default installer registry and top-level menu pre-loaded.
func NewAppModel() AppModel {
	reg := installer.NewRegistry()
	return AppModel{
		registry: reg,
		screen:   ScreenMenu,
		menu:     NewMenuModel(),
	}
}

// NewAppModelWithResume creates a root model that immediately
// navigates to the preflight screen for the given app ID.
// Used after a privilege-escalated relaunch so the user is returned
// to the installer they were viewing.
func NewAppModelWithResume(appID string) AppModel {
	reg := installer.NewRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	return AppModel{
		registry:    reg,
		screen:      ScreenPreflight,
		menu:        NewMenuModel(),
		preflight:   NewPreflightModel(ctx, reg, appID),
		cancel:      cancel,
		resumeAppID: appID,
	}
}

// Init implements tea.Model. Initializes the active sub-model.
// When resuming after an elevated relaunch, the preflight check
// is started immediately instead of the menu.
func (m AppModel) Init() tea.Cmd {
	if m.resumeAppID != "" {
		return m.preflight.Init()
	}
	return m.menu.Init()
}

// Update implements tea.Model. Dispatches messages to the active
// screen's sub-model and handles global keys and screen transitions.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelRunning()
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case CategorySelectedMsg:
		// Transition from top-level menu to a category screen.
		switch msg.Category {
		case CategoryInstallers:
			m.installerList = NewInstallerListModel(m.registry)
			m.screen = ScreenInstallers
			return m, nil
		case CategorySettings:
			m.settings = NewSettingsModel()
			m.screen = ScreenSettings
			return m, nil
		case CategoryHelp:
			m.help = NewHelpModel()
			m.screen = ScreenHelp
			return m, nil
		}
	case AppSelectedMsg:
		// Transition from installer list to detail screen.
		m.detail = NewDetailModel(m.registry, msg.AppID)
		m.screen = ScreenDetail
		return m, nil
	case BackToMenuMsg:
		// Cancel any running async operation before navigating away.
		m.cancelRunning()
		// Return to the appropriate parent screen.
		switch m.screen {
		case ScreenDetail:
			m.screen = ScreenInstallers
		default:
			m.screen = ScreenMenu
		}
		return m, nil
	case StartPreflightMsg:
		// Cancel any prior operation and create a new context.
		m.cancelRunning()
		ctx, cancel := context.WithCancel(context.Background())
		m.cancel = cancel
		m.preflight = NewPreflightModel(ctx, m.registry, msg.AppID)
		m.screen = ScreenPreflight
		return m, m.preflight.Init()
	case StartConfigMsg:
		// Cancel the preflight context (check is done).
		m.cancelRunning()
		m.config = NewConfigModel(msg.AppID)
		m.screen = ScreenConfig
		return m, m.config.Init()
	case ConfigDoneMsg:
		// Create a new context for the installation.
		m.cancelRunning()
		ctx, cancel := context.WithCancel(context.Background())
		m.cancel = cancel
		m.progress = NewProgressModel(ctx, m.registry, msg.AppID, msg.Config)
		m.screen = ScreenInstall
		return m, m.progress.Init()
	case StartVerifyMsg:
		// Create a new context for verification.
		m.cancelRunning()
		ctx, cancel := context.WithCancel(context.Background())
		m.cancel = cancel
		m.verify = NewVerifyModel(ctx, m.registry, msg.AppID)
		m.screen = ScreenVerify
		return m, m.verify.Init()
	case ElevateMsg:
		// Cancel running ops, quit TUI for privilege escalation.
		m.cancelRunning()
		m.elevateRelaunch = true
		m.escalation = msg.Escalation
		m.resumeAppID = msg.AppID
		return m, tea.Quit
	}

	// Delegate to the active screen's sub-model.
	var cmd tea.Cmd
	switch m.screen {
	case ScreenMenu:
		m.menu, cmd = m.menu.Update(msg)
	case ScreenInstallers:
		m.installerList, cmd = m.installerList.Update(msg)
	case ScreenDetail:
		m.detail, cmd = m.detail.Update(msg)
	case ScreenPreflight:
		m.preflight, cmd = m.preflight.Update(msg)
	case ScreenConfig:
		m.config, cmd = m.config.Update(msg)
	case ScreenInstall:
		m.progress, cmd = m.progress.Update(msg)
	case ScreenVerify:
		m.verify, cmd = m.verify.Update(msg)
	case ScreenSettings:
		m.settings, cmd = m.settings.Update(msg)
	case ScreenHelp:
		m.help, cmd = m.help.Update(msg)
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
	case ScreenInstallers:
		return m.installerList.View()
	case ScreenDetail:
		return m.detail.View()
	case ScreenPreflight:
		return m.preflight.View()
	case ScreenConfig:
		return m.config.View()
	case ScreenInstall:
		return m.progress.View()
	case ScreenVerify:
		return m.verify.View()
	case ScreenSettings:
		return m.settings.View()
	case ScreenHelp:
		return m.help.View()
	default:
		return ""
	}
}

// Quitting returns whether the model is in a quitting state.
func (m AppModel) Quitting() bool {
	return m.quitting
}

// ElevateRelaunch returns whether the app should relaunch with
// elevated privileges.
func (m AppModel) ElevateRelaunch() bool {
	return m.elevateRelaunch
}

// Escalation returns the detected privilege escalation method,
// or nil if none was detected.
func (m AppModel) Escalation() *installer.EscalationMethod {
	return m.escalation
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

// ResumeAppID returns the app ID to resume at after an elevated
// relaunch, or an empty string if starting normally.
func (m AppModel) ResumeAppID() string {
	return m.resumeAppID
}

// cancelRunning cancels the context for the current async operation
// (preflight, install, verify) if one is active, preventing orphaned
// goroutines from continuing after the user navigates away or quits.
func (m *AppModel) cancelRunning() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}
