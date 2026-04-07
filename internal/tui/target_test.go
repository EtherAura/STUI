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
	if m.mode != installer.TargetModeLocal {
		t.Fatalf("default mode = %q, want %q", m.mode, installer.TargetModeLocal)
	}
}

func TestTargetModelSubmitLocal(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a command on Proceed for local target")
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

func TestTargetModelToggleMode(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.mode != installer.TargetModeLocal {
		t.Fatalf("mode should stay local when selecting Local, got %q", m.mode)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != installer.TargetModeSSH {
		t.Fatalf("mode = %q, want %q", m.mode, installer.TargetModeSSH)
	}
	if m.FocusIndex() != 2 {
		t.Fatalf("FocusIndex() = %d, want 2", m.FocusIndex())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != installer.TargetModeLocal {
		t.Fatalf("mode = %q, want %q", m.mode, installer.TargetModeLocal)
	}
}

func TestTargetModelSubmitSSHRequiresFields(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)
	m.mode = installer.TargetModeSSH

	_, err := m.buildTarget()
	if err == nil {
		t.Fatal("expected validation error for incomplete ssh target")
	}
}

func TestTargetModelProceedRowForSSH(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)
	m.mode = installer.TargetModeSSH

	for i := 0; i < 8; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	if !m.isProceedRow() {
		t.Fatalf("expected focus on Proceed row, got index %d", m.FocusIndex())
	}
}

func TestTargetModelBuildTargetIncludesSSHAuthFields(t *testing.T) {
	m := NewTargetModel(installer.AppPoller)
	m.mode = installer.TargetModeSSH
	m.inputs[1].SetValue("ubuntu")
	m.inputs[2].SetValue("192.168.0.132")
	m.inputs[3].SetValue("22")
	m.inputs[4].SetValue("secret")
	m.inputs[5].SetValue("~/.ssh/id_ed25519")

	target, err := m.buildTarget()
	if err != nil {
		t.Fatalf("buildTarget() error = %v", err)
	}
	if target.Password != "secret" {
		t.Fatalf("Password = %q, want %q", target.Password, "secret")
	}
	if target.KeyPath != "~/.ssh/id_ed25519" {
		t.Fatalf("KeyPath = %q, want %q", target.KeyPath, "~/.ssh/id_ed25519")
	}
}
