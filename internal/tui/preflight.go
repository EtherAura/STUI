// preflight.go implements the preflight check screen.
// After the user confirms an app on the detail screen, this screen
// runs the installer's PreflightCheck asynchronously, shows progress
// with a spinner, and displays pass/fail results. On all-pass, the
// user can proceed; on failure, blockers are shown. If the process
// is not running as root, the user is offered a sudo relaunch.
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// PreflightDoneMsg carries the result of a completed preflight check.
type PreflightDoneMsg struct {
	// Result holds the preflight check outcome.
	Result *installer.PreflightResult
	// Err is non-nil if the check itself failed to run.
	Err error
}

// StartConfigMsg signals transition from preflight to the config wizard.
type StartConfigMsg struct {
	// AppID is the registry key of the selected application.
	AppID string
}

// ElevateMsg signals that the application should quit and
// relaunch itself with the detected escalation command (sudo/doas).
type ElevateMsg struct {
	// Escalation holds the detected privilege escalation method.
	Escalation *installer.EscalationMethod
	// AppID is the registry key of the installer the user was viewing,
	// so the relaunched process can resume at the same screen.
	AppID string
}

// PreflightModel runs and displays preflight check results for a
// selected application installer.
type PreflightModel struct {
	// appID is the registry key of the selected application.
	appID string
	// inst is the instantiated installer.
	inst installer.Installer
	// spinner provides visual feedback while the check is running.
	spinner spinner.Model
	// running is true while the preflight check is in progress.
	running bool
	// result holds the completed preflight check results.
	result *installer.PreflightResult
	// err holds any error from the preflight check execution.
	err error
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewPreflightModel creates a preflight screen for the given app,
// instantiating the installer from the registry.
func NewPreflightModel(reg installer.Registry, appID string) PreflightModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = BannerStyle

	var inst installer.Installer
	if ctor, ok := reg[appID]; ok {
		inst = ctor()
	}

	return PreflightModel{
		appID:   appID,
		inst:    inst,
		spinner: s,
		running: true,
	}
}

// Init implements tea.Model. Starts the spinner and kicks off the
// preflight check as an async command.
func (m PreflightModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runPreflight(),
	)
}

// runPreflight returns a tea.Cmd that executes the installer's
// PreflightCheck and sends a PreflightDoneMsg with the result.
func (m PreflightModel) runPreflight() tea.Cmd {
	inst := m.inst
	return func() tea.Msg {
		if inst == nil {
			return PreflightDoneMsg{
				Err: fmt.Errorf("no installer for app %q", m.appID),
			}
		}
		result, err := inst.PreflightCheck(context.Background())
		return PreflightDoneMsg{Result: result, Err: err}
	}
}

