// settings_test.go contains unit tests for the SettingsModel,
// covering initialization, back navigation, and view rendering.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewSettingsModel verifies the settings screen initializes
// with zero dimensions.
func TestNewSettingsModel(t *testing.T) {
	m := NewSettingsModel()
	if m.width != 0 || m.height != 0 {
		t.Error("new settings model should have zero dimensions")
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
// expected placeholder text.
func TestSettingsModelViewContent(t *testing.T) {
	m := NewSettingsModel()
	view := m.View()

	if !strings.Contains(view, "Settings") {
		t.Error("view should contain 'Settings'")
	}
	if !strings.Contains(view, "No settings available yet") {
		t.Error("view should contain placeholder text")
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
