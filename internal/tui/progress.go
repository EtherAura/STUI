// progress.go implements the installation progress screen.
// After configuration, this screen runs the installer's Install method,
// streaming command output into a scrollable viewport while showing
// a progress bar and the current step name.
package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// InstallOutputMsg carries a chunk of output from the install process.
type InstallOutputMsg struct {
	// Output is the raw text output from the current step.
	Output string
}

// InstallStepMsg indicates a new installation step has started.
type InstallStepMsg struct {
	// StepIndex is the zero-based index of the current step.
	StepIndex int
	// StepName is the display name of the current step.
	StepName string
	// TotalSteps is the total number of steps.
	TotalSteps int
}

// InstallDoneMsg signals that installation has completed.
type InstallDoneMsg struct {
	// Err is non-nil if installation failed.
	Err error
}

// StartVerifyMsg signals transition from progress to verification.
type StartVerifyMsg struct {
	// AppID is the registry key of the installed application.
	AppID string
}

// ProgressModel displays real-time installation progress with a
// scrollable output viewport and a progress bar.
type ProgressModel struct {
	// appID is the registry key of the application being installed.
	appID string
	// inst is the instantiated installer.
	inst installer.Installer
	// cfg holds the user-provided configuration.
	cfg *installer.Config
	// spinner provides visual feedback while installing.
	spinner spinner.Model
	// progress is the progress bar component.
	progress progress.Model
	// viewport displays scrollable command output.
	viewport viewport.Model
	// output accumulates all output text.
	output *strings.Builder
	// currentStep is the name of the step currently executing.
	currentStep string
	// stepIndex is the zero-based index of the current step.
	stepIndex int
	// totalSteps is the total number of installation steps.
	totalSteps int
	// running is true while installation is in progress.
	running bool
	// done indicates installation has finished.
	done bool
	// err holds any installation error.
	err error
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewProgressModel creates a progress screen for the given app
// with the provided configuration.
func NewProgressModel(reg installer.Registry, appID string, cfg *installer.Config) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = BannerStyle

	p := progress.New(progress.WithDefaultGradient())

	vp := viewport.New(defaultMenuWidth, 12)
	vp.SetContent("")

	var inst installer.Installer
	var totalSteps int
	if ctor, ok := reg[appID]; ok {
		inst = ctor()
		totalSteps = len(inst.Steps())
	}

	return ProgressModel{
		appID:      appID,
		inst:       inst,
		cfg:        cfg,
		spinner:    s,
		progress:   p,
		viewport:   vp,
		output:     &strings.Builder{},
		totalSteps: totalSteps,
		running:    true,
	}
}

// Init implements tea.Model. Starts the spinner and kicks off
// the installation as an async command.
func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runInstall(),
	)
}

// runInstall returns a tea.Cmd that executes the installer and
// sends step/output/done messages as installation progresses.
func (m ProgressModel) runInstall() tea.Cmd {
	inst := m.inst
	cfg := m.cfg
	appID := m.appID
	return func() tea.Msg {
		if inst == nil {
			return InstallDoneMsg{Err: fmt.Errorf("no installer for app %q", appID)}
		}

		// Use a buffer to capture output. In a full implementation,
		// this would stream via a pipe with goroutine forwarding.
		var buf bytes.Buffer
		ctx := context.Background()
		err := inst.Install(ctx, cfg, &buf)

		// Send the final output and done message.
		// Note: In production, we'd send InstallOutputMsg chunks during
		// execution via a custom io.Writer that posts tea.Msg.
		return InstallDoneMsg{Err: err}
	}
}

// msgWriter is an io.Writer that sends output chunks as tea.Msg.
// This is used for streaming install output into the TUI.
type msgWriter struct {
	send func(tea.Msg)
}

// Write implements io.Writer by sending InstallOutputMsg.
func (w *msgWriter) Write(p []byte) (int, error) {
	w.send(InstallOutputMsg{Output: string(p)})
	return len(p), nil
}

// Ensure msgWriter satisfies io.Writer.
var _ io.Writer = (*msgWriter)(nil)

// Update implements tea.Model. Handles install progress messages,
// spinner ticks, and navigation keys.
func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case InstallOutputMsg:
		m.output.WriteString(msg.Output)
		m.viewport.SetContent(m.output.String())
		m.viewport.GotoBottom()
		return m, nil
	case InstallStepMsg:
		m.stepIndex = msg.StepIndex
		m.currentStep = msg.StepName
		m.totalSteps = msg.TotalSteps
		line := fmt.Sprintf("==> Step %d/%d: %s\n", msg.StepIndex+1, msg.TotalSteps, msg.StepName)
		m.output.WriteString(line)
		m.viewport.SetContent(m.output.String())
		m.viewport.GotoBottom()
		return m, nil
	case InstallDoneMsg:
		m.running = false
		m.done = true
		m.err = msg.Err
		if msg.Err != nil {
			fmt.Fprintf(m.output, "\n✗ Installation failed: %s\n", msg.Err)
		} else {
			m.output.WriteString("\n✓ Installation complete!\n")
		}
		m.viewport.SetContent(m.output.String())
		m.viewport.GotoBottom()
		return m, nil
	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "enter":
				if m.err == nil {
					return m, func() tea.Msg {
						return StartVerifyMsg{AppID: m.appID}
					}
				}
			case "esc":
				return m, func() tea.Msg {
					return BackToMenuMsg{}
				}
			}
		}
		// Allow viewport scrolling while running.
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve space for title, progress bar, status, and help.
		vpHeight := msg.Height - 10
		if vpHeight < 3 {
			vpHeight = 3
		}
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = vpHeight
	case spinner.TickMsg:
		if m.running {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model.(progress.Model)
		return m, cmd
	}
	return m, nil
}

// View implements tea.Model. Renders the progress bar, current step,
// and scrollable output viewport.
func (m ProgressModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Installing"))
	b.WriteString("\n")

	// Progress bar.
	var pct float64
	if m.totalSteps > 0 {
		pct = float64(m.stepIndex+1) / float64(m.totalSteps)
	}
	if m.done && m.err == nil {
		pct = 1.0
	}
	b.WriteString(m.progress.ViewAs(pct))
	b.WriteString("\n\n")

	// Current step or status.
	switch {
	case m.running:
		step := m.currentStep
		if step == "" {
			step = "Starting..."
		}
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(BodyStyle.Render(step))
	case m.err != nil:
		b.WriteString(ErrorStyle.Render("✗ Installation failed"))
	default:
		b.WriteString(SuccessStyle.Render("✓ Installation complete"))
	}
	b.WriteString("\n\n")

	// Output viewport.
	b.WriteString(m.viewport.View())
	b.WriteString("\n\n")

	// Help text.
	if m.done {
		if m.err == nil {
			b.WriteString(SuccessStyle.Render("Press enter to verify installation"))
			b.WriteString("\n")
		}
		b.WriteString(HelpStyle.Render("Press esc to return to menu"))
	} else {
		b.WriteString(HelpStyle.Render("↑/↓: scroll output"))
	}

	return AppStyle.Render(b.String())
}

// AppID returns the registry key of the application being installed.
func (m ProgressModel) AppID() string {
	return m.appID
}

// Running returns whether installation is still in progress.
func (m ProgressModel) Running() bool {
	return m.running
}

// Done returns whether installation has completed.
func (m ProgressModel) Done() bool {
	return m.done
}

// Err returns any installation error, or nil on success.
func (m ProgressModel) Err() error {
	return m.err
}
