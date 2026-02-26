// preflight_tui_test.go contains unit tests for the PreflightModel,
// covering initialization, spinner behavior, result rendering for
// pass/fail/error states, and key navigation.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// TestNewPreflightModel verifies a new preflight model starts in
// the running state with the correct app ID.
func TestNewPreflightModel(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)

	if m.AppID() != installer.AppCustomerPortal {
		t.Errorf("AppID() = %q, want %q", m.AppID(), installer.AppCustomerPortal)
	}
	if !m.Running() {
		t.Error("new preflight model should be running")
	}
	if m.Result() != nil {
		t.Error("new preflight model should have nil result")
	}
}

// TestNewPreflightModelUnknown verifies handling of an unknown app ID.
func TestNewPreflightModelUnknown(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, "unknown-app")

	if m.inst != nil {
		t.Error("installer should be nil for unknown app")
	}
}

// TestPreflightModelInit verifies Init returns commands (spinner + check).
func TestPreflightModelInit(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a batch command")
	}
}

// TestPreflightModelDoneMsg verifies that PreflightDoneMsg stops the
// spinner and stores the result.
func TestPreflightModelDoneMsg(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)

	result := &installer.PreflightResult{
		Passed:  true,
		OS:      "ubuntu",
		Version: "24.04",
	}

	m, _ = m.Update(PreflightDoneMsg{Result: result})

	if m.Running() {
		t.Error("model should not be running after PreflightDoneMsg")
	}
	if m.Result() == nil {
		t.Fatal("result should not be nil")
	}
	if !m.Result().Passed {
		t.Error("result should be passed")
	}
}

// TestPreflightModelDoneMsgError verifies error handling when the
// preflight check itself fails.
func TestPreflightModelDoneMsgError(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)

	m, _ = m.Update(PreflightDoneMsg{Err: errTestPreflight})

	if m.Running() {
		t.Error("model should not be running after error")
	}
	if m.err == nil {
		t.Error("err should be set")
	}
}

// errTestPreflight is a sentinel error for testing.
var errTestPreflight = &testError{"test preflight error"}

// testError is a simple error type for testing.
type testError struct{ msg string }

// Error implements the error interface.
func (e *testError) Error() string { return e.msg }

// TestPreflightViewRunning verifies the running view shows a spinner.
func TestPreflightViewRunning(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	view := m.View()

	if !strings.Contains(view, "Running") {
		t.Error("running view should contain 'Running'")
	}
	if !strings.Contains(view, "Preflight") {
		t.Error("running view should contain 'Preflight'")
	}
}

// TestPreflightViewRequirements verifies the requirements section is
// displayed for the selected application.
func TestPreflightViewRequirements(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:  true,
		OS:      "ubuntu",
		Version: "24.04",
	}})
	view := m.View()

	if !strings.Contains(view, "Requirements") {
		t.Error("view should contain Requirements heading")
	}
	if !strings.Contains(view, "Ubuntu") {
		t.Error("view should list OS requirement")
	}
	if !strings.Contains(view, "git") {
		t.Error("view should list required commands")
	}
}

// TestPreflightViewIconAlignment verifies that status icons use
// consistent spacing for visual alignment.
func TestPreflightViewIconAlignment(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:   true,
		OS:       "ubuntu",
		Version:  "24.04",
		Warnings: []string{"test warning"},
	}})
	view := m.View()

	// Both ✓ and ⚠ should have double-space after the icon.
	if !strings.Contains(view, "✓  ") {
		t.Error("pass icon should have double-space after it")
	}
	if !strings.Contains(view, "⚠  ") {
		t.Error("warning icon should have double-space after it")
	}
}

// TestPreflightViewPassed verifies the passed view shows success indicators.
func TestPreflightViewPassed(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:  true,
		OS:      "ubuntu",
		Version: "24.04",
	}})
	view := m.View()

	if !strings.Contains(view, "All checks passed") {
		t.Error("passed view should contain success message")
	}
	if !strings.Contains(view, "ubuntu") {
		t.Error("passed view should show OS")
	}
	if !strings.Contains(view, "enter") {
		t.Error("passed view should mention enter to continue")
	}
}

// TestPreflightViewFailed verifies the failed view shows error details.
func TestPreflightViewFailed(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed: false,
		OS:     "ubuntu",
		Errors: []string{"required command not found: git"},
	}})
	view := m.View()

	if !strings.Contains(view, "required command") {
		t.Error("failed view should show error message")
	}
	if !strings.Contains(view, "failed") {
		t.Error("failed view should indicate failure")
	}
}

// TestPreflightViewWarnings verifies warnings are displayed.
func TestPreflightViewWarnings(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:   true,
		OS:       "ubuntu",
		Version:  "24.04",
		Warnings: []string{"not running as root"},
	}})
	view := m.View()

	if !strings.Contains(view, "not running as root") {
		t.Error("view should show warning text")
	}
}

// TestPreflightViewError verifies the error view when the check fails to run.
func TestPreflightViewError(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Err: errTestPreflight})
	view := m.View()

	if !strings.Contains(view, "test preflight error") {
		t.Error("error view should show the error message")
	}
}

