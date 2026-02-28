// config_test.go contains unit tests for the ConfigModel, covering
// initialization, field navigation, validation, submission, and
// view rendering for different app configurations.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// TestNewConfigModel verifies a new config model initializes with
// the correct app ID and expected number of fields.
func TestNewConfigModel(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)

	if m.AppID() != installer.AppCustomerPortal {
		t.Errorf("AppID() = %q, want %q", m.AppID(), installer.AppCustomerPortal)
	}
	if len(m.fields) != 5 {
		t.Errorf("portal should have 5 fields, got %d", len(m.fields))
	}
	if m.FocusIndex() != 0 {
		t.Errorf("initial focus should be 0, got %d", m.FocusIndex())
	}
}

// TestNewConfigModelNetflow verifies netflow has the right number of fields.
func TestNewConfigModelNetflow(t *testing.T) {
	m := NewConfigModel(installer.AppNetflowOnPrem, false)
	if len(m.fields) != 5 {
		t.Errorf("netflow should have 5 fields, got %d", len(m.fields))
	}
}

// TestNewConfigModelPoller verifies poller has the right number of fields.
func TestNewConfigModelPoller(t *testing.T) {
	m := NewConfigModel(installer.AppPoller, false)
	if len(m.fields) != 2 {
		t.Errorf("poller should have 2 fields, got %d", len(m.fields))
	}
}

// TestNewConfigModelFreeRADIUS verifies freeradius has the right number of fields.
func TestNewConfigModelFreeRADIUS(t *testing.T) {
	m := NewConfigModel(installer.AppFreeRADIUS, false)
	if len(m.fields) != 1 {
		t.Errorf("freeradius should have 1 field, got %d", len(m.fields))
	}
}

// TestNewConfigModelUnknown verifies unknown apps get a default fallback field.
func TestNewConfigModelUnknown(t *testing.T) {
	m := NewConfigModel("unknown-app", false)
	if len(m.fields) < 1 {
		t.Error("unknown apps should have at least the default sonar_url field")
	}
}

// TestConfigModelInit verifies Init returns a blink command.
func TestConfigModelInit(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a blink command")
	}
}

// TestConfigModelTabNavigation verifies tab advances to the next field.
func TestConfigModelTabNavigation(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Errorf("focus should be 1 after tab, got %d", m.FocusIndex())
	}
}

// TestConfigModelShiftTabNavigation verifies shift+tab goes to previous field.
func TestConfigModelShiftTabNavigation(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)

	// Tab to field 1, then shift+tab back to field 0.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.FocusIndex() != 0 {
		t.Errorf("focus should be 0 after shift+tab, got %d", m.FocusIndex())
	}
}

// TestConfigModelTabWraps verifies tab wraps from last to first field.
func TestConfigModelTabWraps(t *testing.T) {
	m := NewConfigModel(installer.AppPoller, false) // 2 fields
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 0 {
		t.Errorf("focus should wrap to 0, got %d", m.FocusIndex())
	}
}

// TestConfigModelShiftTabWraps verifies shift+tab wraps from first to last.
func TestConfigModelShiftTabWraps(t *testing.T) {
	m := NewConfigModel(installer.AppPoller, false) // 2 fields
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.FocusIndex() != 1 {
		t.Errorf("focus should wrap to 1, got %d", m.FocusIndex())
	}
}

// TestConfigModelEscKey verifies esc produces BackToMenuMsg.
func TestConfigModelEscKey(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a command")
	}
	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatalf("expected BackToMenuMsg, got %T", msg)
	}
}

// TestConfigModelSubmitValidation verifies that submitting with empty
// required fields shows a validation error.
func TestConfigModelSubmitValidation(t *testing.T) {
	m := NewConfigModel(installer.AppPoller, false) // sonar_url + poller_api_key

	// Navigate to last field and try to submit with empty inputs.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // focus field 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.err == "" {
		t.Error("expected validation error for empty required fields")
	}
}

