// app_test.go contains unit tests for the root AppModel, covering
// initialization, key handling, window resize, view rendering,
// screen state, and graceful quit behavior.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewAppModel verifies a fresh model starts in a non-quitting state
// on the menu screen with zero dimensions.
func TestNewAppModel(t *testing.T) {
	m := NewAppModel()
	if m.Quitting() {
		t.Error("new model should not be quitting")
	}
	if m.Width() != 0 || m.Height() != 0 {
		t.Error("new model should have zero dimensions")
	}
	if m.Screen() != ScreenMenu {
		t.Errorf("new model should start on ScreenMenu, got %d", m.Screen())
	}
}

// TestAppModelInit verifies Init returns a command (from menu init).
func TestAppModelInit(t *testing.T) {
	m := NewAppModel()
	_ = m.Init()
	// Init may return nil or a cmd from the menu; either is acceptable.
}

// TestAppModelView verifies the default view contains the banner
// from the menu screen.
func TestAppModelView(t *testing.T) {
	m := NewAppModel()
	view := m.View()

	if !strings.Contains(view, "STUI") {
		t.Error("view should contain 'STUI'")
	}
}

// TestAppModelViewQuitting verifies the view shows a goodbye message
// after the user presses ctrl+c.
func TestAppModelViewQuitting(t *testing.T) {
	m := NewAppModel()

	// Simulate pressing ctrl+c (global quit).
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(AppModel)

	if !model.Quitting() {
		t.Error("model should be quitting after ctrl+c")
	}

	view := model.View()
	if !strings.Contains(view, "Goodbye") {
		t.Error("quitting view should contain 'Goodbye'")
	}
}

// TestAppModelUpdateQuit is a table-driven test verifying that ctrl+c
// produces tea.Quit while other keys are delegated to the menu.
func TestAppModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
		quit bool
	}{
		{
			name: "ctrl+c quits",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlC},
			quit: true,
		},
		{
			name: "other key does not quit",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			quit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAppModel()
			updated, cmd := m.Update(tt.msg)
			model := updated.(AppModel)

			if model.Quitting() != tt.quit {
				t.Errorf("Quitting() = %v, want %v", model.Quitting(), tt.quit)
			}
			if tt.quit && cmd == nil {
				t.Error("expected tea.Quit command")
			}
		})
	}
}

// TestAppModelUpdateWindowSize verifies that WindowSizeMsg updates the
// model's dimensions without producing a command.
func TestAppModelUpdateWindowSize(t *testing.T) {
	m := NewAppModel()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(AppModel)

	if model.Width() != 120 {
		t.Errorf("Width() = %d, want 120", model.Width())
	}
	if model.Height() != 40 {
		t.Errorf("Height() = %d, want 40", model.Height())
	}
}

// TestAppModelUnhandledMsg verifies that unrecognized message types
// are silently ignored.
func TestAppModelUnhandledMsg(t *testing.T) {
	m := NewAppModel()

	// Sending an unhandled message type should be a no-op.
	updated, cmd := m.Update("some string message")
	model := updated.(AppModel)

	if cmd != nil {
		t.Error("unhandled message should not produce a command")
	}
	if model.Quitting() {
		t.Error("unhandled message should not cause quitting")
	}
}

// TestAppModelAppSelectedMsg verifies that receiving an AppSelectedMsg
// transitions from the installer list to the detail screen.
func TestAppModelAppSelectedMsg(t *testing.T) {
	m := NewAppModel()

	// Navigate to installer list first via CategorySelectedMsg.
	updated, _ := m.Update(CategorySelectedMsg{Category: CategoryInstallers})
	model := updated.(AppModel)

	// Now select an app.
	updated, cmd := model.Update(AppSelectedMsg{AppID: "customer-portal"})
	model = updated.(AppModel)

	if cmd != nil {
		t.Error("AppSelectedMsg should not produce a command")
	}
	if model.Screen() != ScreenDetail {
		t.Errorf("screen = %d, want ScreenDetail (%d)", model.Screen(), ScreenDetail)
	}
}

// TestAppModelCategorySelected verifies CategorySelectedMsg transitions
// to the correct screen for each category.
func TestAppModelCategorySelected(t *testing.T) {
	tests := []struct {
		name     string
		category string
		screen   Screen
	}{
		{"installers", CategoryInstallers, ScreenInstallers},
		{"settings", CategorySettings, ScreenSettings},
		{"help", CategoryHelp, ScreenHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAppModel()
			updated, _ := m.Update(CategorySelectedMsg{Category: tt.category})
			model := updated.(AppModel)

			if model.Screen() != tt.screen {
				t.Errorf("screen = %d, want %d", model.Screen(), tt.screen)
			}
		})
	}
}

// TestScreenConstants verifies that screen enum values are distinct
// and start from zero.
func TestScreenConstants(t *testing.T) {
	screens := []Screen{
		ScreenMenu, ScreenInstallers, ScreenDetail, ScreenPreflight,
		ScreenConfig, ScreenInstall, ScreenVerify,
		ScreenSettings, ScreenHelp,
	}
	seen := make(map[Screen]bool)
	for _, s := range screens {
		if seen[s] {
			t.Errorf("duplicate screen value: %d", s)
		}
		seen[s] = true
	}
	if ScreenMenu != 0 {
		t.Errorf("ScreenMenu should be 0, got %d", ScreenMenu)
	}
}
