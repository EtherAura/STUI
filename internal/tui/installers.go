// installers.go implements the installer selection sub-menu.
// After selecting the "Installers" category from the main menu,
// this screen lists all available Sonar application installers
// using a bubbles/list component populated from the registry.
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// defaultInstallerListWidth is the fallback width before a WindowSizeMsg arrives.
const defaultInstallerListWidth = 60

// defaultInstallerListHeight is the fallback height before a WindowSizeMsg arrives.
const defaultInstallerListHeight = 20

// AppSelectedMsg is sent when the user selects an application to install.
type AppSelectedMsg struct {
	// AppID is the registry key of the selected application.
	AppID string
}

// installerItem adapts an installer entry into a bubbles/list item.
type installerItem struct {
	// appID is the registry key for looking up the installer.
	appID string
	// name is the display name of the application.
	name string
	// desc is the short description of the application.
	desc string
}

// FilterValue implements list.Item. Returns the name for filtering.
func (i installerItem) FilterValue() string { return i.name }

// Title implements list.DefaultItem. Returns the display name.
func (i installerItem) Title() string { return i.name }

// Description implements list.DefaultItem. Returns the short description.
func (i installerItem) Description() string { return i.desc }

// installerDelegate renders installer items with the shared style palette.
type installerDelegate struct {
	list.DefaultDelegate
}

// newInstallerDelegate creates a delegate with STUI-themed colors.
func newInstallerDelegate() installerDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(ColorPurple).
		BorderLeftForeground(ColorPurple)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(ColorDimWhite).
		BorderLeftForeground(ColorPurple)
	return installerDelegate{DefaultDelegate: d}
}

// InstallerListModel is the Bubble Tea model for the installer
// selection sub-menu.
type InstallerListModel struct {
	// list is the underlying bubbles list component.
	list list.Model
	// registry holds the installer constructors for reference.
	registry installer.Registry
}

// NewInstallerListModel creates a new installer list model populated
// from the given registry.
func NewInstallerListModel(reg installer.Registry) InstallerListModel {
	items := buildInstallerItems(reg)
	delegate := newInstallerDelegate()

	l := list.New(items, delegate, defaultInstallerListWidth, defaultInstallerListHeight)
	l.Title = "Select Application"
	l.Styles.Title = BannerStyle.
		Padding(0, 1).
		MarginBottom(1)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		}
	}

	return InstallerListModel{
		list:     l,
		registry: reg,
	}
}

// buildInstallerItems converts registry entries into bubbles list items,
// preserving the display order defined by Registry.List().
func buildInstallerItems(reg installer.Registry) []list.Item {
	ids := reg.List()
	items := make([]list.Item, 0, len(ids))
	for _, id := range ids {
		ctor, ok := reg[id]
		if !ok {
			continue
		}
		inst := ctor()
		items = append(items, installerItem{
			appID: id,
			name:  inst.Name(),
			desc:  inst.Description(),
		})
	}
	return items
}

// Init implements tea.Model. No initial command needed.
func (m InstallerListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles key events and delegates to
// the underlying list component.
func (m InstallerListModel) Update(msg tea.Msg) (InstallerListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.selectCurrentItem()
		case "esc", "backspace":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
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
// for the currently highlighted installer item.
func (m InstallerListModel) selectCurrentItem() tea.Cmd {
	item, ok := m.list.SelectedItem().(installerItem)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return AppSelectedMsg{AppID: item.appID}
	}
}

// View implements tea.Model. Renders the installer list.
func (m InstallerListModel) View() string {
	return AppStyle.Render(m.list.View())
}

// SelectedAppID returns the app ID of the currently highlighted item,
// or an empty string if nothing is selected.
func (m InstallerListModel) SelectedAppID() string {
	item, ok := m.list.SelectedItem().(installerItem)
	if !ok {
		return ""
	}
	return item.appID
}

// Ensure installerItem satisfies the list.DefaultItem interface at compile time.
var _ list.DefaultItem = installerItem{}
