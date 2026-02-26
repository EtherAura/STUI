// verify_test.go contains unit tests for the VerifyModel, covering
// initialization, async verification, result rendering for success
// and failure, and navigation key handling.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// TestNewVerifyModel verifies a new verify model starts in the
// running state with the correct app ID.
func TestNewVerifyModel(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)

	if m.AppID() != installer.AppCustomerPortal {
		t.Errorf("AppID() = %q, want %q", m.AppID(), installer.AppCustomerPortal)
	}
	if !m.Running() {
		t.Error("new verify model should be running")
	}
	if m.Done() {
		t.Error("new verify model should not be done")
	}
}

// TestNewVerifyModelUnknown verifies handling of an unknown app ID.
func TestNewVerifyModelUnknown(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, "unknown-app")

	if m.inst != nil {
		t.Error("installer should be nil for unknown app")
	}
}

// TestVerifyModelInit verifies Init returns commands.
func TestVerifyModelInit(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a batch command")
	}
}

// TestVerifyModelDoneSuccess verifies successful verification.
func TestVerifyModelDoneSuccess(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)

	m, _ = m.Update(VerifyDoneMsg{Err: nil})

	if m.Running() {
		t.Error("should not be running after done")
	}
	if !m.Done() {
		t.Error("should be done")
	}
	if m.Err() != nil {
		t.Errorf("err should be nil, got %v", m.Err())
	}
}

// TestVerifyModelDoneError verifies error verification.
func TestVerifyModelDoneError(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)

	m, _ = m.Update(VerifyDoneMsg{Err: errTestPreflight})

	if !m.Done() {
		t.Error("should be done")
	}
	if m.Err() == nil {
		t.Error("err should be set")
	}
}

// TestVerifyViewRunning verifies the running view shows a spinner.
func TestVerifyViewRunning(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	view := m.View()

	if !strings.Contains(view, "Verifying") {
		t.Error("running view should contain 'Verifying'")
	}
}

// TestVerifyViewSuccess verifies the success view.
func TestVerifyViewSuccess(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(VerifyDoneMsg{Err: nil})
	view := m.View()

	if !strings.Contains(view, "verified") {
		t.Error("success view should contain 'verified'")
	}
	if !strings.Contains(view, "Customer Portal") {
		t.Error("success view should contain app name")
	}
}

// TestVerifyViewError verifies the error view.
func TestVerifyViewError(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(VerifyDoneMsg{Err: errTestPreflight})
	view := m.View()

	if !strings.Contains(view, "failed") {
		t.Error("error view should contain 'failed'")
	}
	if !strings.Contains(view, "test preflight error") {
		t.Error("error view should contain error message")
	}
}

// TestVerifyEnterReturnsToMenu verifies enter goes back to menu.
func TestVerifyEnterReturnsToMenu(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(VerifyDoneMsg{Err: nil})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a command")
	}
	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestVerifyEscReturnsToMenu verifies esc goes back to menu.
func TestVerifyEscReturnsToMenu(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(VerifyDoneMsg{Err: nil})

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

// TestVerifyQKeyReturnsToMenu verifies q goes back to menu.
func TestVerifyQKeyReturnsToMenu(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)
	m, _ = m.Update(VerifyDoneMsg{Err: nil})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should produce a command")
	}
	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestVerifyKeysIgnoredWhileRunning verifies keys are ignored while running.
func TestVerifyKeysIgnoredWhileRunning(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter should be ignored while running")
	}
}

// TestVerifyWindowResize verifies dimensions are updated.
func TestVerifyWindowResize(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m.width != 100 || m.height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", m.width, m.height)
	}
}

// TestVerifyAppName verifies appName returns the installer name.
func TestVerifyAppName(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, installer.AppCustomerPortal)

	if m.appName() != "Customer Portal" {
		t.Errorf("appName() = %q, want %q", m.appName(), "Customer Portal")
	}
}

// TestVerifyAppNameUnknown verifies appName falls back to app ID.
func TestVerifyAppNameUnknown(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewVerifyModel(reg, "unknown")

	if m.appName() != "unknown" {
		t.Errorf("appName() = %q, want %q", m.appName(), "unknown")
	}
}

// TestAppModelTransitionToVerify verifies StartVerifyMsg transitions
// to the verify screen.
func TestAppModelTransitionToVerify(t *testing.T) {
	m := NewAppModel()

	updated, _ := m.Update(StartVerifyMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)

	if model.Screen() != ScreenVerify {
		t.Errorf("screen = %d, want ScreenVerify (%d)", model.Screen(), ScreenVerify)
	}
}

// TestAppModelVerifyBackToMenu verifies BackToMenuMsg from verify
// returns to menu.
func TestAppModelVerifyBackToMenu(t *testing.T) {
	m := NewAppModel()

	// Go to verify.
	updated, _ := m.Update(StartVerifyMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)

	// Go back.
	updated, _ = model.Update(BackToMenuMsg{})
	model = updated.(AppModel)

	if model.Screen() != ScreenMenu {
		t.Errorf("screen = %d, want ScreenMenu (%d)", model.Screen(), ScreenMenu)
	}
}
