// verify.go implements the post-installation verification screen.
// After installation completes, this screen runs the installer's
// Verify method and displays success/failure with options to return
// to the menu or quit the application.
package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// VerifyDoneMsg carries the result of the verification check.
type VerifyDoneMsg struct {
	// Err is non-nil if verification failed.
	Err error
}

// VerifyModel runs and displays post-install verification results.
type VerifyModel struct {
	// appID is the registry key of the verified application.
	appID string
	// inst is the instantiated installer.
	inst installer.Installer
	// ctx is the cancellable context for the verification check.
	ctx context.Context
	// spinner provides visual feedback while verifying.
	spinner spinner.Model
	// running is true while verification is in progress.
	running bool
	// err holds any verification error.
	err error
	// done indicates verification has completed.
	done bool
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewVerifyModel creates a verification screen for the given app,
// instantiating the installer from the registry. The provided context
// allows the caller to cancel the verification on navigation or quit.
func NewVerifyModel(ctx context.Context, reg installer.Registry, appID string) VerifyModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = BannerStyle

	var inst installer.Installer
	if ctor, ok := reg[appID]; ok {
		inst = ctor()
	}

	return VerifyModel{
		appID:   appID,
		inst:    inst,
		ctx:     ctx,
		spinner: s,
		running: true,
	}
}

// Init implements tea.Model. Starts the spinner and kicks off
// the verification check as an async command.
func (m VerifyModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runVerify(),
	)
}

// runVerify returns a tea.Cmd that executes the installer's Verify
// method and sends a VerifyDoneMsg with the result.
func (m VerifyModel) runVerify() tea.Cmd {
	inst := m.inst
	appID := m.appID
	ctx := m.ctx
	return func() tea.Msg {
		if inst == nil {
			return VerifyDoneMsg{Err: &verifyError{"no installer for app " + appID}}
		}
		err := inst.Verify(ctx)
		return VerifyDoneMsg{Err: err}
	}
}

// verifyError is a simple error type for verification failures.
type verifyError struct {
	msg string
}

// Error implements the error interface.
func (e *verifyError) Error() string { return e.msg }

// Update implements tea.Model. Handles verification completion,
// spinner ticks, and navigation keys.
func (m VerifyModel) Update(msg tea.Msg) (VerifyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case VerifyDoneMsg:
		m.running = false
		m.done = true
		m.err = msg.Err
		return m, nil
	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "enter", "esc", "q":
				return m, func() tea.Msg {
					return BackToMenuMsg{}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		if m.running {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// View implements tea.Model. Renders the verification state:
// spinner while running, results when done.
func (m VerifyModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Verification"))
	b.WriteString("\n")

	if m.running {
		b.WriteString(m.spinner.View())
		b.WriteString(" Verifying installation...")
		return AppStyle.Render(b.String())
	}

	if m.err != nil {
		b.WriteString(ErrorStyle.Render("✗ Verification failed"))
		b.WriteString("\n\n")
		b.WriteString(ErrorStyle.Render("  " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Press any key to return to menu"))
	} else {
		b.WriteString(SuccessStyle.Render("✓ Installation verified successfully!"))
		b.WriteString("\n\n")
		b.WriteString(BodyStyle.Render(m.appName() + " is installed and ready to use."))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Press any key to return to menu"))
	}

	return AppStyle.Render(b.String())
}

// appName returns the display name of the installer, or the app ID
// if the installer is nil.
func (m VerifyModel) appName() string {
	if m.inst != nil {
		return m.inst.Name()
	}
	return m.appID
}

// AppID returns the registry key of the verified application.
func (m VerifyModel) AppID() string {
	return m.appID
}

// Running returns whether verification is still in progress.
func (m VerifyModel) Running() bool {
	return m.running
}

// Done returns whether verification has completed.
func (m VerifyModel) Done() bool {
	return m.done
}

// Err returns any verification error, or nil on success.
func (m VerifyModel) Err() error {
	return m.err
}
