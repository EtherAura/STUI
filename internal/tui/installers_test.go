// installers_test.go contains unit tests for the InstallerListModel,
// covering initialization, item population from the registry, keyboard
// navigation, selection handling, back navigation, and view rendering.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// TestNewInstallerListModel verifies the installer list initializes
// with all registry entries and the correct title.
func TestNewInstallerListModel(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

	view := m.View()
	if !strings.Contains(view, "Select Application") {
		t.Error("installer list view should contain 'Select Application' in the title")
	}
}

// TestInstallerListItemCount verifies the list has one item per
// registry entry.
func TestInstallerListItemCount(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

	items := m.list.Items()
	if len(items) != len(reg.List()) {
		t.Errorf("installer list has %d items, want %d", len(items), len(reg.List()))
	}
}

// TestInstallerListItemOrder verifies that items appear in registry
// list order.
func TestInstallerListItemOrder(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)
	ids := reg.List()

	items := m.list.Items()
	for i, item := range items {
		ii, ok := item.(installerItem)
		if !ok {
			t.Fatalf("item %d is not an installerItem", i)
		}
		if ii.appID != ids[i] {
			t.Errorf("item %d appID = %q, want %q", i, ii.appID, ids[i])
		}
	}
}

// TestInstallerItemFilterValue verifies the list filter uses the app name.
func TestInstallerItemFilterValue(t *testing.T) {
	item := installerItem{appID: "test", name: "Test App", desc: "A test"}
	if item.FilterValue() != "Test App" {
		t.Errorf("FilterValue() = %q, want %q", item.FilterValue(), "Test App")
	}
}

// TestInstallerItemTitle verifies the list item title is the app name.
func TestInstallerItemTitle(t *testing.T) {
	item := installerItem{appID: "test", name: "Test App", desc: "A test"}
	if item.Title() != "Test App" {
		t.Errorf("Title() = %q, want %q", item.Title(), "Test App")
	}
}

// TestInstallerItemDescription verifies the list item description.
func TestInstallerItemDescription(t *testing.T) {
	item := installerItem{appID: "test", name: "Test App", desc: "A test"}
	if item.Description() != "A test" {
		t.Errorf("Description() = %q, want %q", item.Description(), "A test")
	}
}

// TestInstallerListSelectedAppID verifies the default selection is the
// first item in the list.
func TestInstallerListSelectedAppID(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

	ids := reg.List()
	if m.SelectedAppID() != ids[0] {
		t.Errorf("SelectedAppID() = %q, want %q", m.SelectedAppID(), ids[0])
	}
}

// TestInstallerListEnterSendsMsg verifies that pressing enter returns
// a command that produces an AppSelectedMsg.
func TestInstallerListEnterSendsMsg(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter key should produce a command")
	}

	msg := cmd()
	selected, ok := msg.(AppSelectedMsg)
	if !ok {
		t.Fatalf("command should produce AppSelectedMsg, got %T", msg)
	}
	if selected.AppID != reg.List()[0] {
		t.Errorf("AppSelectedMsg.AppID = %q, want %q", selected.AppID, reg.List()[0])
	}
}

// TestInstallerListEscSendsBackToMenu verifies that pressing esc
// produces a BackToMenuMsg.
func TestInstallerListEscSendsBackToMenu(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

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

// TestInstallerListBackspaceSendsBackToMenu verifies that pressing
// backspace also produces a BackToMenuMsg.
func TestInstallerListBackspaceSendsBackToMenu(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

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

// TestInstallerListWindowResize verifies that WindowSizeMsg updates
// the list dimensions.
func TestInstallerListWindowResize(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	view := m.View()
	if view == "" {
		t.Error("view should not be empty after resize")
	}
}

// TestInstallerListInit verifies Init returns no command.
func TestInstallerListInit(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewInstallerListModel(reg)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// TestBuildInstallerItems verifies buildInstallerItems creates the
// correct number and content of items from the registry.
func TestBuildInstallerItems(t *testing.T) {
	reg := installer.NewRegistry()
	items := buildInstallerItems(reg)

	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}

	// Verify first item is Customer Portal.
	first, ok := items[0].(installerItem)
	if !ok {
		t.Fatal("first item is not an installerItem")
	}
	if first.appID != installer.AppCustomerPortal {
		t.Errorf("first item appID = %q, want %q", first.appID, installer.AppCustomerPortal)
	}
	if first.name != "Customer Portal" {
		t.Errorf("first item name = %q, want %q", first.name, "Customer Portal")
	}
}
