// menu.go implements the main application selection menu screen.
// It wraps a bubbles/list component populated from the installer
// registry, letting the user browse and select a Sonar application
// to install.
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// defaultMenuWidth is the fallback menu width before a WindowSizeMsg arrives.
const defaultMenuWidth = 60

// defaultMenuHeight is the fallback menu height before a WindowSizeMsg arrives.
const defaultMenuHeight = 20

// AppSelectedMsg is sent when the user selects an application from the menu.
type AppSelectedMsg struct {
	// AppID is the registry key of the selected application.
	AppID string
}

// menuItem adapts an installer entry into a bubbles/list item.
type menuItem struct {
	// appID is the registry key for looking up the installer.
	appID string
	// name is the display name of the application.
	name string
	// desc is the short description of the application.
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

// MenuModel is the Bubble Tea model for the main menu screen.
type MenuModel struct {
	// list is the underlying bubbles list component.
	list list.Model
	// registry holds the installer constructors for reference.
	registry installer.Registry
}

// NewMenuModel creates a new menu model populated from the given registry.
func NewMenuModel(reg installer.Registry) MenuModel {
	items := buildMenuItems(reg)
	delegate := newMenuDelegate()

	l := list.New(items, delegate, defaultMenuWidth, defaultMenuHeight)
	l.Title = "STUI — Sonar Terminal User Interface"
	l.Styles.Title = BannerStyle.
		Padding(0, 1).
		MarginBottom(1)
	l.SetShowStatusBar(true)
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

	return MenuModel{
		list:     l,
		registry: reg,
	}
}

// buildMenuItems converts registry entries into bubbles list items,
// preserving the display order defined by Registry.List().
func buildMenuItems(reg installer.Registry) []list.Item {
	ids := reg.List()
	items := make([]list.Item, 0, len(ids))
	for _, id := range ids {
		ctor, ok := reg[id]
		if !ok {
			continue
		}
		inst := ctor()
		items = append(items, menuItem{
			appID: id,
			name:  inst.Name(),
			desc:  inst.Description(),
		})
	}
	return items
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
			// Let the parent handle quit.
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

// selectCurrentItem returns a command that produces an AppSelectedMsg
// for the currently highlighted menu item.
func (m MenuModel) selectCurrentItem() tea.Cmd {
	item, ok := m.list.SelectedItem().(menuItem)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return AppSelectedMsg{AppID: item.appID}
	}
}

// View implements tea.Model. Renders the list.
func (m MenuModel) View() string {
	return AppStyle.Render(m.list.View())
}

// SelectedAppID returns the app ID of the currently highlighted item,
// or an empty string if nothing is selected.
func (m MenuModel) SelectedAppID() string {
	item, ok := m.list.SelectedItem().(menuItem)
	if !ok {
		return ""
	}
	return item.appID
}

// Ensure menuItem satisfies the list.DefaultItem interface at compile time.
var _ list.DefaultItem = menuItem{}
