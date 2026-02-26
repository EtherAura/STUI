// detail_test.go contains unit tests for the DetailModel, covering
// initialization, view rendering, key handling for confirm and back,
// and edge cases with unknown app IDs.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// TestNewDetailModel verifies a detail model initializes with the
// correct app ID and a non-nil installer.
func TestNewDetailModel(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppCustomerPortal)

	if m.AppID() != installer.AppCustomerPortal {
		t.Errorf("AppID() = %q, want %q", m.AppID(), installer.AppCustomerPortal)
	}
	if m.installer == nil {
		t.Error("installer should not be nil for a known app")
	}
}

// TestNewDetailModelUnknown verifies a detail model for an unknown
// app ID has a nil installer.
func TestNewDetailModelUnknown(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, "unknown-app")

	if m.AppID() != "unknown-app" {
		t.Errorf("AppID() = %q, want %q", m.AppID(), "unknown-app")
	}
	if m.installer != nil {
		t.Error("installer should be nil for an unknown app")
	}
}

// TestDetailModelInit verifies Init returns no command.
func TestDetailModelInit(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppCustomerPortal)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// TestDetailModelViewContent verifies the view shows the app name,
// description, requirements, and install steps.
func TestDetailModelViewContent(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppCustomerPortal)
	view := m.View()

	if !strings.Contains(view, "Customer Portal") {
		t.Error("view should contain app name")
	}
	if !strings.Contains(view, "Requirements") {
		t.Error("view should contain Requirements heading")
	}
	if !strings.Contains(view, "docker") {
		t.Error("view should list docker as requirement for portal")
	}
	if !strings.Contains(view, "Install prerequisites") {
		t.Error("view should contain first step name")
	}
	if !strings.Contains(view, "enter") {
		t.Error("view should contain enter prompt")
	}
	if !strings.Contains(view, "esc") {
		t.Error("view should contain esc prompt")
	}
}

// TestDetailModelViewUnknown verifies the error view for unknown apps.
func TestDetailModelViewUnknown(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, "bad-id")
	view := m.View()

	if !strings.Contains(view, "bad-id") {
		t.Error("error view should contain the bad app ID")
	}
	if !strings.Contains(view, "esc") {
		t.Error("error view should mention esc to go back")
	}
}

// TestDetailModelEnterKey verifies pressing enter produces a StartPreflightMsg.
func TestDetailModelEnterKey(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppPoller)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter key should produce a command")
	}

	msg := cmd()
	preflight, ok := msg.(StartPreflightMsg)
	if !ok {
		t.Fatalf("command should produce StartPreflightMsg, got %T", msg)
	}
	if preflight.AppID != installer.AppPoller {
		t.Errorf("StartPreflightMsg.AppID = %q, want %q", preflight.AppID, installer.AppPoller)
	}
}

// TestDetailModelEscKey verifies pressing esc produces a BackToMenuMsg.
func TestDetailModelEscKey(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppCustomerPortal)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc key should produce a command")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("command should produce BackToMenuMsg, got %T", msg)
	}
}

// TestDetailModelBackspaceKey verifies pressing backspace also goes back.
func TestDetailModelBackspaceKey(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppCustomerPortal)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd == nil {
		t.Fatal("backspace key should produce a command")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("command should produce BackToMenuMsg, got %T", msg)
	}
}

// TestDetailModelWindowResize verifies WindowSizeMsg updates dimensions.
func TestDetailModelWindowResize(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewDetailModel(reg, installer.AppCustomerPortal)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m.width != 100 || m.height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", m.width, m.height)
	}
}

// TestDetailModelAllApps verifies detail views render for every
// registered application without panicking.
func TestDetailModelAllApps(t *testing.T) {
	reg := installer.NewRegistry()
	for _, appID := range reg.List() {
		t.Run(appID, func(t *testing.T) {
			m := NewDetailModel(reg, appID)
			view := m.View()
			if view == "" {
				t.Error("view should not be empty")
			}
		})
	}
}

// TestAppModelTransitionToDetail verifies that AppSelectedMsg
// transitions from menu to detail screen.
func TestAppModelTransitionToDetail(t *testing.T) {
	m := NewAppModel()

	updated, _ := m.Update(AppSelectedMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)

	if model.Screen() != ScreenDetail {
		t.Errorf("screen = %d, want ScreenDetail (%d)", model.Screen(), ScreenDetail)
	}
}

// TestAppModelTransitionBackToMenu verifies that BackToMenuMsg
// returns from detail to the installer list screen.
func TestAppModelTransitionBackToMenu(t *testing.T) {
	m := NewAppModel()

	// Go to detail first.
	updated, _ := m.Update(AppSelectedMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)

	// Then go back.
	updated, _ = model.Update(BackToMenuMsg{})
	model = updated.(AppModel)

	if model.Screen() != ScreenInstallers {
		t.Errorf("screen = %d, want ScreenInstallers (%d)", model.Screen(), ScreenInstallers)
	}
}
