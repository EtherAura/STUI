package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewAppModel(t *testing.T) {
	m := NewAppModel()
	if m.Quitting() {
		t.Error("new model should not be quitting")
	}
	if m.Width() != 0 || m.Height() != 0 {
		t.Error("new model should have zero dimensions")
	}
}

func TestAppModelInit(t *testing.T) {
	m := NewAppModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestAppModelView(t *testing.T) {
	m := NewAppModel()
	view := m.View()

	if !strings.Contains(view, "STUI") {
		t.Error("view should contain 'STUI'")
	}
	if !strings.Contains(view, "Press q to quit") {
		t.Error("view should contain quit instruction")
	}
}

func TestAppModelViewQuitting(t *testing.T) {
	m := NewAppModel()

	// Simulate pressing 'q'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(AppModel)

	if !model.Quitting() {
		t.Error("model should be quitting after 'q'")
	}

	view := model.View()
	if !strings.Contains(view, "Goodbye") {
		t.Error("quitting view should contain 'Goodbye'")
	}
}

func TestAppModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
		quit bool
	}{
		{
			name: "q key quits",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			quit: true,
		},
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
			if !tt.quit && cmd != nil {
				t.Error("expected no command for non-quit key")
			}
		})
	}
}

func TestAppModelUpdateWindowSize(t *testing.T) {
	m := NewAppModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(AppModel)

	if cmd != nil {
		t.Error("WindowSizeMsg should not produce a command")
	}
	if model.Width() != 120 {
		t.Errorf("Width() = %d, want 120", model.Width())
	}
	if model.Height() != 40 {
		t.Errorf("Height() = %d, want 40", model.Height())
	}
}

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
