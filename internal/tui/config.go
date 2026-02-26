// config.go implements the interactive configuration wizard screen.
// After preflight checks pass, this screen presents a series of
// text input prompts for the configuration values required by the
// selected application's installer.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// ConfigDoneMsg signals that configuration is complete and carries
// the filled Config ready for installation.
type ConfigDoneMsg struct {
	// AppID is the registry key of the selected application.
	AppID string
	// Config holds the user-provided configuration values.
	Config *installer.Config
}

// configField describes a single configuration input field.
type configField struct {
	// key identifies which Config struct field this maps to.
	key string
	// label is the user-visible prompt text.
	label string
	// placeholder shows example text in the empty input.
	placeholder string
	// required marks the field as mandatory.
	required bool
}

// appConfigFields maps each app ID to its required/optional config fields.
// This centralizes field definitions for the wizard.
var appConfigFields = map[string][]configField{
	installer.AppCustomerPortal: {
		{key: "sonar_url", label: "Sonar Instance URL", placeholder: "https://myisp.sonar.software", required: true},
		{key: "api_username", label: "API Username", placeholder: "admin", required: true},
		{key: "api_password", label: "API Password", placeholder: "password", required: true},
		{key: "domain", label: "Portal Domain", placeholder: "portal.example.com", required: true},
		{key: "email", label: "Admin Email (for TLS)", placeholder: "admin@example.com", required: true},
	},
	installer.AppNetflowOnPrem: {
		{key: "sonar_url", label: "Sonar Instance URL", placeholder: "https://myisp.sonar.software", required: true},
		{key: "api_token", label: "API Token", placeholder: "bearer token", required: true},
		{key: "netflow_name", label: "Collector Name", placeholder: "netflow-collector-1", required: false},
		{key: "public_ip", label: "Public IP Address", placeholder: "203.0.113.1", required: true},
		{key: "db_password", label: "Database Password", placeholder: "secure-password", required: false},
	},
	installer.AppFreeRADIUS: {
		{key: "sonar_url", label: "Sonar Instance URL", placeholder: "https://myisp.sonar.software", required: true},
	},
	installer.AppPoller: {
		{key: "sonar_url", label: "Sonar Instance URL", placeholder: "https://myisp.sonar.software", required: true},
		{key: "poller_api_key", label: "Poller API Key", placeholder: "sonar-api-key", required: true},
	},
}

// ConfigModel is the Bubble Tea model for the config wizard screen.
type ConfigModel struct {
	// appID is the registry key of the selected application.
	appID string
	// fields defines what config values to collect.
	fields []configField
	// inputs holds the text input models for each field.
	inputs []textinput.Model
	// focusIndex tracks which input is currently focused.
	focusIndex int
	// err holds any validation error to display.
	err string
	// width and height track the terminal dimensions.
	width  int
	height int
}

// NewConfigModel creates a config wizard for the given app ID.
func NewConfigModel(appID string) ConfigModel {
	fields, ok := appConfigFields[appID]
	if !ok {
		fields = []configField{
			{key: "sonar_url", label: "Sonar Instance URL", placeholder: "https://myisp.sonar.software", required: true},
		}
	}

	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.placeholder
		ti.CharLimit = 256
		ti.Width = 40
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	return ConfigModel{
		appID:  appID,
		fields: fields,
		inputs: inputs,
	}
}

// Init implements tea.Model. Returns the blink command for the
// focused text input cursor.
func (m ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model. Handles tab/shift-tab for field
// navigation, enter for submission, and esc for going back.
func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		case "tab", "down":
			return m.nextField()
		case "shift+tab", "up":
			return m.prevField()
		case "enter":
			// If on the last field, try to submit.
			if m.focusIndex == len(m.inputs)-1 {
				return m.submit()
			}
			// Otherwise advance to next field.
			return m.nextField()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update the focused input.
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

// nextField advances focus to the next input field.
func (m ConfigModel) nextField() (ConfigModel, tea.Cmd) {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	return m, m.inputs[m.focusIndex].Focus()
}

// prevField moves focus to the previous input field.
func (m ConfigModel) prevField() (ConfigModel, tea.Cmd) {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
	return m, m.inputs[m.focusIndex].Focus()
}

// submit validates the inputs and produces a ConfigDoneMsg if valid.
func (m ConfigModel) submit() (ConfigModel, tea.Cmd) {
	// Validate required fields.
	for i, f := range m.fields {
		if f.required && strings.TrimSpace(m.inputs[i].Value()) == "" {
			m.err = fmt.Sprintf("%s is required", f.label)
			// Focus the offending field.
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = i
			return m, m.inputs[i].Focus()
		}
	}

	cfg := m.buildConfig()
	return m, func() tea.Msg {
		return ConfigDoneMsg{AppID: m.appID, Config: cfg}
	}
}

// buildConfig constructs an installer.Config from the input values.
func (m ConfigModel) buildConfig() *installer.Config {
	cfg := &installer.Config{}
	for i, f := range m.fields {
		val := strings.TrimSpace(m.inputs[i].Value())
		switch f.key {
		case "sonar_url":
			cfg.SonarURL = val
		case "api_token":
			cfg.APIToken = val
		case "api_username":
			cfg.APIUsername = val
		case "api_password":
			cfg.APIPassword = val
		case "domain":
			cfg.Domain = val
		case "email":
			cfg.Email = val
		case "netflow_name":
			cfg.NetflowName = val
		case "public_ip":
			cfg.PublicIP = val
		case "db_password":
			cfg.DBPassword = val
		case "poller_api_key":
			cfg.PollerAPIKey = val
		}
	}
	return cfg
}

// View implements tea.Model. Renders the input fields with labels.
func (m ConfigModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configuration"))
	b.WriteString("\n")
	b.WriteString(DimStyle.Render("Fill in the values below. Tab/Shift+Tab to navigate."))
	b.WriteString("\n\n")

	for i, f := range m.fields {
		// Label with required indicator.
		label := f.label
		if f.required {
			label += " *"
		}

		if i == m.focusIndex {
			b.WriteString(BannerStyle.Render(label))
		} else {
			b.WriteString(BodyStyle.Render(label))
		}
		b.WriteString("\n")
		b.WriteString("  " + m.inputs[i].View())
		b.WriteString("\n\n")
	}

	// Validation error.
	if m.err != "" {
		b.WriteString(ErrorStyle.Render("✗ " + m.err))
		b.WriteString("\n\n")
	}

	b.WriteString(HelpStyle.Render("enter: submit • tab/shift+tab: navigate • esc: back"))

	return AppStyle.Render(b.String())
}

// AppID returns the registry key of the application being configured.
func (m ConfigModel) AppID() string {
	return m.appID
}

// FocusIndex returns the index of the currently focused field.
func (m ConfigModel) FocusIndex() int {
	return m.focusIndex
}
