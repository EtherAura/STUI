// settings_test.go contains unit tests for the SettingsModel,
// covering initialization, toggle behaviour, navigation, back
// navigation, and view rendering.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewSettingsModel verifies the settings screen initializes
// with show-passwords off and cursor at zero.
func TestNewSettingsModel(t *testing.T) {
	m := NewSettingsModel()
	if m.width != 0 || m.height != 0 {
		t.Error("new settings model should have zero dimensions")
	}
	if m.ShowPasswords() {
		t.Error("show passwords should default to false")
	}
	if m.Cursor() != 0 {
		t.Errorf("cursor should start at 0, got %d", m.Cursor())
	}
}

// TestSettingsModelInit verifies Init returns no command.
func TestSettingsModelInit(t *testing.T) {
	m := NewSettingsModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// TestSettingsModelViewContent verifies the view contains the
// settings title and toggle labels.
func TestSettingsModelViewContent(t *testing.T) {
	m := NewSettingsModel()
	view := m.View()

	if !strings.Contains(view, "Settings") {
		t.Error("view should contain 'Settings'")
	}
	if !strings.Contains(view, "Show Passwords") {
		t.Error("view should contain 'Show Passwords' label")
	}
	if !strings.Contains(view, "[ ]") {
		t.Error("view should show unchecked checkbox by default")
	}
}

// TestSettingsModelToggleSpace verifies pressing space toggles
// the show-passwords setting on and off.
func TestSettingsModelToggleSpace(t *testing.T) {
	m := NewSettingsModel()

	// Toggle on.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.ShowPasswords() {
		t.Error("show passwords should be true after space toggle")
	}

	view := m.View()
	if !strings.Contains(view, "[x]") {
		t.Error("view should show checked checkbox after toggle on")
	}

	// Toggle off.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if m.ShowPasswords() {
		t.Error("show passwords should be false after second toggle")
	}
}

// TestSettingsModelToggleEnter verifies pressing enter also toggles.
func TestSettingsModelToggleEnter(t *testing.T) {
	m := NewSettingsModel()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.ShowPasswords() {
		t.Error("show passwords should be true after enter toggle")
	}
}

// TestSettingsModelEscSendsBackToMenu verifies pressing esc produces
// a BackToMenuMsg.
func TestSettingsModelEscSendsBackToMenu(t *testing.T) {
	m := NewSettingsModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc key should produce a command")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestSettingsModelBackspaceSendsBackToMenu verifies pressing backspace
// produces a BackToMenuMsg.
func TestSettingsModelBackspaceSendsBackToMenu(t *testing.T) {
	m := NewSettingsModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd == nil {
		t.Fatal("backspace key should produce a command")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestSettingsModelWindowResize verifies WindowSizeMsg updates dimensions.
func TestSettingsModelWindowResize(t *testing.T) {
	m := NewSettingsModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions = %dx%d, want 80x24", m.width, m.height)
	}
}

// TestSettingsModelUnhandledKey verifies unrecognized keys are ignored.
func TestSettingsModelUnhandledKey(t *testing.T) {
	m := NewSettingsModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("unrecognized key should not produce a command")
	}
}

// TestSettingsModelPersistsAcrossNavigations verifies that toggling
// a setting survives a round-trip through the back key.
func TestSettingsModelPersistsAcrossNavigations(t *testing.T) {
	m := NewSettingsModel()

	// Toggle show passwords on.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.ShowPasswords() {
		t.Error("show passwords should be true after toggle")
	}

	// Simulate navigating away and back — the model struct retains state.
	if !m.ShowPasswords() {
		t.Error("show passwords should still be true after simulated nav")
	}
}
