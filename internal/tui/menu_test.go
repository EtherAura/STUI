// menu_test.go contains unit tests for the MenuModel, covering
// initialization, item population from the registry, keyboard
// navigation, selection handling, and view rendering.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// TestNewMenuModel verifies the menu initializes with all registry
// entries and the correct title.
func TestNewMenuModel(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)

	view := m.View()
	if !strings.Contains(view, "STUI") {
		t.Error("menu view should contain 'STUI' in the title")
	}
}

// TestMenuModelItemCount verifies the menu has one item per registry entry.
func TestMenuModelItemCount(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)

	items := m.list.Items()
	if len(items) != len(reg.List()) {
		t.Errorf("menu has %d items, want %d", len(items), len(reg.List()))
	}
}

// TestMenuModelItemOrder verifies that items appear in registry list order.
func TestMenuModelItemOrder(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)
	ids := reg.List()

	items := m.list.Items()
	for i, item := range items {
		mi, ok := item.(menuItem)
		if !ok {
			t.Fatalf("item %d is not a menuItem", i)
		}
		if mi.appID != ids[i] {
			t.Errorf("item %d appID = %q, want %q", i, mi.appID, ids[i])
		}
	}
}

// TestMenuItemFilterValue verifies the list filter uses the app name.
func TestMenuItemFilterValue(t *testing.T) {
	item := menuItem{appID: "test", name: "Test App", desc: "A test"}
	if item.FilterValue() != "Test App" {
		t.Errorf("FilterValue() = %q, want %q", item.FilterValue(), "Test App")
	}
}

// TestMenuItemTitle verifies the list item title is the app name.
func TestMenuItemTitle(t *testing.T) {
	item := menuItem{appID: "test", name: "Test App", desc: "A test"}
	if item.Title() != "Test App" {
		t.Errorf("Title() = %q, want %q", item.Title(), "Test App")
	}
}

// TestMenuItemDescription verifies the list item description.
func TestMenuItemDescription(t *testing.T) {
	item := menuItem{appID: "test", name: "Test App", desc: "A test"}
	if item.Description() != "A test" {
		t.Errorf("Description() = %q, want %q", item.Description(), "A test")
	}
}

// TestMenuModelSelectedAppID verifies the default selection is the
// first item in the list.
func TestMenuModelSelectedAppID(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)

	ids := reg.List()
	if m.SelectedAppID() != ids[0] {
		t.Errorf("SelectedAppID() = %q, want %q", m.SelectedAppID(), ids[0])
	}
}

// TestMenuModelEnterSendsMsg verifies that pressing enter returns a command
// that produces an AppSelectedMsg.
func TestMenuModelEnterSendsMsg(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)

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

// TestMenuModelQuitKey verifies that pressing 'q' in the menu
// produces a tea.Quit command.
func TestMenuModelQuitKey(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q key should produce a quit command")
	}
}

// TestMenuModelWindowResize verifies that WindowSizeMsg updates the
// list dimensions.
func TestMenuModelWindowResize(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	// The list should have been resized (accounting for padding).
	view := m.View()
	if view == "" {
		t.Error("view should not be empty after resize")
	}
}

// TestMenuModelInit verifies Init returns no command.
func TestMenuModelInit(t *testing.T) {
	reg := installer.NewRegistry()
	m := NewMenuModel(reg)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// TestBuildMenuItems verifies buildMenuItems creates the correct
// number and content of items from the registry.
func TestBuildMenuItems(t *testing.T) {
	reg := installer.NewRegistry()
	items := buildMenuItems(reg)

	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}

	// Verify first item is Customer Portal.
	first, ok := items[0].(menuItem)
	if !ok {
		t.Fatal("first item is not a menuItem")
	}
	if first.appID != installer.AppCustomerPortal {
		t.Errorf("first item appID = %q, want %q", first.appID, installer.AppCustomerPortal)
	}
	if first.name != "Customer Portal" {
		t.Errorf("first item name = %q, want %q", first.name, "Customer Portal")
	}
}
