// help_test.go contains unit tests for the HelpModel,
// covering initialization, back navigation via esc and q keys,
// and view rendering with keyboard shortcuts and about info.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewHelpModel verifies the help screen initializes with
// zero dimensions.
func TestNewHelpModel(t *testing.T) {
	m := NewHelpModel()
	if m.width != 0 || m.height != 0 {
		t.Error("new help model should have zero dimensions")
	}
}

// TestHelpModelInit verifies Init returns no command.
func TestHelpModelInit(t *testing.T) {
	m := NewHelpModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// TestHelpModelViewContainsShortcuts verifies the view contains
// keyboard shortcut documentation.
func TestHelpModelViewContainsShortcuts(t *testing.T) {
	m := NewHelpModel()
	view := m.View()

	shortcuts := []string{"enter", "esc", "ctrl+c", "Keyboard Shortcuts"}
	for _, s := range shortcuts {
		if !strings.Contains(view, s) {
			t.Errorf("view should contain %q", s)
		}
	}
}

// TestHelpModelViewContainsAbout verifies the view contains
// about information.
func TestHelpModelViewContainsAbout(t *testing.T) {
	m := NewHelpModel()
	view := m.View()

	if !strings.Contains(view, "About") {
		t.Error("view should contain 'About'")
	}
	if !strings.Contains(view, "github.com") {
		t.Error("view should contain GitHub URL")
	}
}

// TestHelpModelEscSendsBackToMenu verifies pressing esc produces
// a BackToMenuMsg.
func TestHelpModelEscSendsBackToMenu(t *testing.T) {
	m := NewHelpModel()

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

// TestHelpModelQSendsBackToMenu verifies pressing q produces
// a BackToMenuMsg.
func TestHelpModelQSendsBackToMenu(t *testing.T) {
	m := NewHelpModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q key should produce a command")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestHelpModelBackspaceSendsBackToMenu verifies pressing backspace
// produces a BackToMenuMsg.
func TestHelpModelBackspaceSendsBackToMenu(t *testing.T) {
	m := NewHelpModel()

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

// TestHelpModelWindowResize verifies WindowSizeMsg updates dimensions.
func TestHelpModelWindowResize(t *testing.T) {
	m := NewHelpModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions = %dx%d, want 80x24", m.width, m.height)
	}
}

// TestHelpModelUnhandledKey verifies unrecognized keys are ignored.
func TestHelpModelUnhandledKey(t *testing.T) {
	m := NewHelpModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("unrecognized key should not produce a command")
	}
}
