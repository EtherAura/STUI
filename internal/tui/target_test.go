package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

func TestNewTargetModel(t *testing.T) {
	m := NewTargetModel(installer.AppCustomerPortal)

	if m.AppID() != installer.AppCustomerPortal {
		t.Fatalf("AppID() = %q, want %q", m.AppID(), installer.AppCustomerPortal)
	}
	if m.FocusIndex() != 0 {
		t.Fatalf("FocusIndex() = %d, want 0", m.FocusIndex())
	}
	if got := m.inputs[0].Value(); got != string(installer.TargetModeLocal) {
		t.Fatalf("default mode = %q, want %q", got, installer.TargetModeLocal)
	}
}

func TestTargetModelSubmitLocal(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a command for local target")
	}

	msg := cmd()
	preflight, ok := msg.(StartPreflightMsg)
	if !ok {
		t.Fatalf("expected StartPreflightMsg, got %T", msg)
	}
	if preflight.Target.Mode != installer.TargetModeLocal {
		t.Fatalf("target mode = %q, want %q", preflight.Target.Mode, installer.TargetModeLocal)
	}
}

func TestTargetModelSubmitSSHRequiresFields(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)
	m.inputs[0].SetValue(string(installer.TargetModeSSH))

	_, err := m.buildTarget()
	if err == nil {
		t.Fatal("expected validation error for incomplete ssh target")
	}
}
