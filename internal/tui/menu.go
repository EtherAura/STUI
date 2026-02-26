// menu.go implements the top-level category menu screen.
// This is the first screen the user sees, presenting three categories:
// Installers (Sonar app installation), Settings, and Help.
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// defaultMenuWidth is the fallback menu width before a WindowSizeMsg arrives.
const defaultMenuWidth = 60

// defaultMenuHeight is the fallback menu height before a WindowSizeMsg arrives.
const defaultMenuHeight = 20

// Category constants identify the top-level menu categories.
const (
	// CategoryInstallers opens the Sonar application installer list.
	CategoryInstallers = "installers"
	// CategorySettings opens the settings screen.
	CategorySettings = "settings"
	// CategoryHelp opens the help screen.
	CategoryHelp = "help"
)

// CategorySelectedMsg is sent when the user selects a top-level category.
type CategorySelectedMsg struct {
	// Category is the selected category identifier.
	Category string
}

// menuItem adapts a category entry into a bubbles/list item.
type menuItem struct {
	// category is the identifier for the category.
	category string
	// name is the display name of the category.
	name string
	// desc is a short description of the category.
	desc string
}

// FilterValue implements list.Item. Returns the name for filtering.
func (i menuItem) FilterValue() string { return i.name }

// Title implements list.DefaultItem. Returns the display name.
func (i menuItem) Title() string { return i.name }

// Description implements list.DefaultItem. Returns the short description.
func (i menuItem) Description() string { return i.desc }

// menuDelegate renders menu items with the shared style palette.
type menuDelegate struct {
	list.DefaultDelegate
}

// newMenuDelegate creates a delegate with STUI-themed colors.
func newMenuDelegate() menuDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(ColorPurple).
		BorderLeftForeground(ColorPurple)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(ColorDimWhite).
		BorderLeftForeground(ColorPurple)
	return menuDelegate{DefaultDelegate: d}
}

// MenuModel is the Bubble Tea model for the top-level category menu.
type MenuModel struct {
	// list is the underlying bubbles list component.
	list list.Model
}

// NewMenuModel creates a new top-level menu with the three categories.
func NewMenuModel() MenuModel {
	items := []list.Item{
		menuItem{
			category: CategoryInstallers,
			name:     "Installers",
			desc:     "Install and manage Sonar applications",
		},
		menuItem{
			category: CategorySettings,
			name:     "Settings",
			desc:     "Configure STUI preferences",
		},
		menuItem{
			category: CategoryHelp,
			name:     "Help",
			desc:     "Documentation and support information",
		},
	}

	delegate := newMenuDelegate()
	l := list.New(items, delegate, defaultMenuWidth, defaultMenuHeight)
	l.Title = "STUI — Sonar Terminal User Interface"
	l.Styles.Title = BannerStyle.
		Padding(0, 1).
		MarginBottom(1)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	// Override default help-key text with quit shortcut.
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "quit"),
			),
		}
	}

	return MenuModel{list: l}
}

// Init implements tea.Model. No initial command needed.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles key events and delegates
// to the underlying list component.
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			return m, m.selectCurrentItem()
		}
	case tea.WindowSizeMsg:
		// Reserve space for app-level chrome (padding).
		m.list.SetSize(msg.Width-4, msg.Height-4)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// selectCurrentItem returns a command that produces a CategorySelectedMsg
// for the currently highlighted menu item.
func (m MenuModel) selectCurrentItem() tea.Cmd {
	item, ok := m.list.SelectedItem().(menuItem)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return CategorySelectedMsg{Category: item.category}
	}
}

// View implements tea.Model. Renders the menu list.
func (m MenuModel) View() string {
	return AppStyle.Render(m.list.View())
}

// SelectedCategory returns the category of the currently highlighted item,
// or an empty string if nothing is selected.
func (m MenuModel) SelectedCategory() string {
	item, ok := m.list.SelectedItem().(menuItem)
	if !ok {
		return ""
	}
	return item.category
}

// Ensure menuItem satisfies the list.DefaultItem interface at compile time.
var _ list.DefaultItem = menuItem{}