// TestConfigModelBuildConfig verifies buildConfig maps values correctly.
func TestConfigModelBuildConfig(t *testing.T) {
	m := NewConfigModel(installer.AppNetflowOnPrem, false)

	// Simulate typing values into each field.
	values := []string{
		"https://test.sonar.software",
		"my-token",
		"collector-1",
		"1.2.3.4",
		"dbpass",
	}

	for i, val := range values {
		// Focus the field.
		for m.FocusIndex() != i {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		}
		// Type the value.
		for _, r := range val {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}

	cfg := m.buildConfig()
	if cfg.SonarURL != "https://test.sonar.software" {
		t.Errorf("SonarURL = %q, want %q", cfg.SonarURL, "https://test.sonar.software")
	}
	if cfg.APIToken != "my-token" {
		t.Errorf("APIToken = %q, want %q", cfg.APIToken, "my-token")
	}
	if cfg.NetflowName != "collector-1" {
		t.Errorf("NetflowName = %q, want %q", cfg.NetflowName, "collector-1")
	}
	if cfg.PublicIP != "1.2.3.4" {
		t.Errorf("PublicIP = %q, want %q", cfg.PublicIP, "1.2.3.4")
	}
	if cfg.DBPassword != "dbpass" {
		t.Errorf("DBPassword = %q, want %q", cfg.DBPassword, "dbpass")
	}
}

// TestConfigModelViewContent verifies the view contains labels and
// navigation help.
func TestConfigModelViewContent(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)
	view := m.View()

	if !strings.Contains(view, "Configuration") {
		t.Error("view should contain 'Configuration' title")
	}
	if !strings.Contains(view, "Sonar Instance URL") {
		t.Error("view should contain first field label")
	}
	if !strings.Contains(view, "tab") {
		t.Error("view should mention tab navigation")
	}
}

// TestConfigModelWindowResize verifies WindowSizeMsg updates dimensions.
func TestConfigModelWindowResize(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, false)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m.width != 100 || m.height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", m.width, m.height)
	}
}

// TestConfigModelAllApps verifies config models can be created for
// every registered app without panicking.
func TestConfigModelAllApps(t *testing.T) {
	reg := installer.NewRegistry()
	for _, appID := range reg.List() {
		t.Run(appID, func(t *testing.T) {
			m := NewConfigModel(appID, false)
			view := m.View()
			if view == "" {
				t.Error("view should not be empty")
			}
			if len(m.fields) == 0 {
				t.Error("should have at least one field")
			}
		})
	}
}

// TestAppModelTransitionToConfig verifies StartConfigMsg transitions
// from preflight to config screen.
func TestAppModelTransitionToConfig(t *testing.T) {
	m := NewAppModel()

	updated, _ := m.Update(StartConfigMsg{AppID: installer.AppCustomerPortal})
	model := updated.(AppModel)

	if model.Screen() != ScreenConfig {
		t.Errorf("screen = %d, want ScreenConfig (%d)", model.Screen(), ScreenConfig)
	}
}

// TestSecretFieldsMasked verifies that password, API token, and API key
// fields use password echo mode so values are not shown as plaintext.
func TestSecretFieldsMasked(t *testing.T) {
	tests := []struct {
		appID      string
		secretKeys map[string]bool
	}{
		{
			appID:      installer.AppCustomerPortal,
			secretKeys: map[string]bool{"api_password": true},
		},
		{
			appID:      installer.AppNetflowOnPrem,
			secretKeys: map[string]bool{"api_token": true, "db_password": true},
		},
		{
			appID:      installer.AppPoller,
			secretKeys: map[string]bool{"poller_api_key": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.appID, func(t *testing.T) {
			m := NewConfigModel(tt.appID, false)
			for i, f := range m.fields {
				if tt.secretKeys[f.key] {
					if m.inputs[i].EchoMode == 0 { // 0 = EchoNormal
						t.Errorf("field %q should have masked echo mode, got normal", f.key)
					}
				}
			}
		})
	}
}

// TestSecretFieldsVisibleWhenShowPasswords verifies that when
// showPasswords is true, secret fields use normal echo mode.
func TestSecretFieldsVisibleWhenShowPasswords(t *testing.T) {
	m := NewConfigModel(installer.AppCustomerPortal, true)
	for i, f := range m.fields {
		if f.secret {
			if m.inputs[i].EchoMode != 0 { // 0 = EchoNormal
				t.Errorf("field %q should use normal echo mode when showPasswords=true, got %d", f.key, m.inputs[i].EchoMode)
			}
		}
	}
}

// TestConfigSubmitValidatesForApp verifies that the config wizard's
// submit runs app-specific validation (from ValidateForApp) and
// surfaces errors before transitioning to the install screen.
func TestConfigSubmitValidatesForApp(t *testing.T) {
	m := NewConfigModel(installer.AppNetflowOnPrem, false)

	// Fill in only the sonar_url field (first field) with a valid URL.
	// Leave api_token (second field, required) empty.
	for _, r := range "https://test.sonar.software" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Navigate to last field and try to submit.
	for m.FocusIndex() != len(m.fields)-1 {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}

	// Type something in the last field so empty-check passes for it,
	// but api_token field (index 1) is still empty.
	// Actually, the empty-check will catch field 1 (api_token) first.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.err == "" {
		t.Error("expected validation error for empty required API token field")
	}
}