// TestPreflightEnterOnPassed verifies enter produces StartConfigMsg
// when checks passed.
func TestPreflightEnterOnPassed(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed: true,
	}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a command when checks passed")
	}

	msg := cmd()
	cfg, ok := msg.(StartConfigMsg)
	if !ok {
		t.Fatalf("expected StartConfigMsg, got %T", msg)
	}
	if cfg.AppID != installer.AppCustomerPortal {
		t.Errorf("AppID = %q, want %q", cfg.AppID, installer.AppCustomerPortal)
	}
}

// TestPreflightEnterOnFailed verifies enter does nothing when checks failed.
func TestPreflightEnterOnFailed(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed: false,
	}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter should not produce a command when checks failed")
	}
}

// TestPreflightEnterWhileRunning verifies enter is ignored while running.
func TestPreflightEnterWhileRunning(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter should be ignored while running")
	}
}

// TestPreflightEscKey verifies esc goes back to menu.
func TestPreflightEscKey(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{Passed: true}})

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

// TestPreflightWindowResize verifies WindowSizeMsg updates dimensions.
func TestPreflightWindowResize(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions = %dx%d, want 80x24", m.width, m.height)
	}
}

// TestAppModelTransitionToPreflight verifies StartPreflightMsg
// transitions from detail to preflight screen.
func TestAppModelTransitionToPreflight(t *testing.T) {
	m := NewAppModel()

	// Go to detail first.
	updated, _ := m.Update(AppSelectedMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)

	// Then to preflight.
	updated, _ = model.Update(StartPreflightMsg{AppID: installer.AppCustomerPortal})
	model = updated.(AppModel)

	if model.Screen() != ScreenPreflight {
		t.Errorf("screen = %d, want ScreenPreflight (%d)", model.Screen(), ScreenPreflight)
	}
}

// TestAppModelPreflightBackToMenu verifies BackToMenuMsg from preflight
// returns to the menu screen.
func TestAppModelPreflightBackToMenu(t *testing.T) {
	m := NewAppModel()

	// Navigate to preflight.
	updated, _ := m.Update(AppSelectedMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)
	updated, _ = model.Update(StartPreflightMsg{AppID: installer.AppCustomerPortal})
	model = updated.(AppModel)

	// Go back.
	updated, _ = model.Update(BackToMenuMsg{})
	model = updated.(AppModel)

	if model.Screen() != ScreenMenu {
		t.Errorf("screen = %d, want ScreenMenu (%d)", model.Screen(), ScreenMenu)
	}
}

// TestPreflightSudoKeyOnNeedsRoot verifies pressing 's' produces
// SudoRelaunchMsg when preflight reports NeedsRoot.
func TestPreflightSudoKeyOnNeedsRoot(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:    true,
		OS:        "ubuntu",
		NeedsRoot: true,
	}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Fatal("pressing 's' should produce a command when NeedsRoot is true")
	}

	msg := cmd()
	_, ok := msg.(SudoRelaunchMsg)
	if !ok {
		t.Fatalf("expected SudoRelaunchMsg, got %T", msg)
	}
}

// TestPreflightSudoKeyIgnoredWhenRoot verifies pressing 's' does nothing
// when NeedsRoot is false.
func TestPreflightSudoKeyIgnoredWhenRoot(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:    true,
		OS:        "ubuntu",
		NeedsRoot: false,
	}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd != nil {
		t.Error("pressing 's' should be ignored when NeedsRoot is false")
	}
}

// TestPreflightViewNeedsRoot verifies the view shows a sudo relaunch
// option when NeedsRoot is set.
func TestPreflightViewNeedsRoot(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:    true,
		OS:        "ubuntu",
		NeedsRoot: true,
		Warnings:  []string{"not running as root; elevated privileges are required"},
	}})
	view := m.View()

	if !strings.Contains(view, "sudo") {
		t.Error("view should mention sudo when NeedsRoot is true")
	}
	if !strings.Contains(view, "Press s") {
		t.Error("view should show 'Press s' option when NeedsRoot is true")
	}
}

// TestPreflightViewNoSudoWhenRoot verifies the view does not show
// a sudo option when NeedsRoot is false.
func TestPreflightViewNoSudoWhenRoot(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewPreflightModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(PreflightDoneMsg{Result: &installer.PreflightResult{
		Passed:    true,
		OS:        "ubuntu",
		NeedsRoot: false,
	}})
	view := m.View()

	if strings.Contains(view, "Press s") {
		t.Error("view should not show 'Press s' option when NeedsRoot is false")
	}
}

// TestAppModelSudoRelaunch verifies SudoRelaunchMsg sets the flag
// and triggers tea.Quit.
func TestAppModelSudoRelaunch(t *testing.T) {
	m := NewAppModel()

	// Navigate to preflight.
	updated, _ := m.Update(AppSelectedMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)
	updated, _ = model.Update(StartPreflightMsg{AppID: installer.AppCustomerPortal})
	model = updated.(AppModel)

	// Simulate sudo relaunch.
	updated, cmd := model.Update(SudoRelaunchMsg{})
	model = updated.(AppModel)

	if !model.SudoRelaunch() {
		t.Error("SudoRelaunch() should be true after SudoRelaunchMsg")
	}
	if cmd == nil {
		t.Error("SudoRelaunchMsg should produce a quit command")
	}
}
