// progress_test.go contains unit tests for the ProgressModel, covering
// initialization, output streaming, step progression, completion states,
// and navigation key handling.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// newTestConfig creates a minimal valid config for testing.
func newTestConfig() *installer.Config {
	return &installer.Config{
		SonarURL: "https://test.sonar.software",
		Domain:   "portal.example.com",
	}
}

// TestNewProgressModel verifies a new progress model starts in the
// running state with the correct app ID and step count.
func TestNewProgressModel(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)

	if m.AppID() != installer.AppCustomerPortal {
		t.Errorf("AppID() = %q, want %q", m.AppID(), installer.AppCustomerPortal)
	}
	if !m.Running() {
		t.Error("new progress model should be running")
	}
	if m.Done() {
		t.Error("new progress model should not be done")
	}
	if m.totalSteps != 3 {
		t.Errorf("portal should have 3 steps, got %d", m.totalSteps)
	}
}

// TestNewProgressModelUnknown verifies handling of an unknown app ID.
func TestNewProgressModelUnknown(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, "unknown-app", cfg)

	if m.inst != nil {
		t.Error("installer should be nil for unknown app")
	}
	if m.totalSteps != 0 {
		t.Error("unknown app should have 0 steps")
	}
}

// TestProgressModelInit verifies Init returns commands.
func TestProgressModelInit(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a batch command")
	}
}

// TestProgressModelOutputMsg verifies output messages append to viewport.
func TestProgressModelOutputMsg(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)

	m, _ = m.Update(InstallOutputMsg{Output: "line 1\n"})
	m, _ = m.Update(InstallOutputMsg{Output: "line 2\n"})

	if !strings.Contains(m.output.String(), "line 1") {
		t.Error("output should contain 'line 1'")
	}
	if !strings.Contains(m.output.String(), "line 2") {
		t.Error("output should contain 'line 2'")
	}
}

// TestProgressModelStepMsg verifies step messages update the step info.
func TestProgressModelStepMsg(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)

	m, _ = m.Update(InstallStepMsg{StepIndex: 1, StepName: "Clone repo", TotalSteps: 3})

	if m.stepIndex != 1 {
		t.Errorf("stepIndex = %d, want 1", m.stepIndex)
	}
	if m.currentStep != "Clone repo" {
		t.Errorf("currentStep = %q, want %q", m.currentStep, "Clone repo")
	}
}

// TestProgressModelDoneSuccess verifies successful completion.
func TestProgressModelDoneSuccess(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)

	m, _ = m.Update(InstallDoneMsg{Err: nil})

	if m.Running() {
		t.Error("should not be running after done")
	}
	if !m.Done() {
		t.Error("should be done")
	}
	if m.Err() != nil {
		t.Errorf("err should be nil, got %v", m.Err())
	}
	if !strings.Contains(m.output.String(), "complete") {
		t.Error("output should contain completion message")
	}
}

// TestProgressModelDoneError verifies error completion.
func TestProgressModelDoneError(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)

	m, _ = m.Update(InstallDoneMsg{Err: errTestPreflight})

	if !m.Done() {
		t.Error("should be done")
	}
	if m.Err() == nil {
		t.Error("err should be set")
	}
	if !strings.Contains(m.output.String(), "failed") {
		t.Error("output should contain failure message")
	}
}

// TestProgressModelEnterOnSuccess verifies enter produces StartVerifyMsg
// when installation succeeded.
func TestProgressModelEnterOnSuccess(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)
	m, _ = m.Update(InstallDoneMsg{Err: nil})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a command on success")
	}
	msg := cmd()
	verify, ok := msg.(StartVerifyMsg)
	if !ok {
		t.Fatalf("expected StartVerifyMsg, got %T", msg)
	}
	if verify.AppID != installer.AppCustomerPortal {
		t.Errorf("AppID = %q, want %q", verify.AppID, installer.AppCustomerPortal)
	}
}

// TestProgressModelEnterOnError verifies enter does nothing on error.
func TestProgressModelEnterOnError(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)
	m, _ = m.Update(InstallDoneMsg{Err: errTestPreflight})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter should not produce a command on error")
	}
}

// TestProgressModelEscOnDone verifies esc goes back to menu.
func TestProgressModelEscOnDone(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)
	m, _ = m.Update(InstallDoneMsg{Err: nil})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a command")
	}
	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestProgressModelWindowResize verifies dimensions are updated.
func TestProgressModelWindowResize(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m.width != 100 || m.height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", m.width, m.height)
	}
}

// TestProgressModelViewRunning verifies the running view shows a spinner.
func TestProgressModelViewRunning(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)
	view := m.View()

	if !strings.Contains(view, "Installing") {
		t.Error("running view should contain 'Installing'")
	}
}

// TestProgressModelViewDone verifies the done view shows completion status.
func TestProgressModelViewDone(t *testing.T) {
	reg := installer.NewRegistry()
	cfg := newTestConfig()
	m := NewProgressModel(reg, installer.AppCustomerPortal, cfg)
	m, _ = m.Update(InstallDoneMsg{Err: nil})
	view := m.View()

	if !strings.Contains(view, "complete") {
		t.Error("done view should contain 'complete'")
	}
	if !strings.Contains(view, "verify") {
		t.Error("done view should mention verification")
	}
}

// TestAppModelTransitionToInstall verifies ConfigDoneMsg transitions
// to the install screen.
func TestAppModelTransitionToInstall(t *testing.T) {
	m := NewAppModel()
	cfg := newTestConfig()

	updated, _ := m.Update(ConfigDoneMsg{
		AppID:  installer.AppCustomerPortal,
		Config: cfg,
	})
	model := updated.(AppModel)

	if model.Screen() != ScreenInstall {
		t.Errorf("screen = %d, want ScreenInstall (%d)", model.Screen(), ScreenInstall)
	}
}