// Update implements tea.Model. Handles spinner ticks, preflight
// completion, and navigation keys including sudo relaunch.
func (m PreflightModel) Update(msg tea.Msg) (PreflightModel, tea.Cmd) {
	switch msg := msg.(type) {
	case PreflightDoneMsg:
		m.running = false
		m.result = msg.Result
		m.err = msg.Err
		return m, nil
	case tea.KeyMsg:
		if !m.running {
			switch msg.String() {
			case "enter":
				// Only proceed if checks passed.
				if m.result != nil && m.result.Passed {
					return m, func() tea.Msg {
						return StartConfigMsg{AppID: m.appID}
					}
				}
			case "s":
				// Offer escalation relaunch when not running as root.
				if m.result != nil && m.result.NeedsRoot && m.result.Escalation != nil {
					esc := m.result.Escalation
					appID := m.appID
					return m, func() tea.Msg {
						return ElevateMsg{Escalation: esc, AppID: appID}
					}
				}
			case "esc", "backspace":
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

// View implements tea.Model. Renders the preflight check state:
// spinner while running, results when done.
func (m PreflightModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Preflight Checks"))
	b.WriteString("\n")

	if m.running {
		b.WriteString(m.spinner.View())
		b.WriteString(" Running preflight checks...")
		return AppStyle.Render(b.String())
	}

	// Handle execution error.
	if m.err != nil {
		b.WriteString(ErrorStyle.Render("✗ Preflight check failed: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Press esc to go back"))
		return AppStyle.Render(b.String())
	}

	// Display results.
	if m.result != nil {
		// OS info.
		b.WriteString(BodyStyle.Render(fmt.Sprintf("OS: %s %s", m.result.OS, m.result.Version)))
		b.WriteString("\n\n")

		// Hardware checks section — always show detected vs required.
		if m.result.Hardware != nil && m.result.HardwareReqs != nil {
			hw := m.result.Hardware
			reqs := m.result.HardwareReqs

			b.WriteString(BannerStyle.Render("Hardware Checks"))
			b.WriteString("\n")

			// CPU cores.
			if hw.CPUCores >= reqs.MinCPUCores {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf(
					"  ✓  CPU: %d cores (%d required)", hw.CPUCores, reqs.MinCPUCores)))
			} else {
				b.WriteString(WarningStyle.Render(fmt.Sprintf(
					"  ⚠  CPU: %d cores (%d recommended)", hw.CPUCores, reqs.MinCPUCores)))
			}
			b.WriteString("\n")

			// RAM.
			if hw.RAMMB >= reqs.MinRAMMB {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf(
					"  ✓  RAM: %d MB (%d MB required)", hw.RAMMB, reqs.MinRAMMB)))
			} else {
				b.WriteString(WarningStyle.Render(fmt.Sprintf(
					"  ⚠  RAM: %d MB (%d MB recommended)", hw.RAMMB, reqs.MinRAMMB)))
			}
			b.WriteString("\n")

			// Disk space.
			if hw.DiskFreeGB >= reqs.MinDiskGB {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf(
					"  ✓  Disk: %d GB free (%d GB required)", hw.DiskFreeGB, reqs.MinDiskGB)))
			} else {
				b.WriteString(WarningStyle.Render(fmt.Sprintf(
					"  ⚠  Disk: %d GB free (%d GB recommended)", hw.DiskFreeGB, reqs.MinDiskGB)))
			}
			b.WriteString("\n\n")
		}

		// Errors (blocking).
		if len(m.result.Errors) > 0 {
			b.WriteString(ErrorStyle.Render("Blocking Issues:"))
			b.WriteString("\n")
			for _, e := range m.result.Errors {
				b.WriteString(ErrorStyle.Render("  ✗  " + e))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		// Warnings (non-blocking).
		if len(m.result.Warnings) > 0 {
			b.WriteString(WarningStyle.Render("Warnings:"))
			b.WriteString("\n")
			for _, w := range m.result.Warnings {
				b.WriteString(WarningStyle.Render("  ⚠  " + w))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		// Overall status.
		if m.result.Passed {
			if len(m.result.Warnings) > 0 {
				b.WriteString(WarningStyle.Render("  ✓  Checks passed with warnings"))
			} else {
				b.WriteString(SuccessStyle.Render("  ✓  All checks passed"))
			}
			b.WriteString("\n\n")
			b.WriteString(SuccessStyle.Render("Press enter to continue"))
			b.WriteString("\n")
		} else {
			b.WriteString(ErrorStyle.Render("  ✗  Preflight checks failed"))
			b.WriteString("\n\n")
		}

		// Escalation relaunch option when not running as root.
		if m.result.NeedsRoot {
			if m.result.Escalation != nil {
				b.WriteString(WarningStyle.Render(
					fmt.Sprintf("Press s to restart with %s", m.result.Escalation.Name)))
			} else {
				b.WriteString(WarningStyle.Render(
					"No privilege escalation command found (install sudo or doas)"))
			}
			b.WriteString("\n")
		}

		b.WriteString(HelpStyle.Render("Press esc to go back"))
	}

	return AppStyle.Render(b.String())
}

// AppID returns the registry key of the application being checked.
func (m PreflightModel) AppID() string {
	return m.appID
}

// Running returns whether the preflight check is still in progress.
func (m PreflightModel) Running() bool {
	return m.running
}

// Result returns the preflight check result, or nil if not yet complete.
func (m PreflightModel) Result() *installer.PreflightResult {
	return m.result
}
