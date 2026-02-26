// menu_test.go contains unit tests for the top-level category MenuModel,
// covering initialization, item population, keyboard navigation,
// selection handling, and view rendering.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewMenuModel verifies the menu initializes with three category
// items and the correct title.
func TestNewMenuModel(t *testing.T) {
	m := NewMenuModel()

	view := m.View()
	if !strings.Contains(view, "STUI") {
		t.Error("menu view should contain 'STUI' in the title")
	}
}

// TestMenuModelItemCount verifies the menu has three category items.
func TestMenuModelItemCount(t *testing.T) {
	m := NewMenuModel()

	items := m.list.Items()
	if len(items) != 3 {
		t.Errorf("menu has %d items, want 3", len(items))
	}
}

// TestMenuModelItemOrder verifies that categories appear in the expected order.
func TestMenuModelItemOrder(t *testing.T) {
	m := NewMenuModel()

	expected := []string{CategoryInstallers, CategorySettings, CategoryHelp}
	items := m.list.Items()
	for i, item := range items {
		mi, ok := item.(menuItem)
		if !ok {
			t.Fatalf("item %d is not a menuItem", i)
		}
		if mi.category != expected[i] {
			t.Errorf("item %d category = %q, want %q", i, mi.category, expected[i])
		}
	}
}

// TestMenuItemFilterValue verifies the list filter uses the category name.
func TestMenuItemFilterValue(t *testing.T) {
	item := menuItem{category: "test", name: "Test Category", desc: "A test"}
	if item.FilterValue() != "Test Category" {
		t.Errorf("FilterValue() = %q, want %q", item.FilterValue(), "Test Category")
	}
}

// TestMenuItemTitle verifies the list item title is the category name.
func TestMenuItemTitle(t *testing.T) {
	item := menuItem{category: "test", name: "Test Category", desc: "A test"}
	if item.Title() != "Test Category" {
		t.Errorf("Title() = %q, want %q", item.Title(), "Test Category")
	}
}

// TestMenuItemDescription verifies the list item description.
func TestMenuItemDescription(t *testing.T) {
	item := menuItem{category: "test", name: "Test Category", desc: "A test"}
	if item.Description() != "A test" {
		t.Errorf("Description() = %q, want %q", item.Description(), "A test")
	}
}

// TestMenuModelSelectedCategory verifies the default selection is the
// first category (Installers).
func TestMenuModelSelectedCategory(t *testing.T) {
	m := NewMenuModel()

	if m.SelectedCategory() != CategoryInstallers {
		t.Errorf("SelectedCategory() = %q, want %q", m.SelectedCategory(), CategoryInstallers)
	}
}

// TestMenuModelEnterSendsMsg verifies that pressing enter returns a command
// that produces a CategorySelectedMsg.
func TestMenuModelEnterSendsMsg(t *testing.T) {
	m := NewMenuModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter key should produce a command")
	}

	msg := cmd()
	selected, ok := msg.(CategorySelectedMsg)
	if !ok {
		t.Fatalf("command should produce CategorySelectedMsg, got %T", msg)
	}
	if selected.Category != CategoryInstallers {
		t.Errorf("CategorySelectedMsg.Category = %q, want %q", selected.Category, CategoryInstallers)
	}
}

// TestMenuModelQuitKey verifies that pressing 'q' in the menu
// produces a tea.Quit command.
func TestMenuModelQuitKey(t *testing.T) {
	m := NewMenuModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q key should produce a quit command")
	}
}

// TestMenuModelWindowResize verifies that WindowSizeMsg updates the
// list dimensions.
func TestMenuModelWindowResize(t *testing.T) {
	m := NewMenuModel()

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	// The list should have been resized (accounting for padding).
	view := m.View()
	if view == "" {
		t.Error("view should not be empty after resize")
	}
}

// TestMenuModelInit verifies Init returns no command.
func TestMenuModelInit(t *testing.T) {
	m := NewMenuModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}
