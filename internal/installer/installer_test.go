// installer_test.go contains unit tests for the core installer types:
// Config validation, OS release parsing, OS detection, command lookup,
// registry construction, interface compliance, install flow validation,
// and step structure for each Sonar application.
package installer

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

// TestConfigValidate verifies Config.Validate rejects missing, non-HTTPS,
// and trailing-slash URLs while accepting well-formed ones.
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:    "empty URL",
			config:  Config{},
			wantErr: "sonar URL is required",
		},
		{
			name:    "missing https prefix",
			config:  Config{SonarURL: "http://myisp.sonar.software"},
			wantErr: "must start with https://",
		},
		{
			name:    "trailing slash",
			config:  Config{SonarURL: "https://myisp.sonar.software/"},
			wantErr: "must not have a trailing slash",
		},
		{
			name:   "valid URL",
			config: Config{SonarURL: "https://myisp.sonar.software"},
		},
		{
			name:   "valid URL with subdomain",
			config: Config{SonarURL: "https://demo.sonar.software"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestConfigExtra verifies that the Extra map survives validation untouched.
func TestConfigExtra(t *testing.T) {
	cfg := Config{
		SonarURL: "https://myisp.sonar.software",
		Extra:    map[string]string{"custom_key": "custom_value"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Extra["custom_key"] != "custom_value" {
		t.Fatalf("expected extra key to be preserved")
	}
}

// TestParseOSRelease verifies parsing of /etc/os-release content across
// multiple distros, edge cases (comments, blank lines, unquoted values),
// and malformed input.
func TestParseOSRelease(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		wantID     string
		wantVer    string
		wantPretty string
	}{
		{
			name: "ubuntu 24.04",
			data: `NAME="Ubuntu"
VERSION="24.04 LTS (Noble Numbat)"
ID=ubuntu
VERSION_ID="24.04"
PRETTY_NAME="Ubuntu 24.04 LTS"
HOME_URL="https://www.ubuntu.com/"`,
			wantID:     "ubuntu",
			wantVer:    "24.04",
			wantPretty: "Ubuntu 24.04 LTS",
		},
		{
			name: "debian 12",
			data: `PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"
NAME="Debian GNU/Linux"
VERSION_ID="12"
ID=debian`,
			wantID:     "debian",
			wantVer:    "12",
			wantPretty: "Debian GNU/Linux 12 (bookworm)",
		},
		{
			name:       "empty input",
			data:       "",
			wantID:     "",
			wantVer:    "",
			wantPretty: "",
		},
		{
			name: "comments and blank lines",
			data: `# This is a comment

ID=ubuntu
# Another comment
VERSION_ID="22.04"`,
			wantID:  "ubuntu",
			wantVer: "22.04",
		},
		{
			name: "unquoted values",
			data: `ID=gentoo
VERSION_ID=2.15`,
			wantID:  "gentoo",
			wantVer: "2.15",
		},
		{
			name:   "malformed lines ignored",
			data:   "no-equals-sign\nID=ubuntu\nbadline",
			wantID: "ubuntu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ver, pretty := ParseOSRelease(tt.data)
			if id != tt.wantID {
				t.Errorf("ID = %q, want %q", id, tt.wantID)
			}
			if ver != tt.wantVer {
				t.Errorf("VERSION_ID = %q, want %q", ver, tt.wantVer)
			}
			if pretty != tt.wantPretty {
				t.Errorf("PRETTY_NAME = %q, want %q", pretty, tt.wantPretty)
			}
		})
	}
}

// TestDetectOSWith verifies OS detection with injected file readers,
// covering valid data, read errors, and empty os-release files.
func TestDetectOSWith(t *testing.T) {
	t.Run("valid ubuntu", func(t *testing.T) {
		fakeReader := func(path string) ([]byte, error) {
			if path != "/etc/os-release" {
				t.Fatalf("unexpected path: %s", path)
			}
			return []byte(`ID=ubuntu
VERSION_ID="24.04"
PRETTY_NAME="Ubuntu 24.04 LTS"`), nil
		}

		info, err := DetectOSWith(fakeReader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.ID != "ubuntu" {
			t.Errorf("ID = %q, want ubuntu", info.ID)
		}
		if info.VersionID != "24.04" {
			t.Errorf("VersionID = %q, want 24.04", info.VersionID)
		}
		if info.PrettyName != "Ubuntu 24.04 LTS" {
			t.Errorf("PrettyName = %q, want Ubuntu 24.04 LTS", info.PrettyName)
		}
	})

	t.Run("file read error", func(t *testing.T) {
		fakeReader := func(_ string) ([]byte, error) {
			return nil, fmt.Errorf("permission denied")
		}

		_, err := DetectOSWith(fakeReader)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !contains(err.Error(), "reading os-release") {
			t.Errorf("error = %q, want to contain 'reading os-release'", err.Error())
		}
	})

	t.Run("empty os-release", func(t *testing.T) {
		fakeReader := func(_ string) ([]byte, error) {
			return []byte(""), nil
		}

		_, err := DetectOSWith(fakeReader)
		if err == nil {
			t.Fatal("expected error for empty os-release, got nil")
		}
		if !contains(err.Error(), "could not determine OS") {
			t.Errorf("error = %q, want to contain 'could not determine OS'", err.Error())
		}
	})
}

// TestCommandExistsWith verifies command lookup with injected LookPathFunc,
// covering both found and not-found scenarios.
func TestCommandExistsWith(t *testing.T) {
	t.Run("command found", func(t *testing.T) {
		fakeLookPath := func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		}
		if !CommandExistsWith("git", fakeLookPath) {
			t.Error("expected CommandExistsWith to return true for found command")
		}
	})

	t.Run("command not found", func(t *testing.T) {
		fakeLookPath := func(_ string) (string, error) {
			return "", fmt.Errorf("not found")
		}
		if CommandExistsWith("nonexistent", fakeLookPath) {
			t.Error("expected CommandExistsWith to return false for missing command")
		}
	})
}

// TestRegistryList verifies that the registry lists all four apps in
// the expected display order.
func TestRegistryList(t *testing.T) {
	reg := NewRegistry()

	list := reg.List()
	if len(list) != 4 {
		t.Fatalf("expected 4 apps, got %d", len(list))
	}

	expected := []string{AppCustomerPortal, AppNetflowOnPrem, AppFreeRADIUS, AppPoller}
	for i, want := range expected {
		if list[i] != want {
			t.Errorf("list[%d] = %q, want %q", i, list[i], want)
		}
	}
}

// TestRegistryConstructors verifies each registry entry produces a
// non-nil installer with a non-empty Name, Description, and Steps.
func TestRegistryConstructors(t *testing.T) {
	reg := NewRegistry()

	for _, appID := range reg.List() {
		t.Run(appID, func(t *testing.T) {
			constructor, ok := reg[appID]
			if !ok {
				t.Fatalf("app %q not found in registry", appID)
			}
			inst := constructor()
			if inst == nil {
				t.Fatalf("constructor for %q returned nil", appID)
			}
			if inst.Name() == "" {
				t.Errorf("Name() returned empty string for %q", appID)
			}
			if inst.Description() == "" {
				t.Errorf("Description() returned empty string for %q", appID)
			}
			if len(inst.Steps()) == 0 {
				t.Errorf("Steps() returned empty for %q", appID)
			}
		})
	}
}

// TestInstallerInterfaceCompliance is a compile-time check that all
// four installer types satisfy the Installer interface.
func TestInstallerInterfaceCompliance(_ *testing.T) {
	// Verify all installer types satisfy the Installer interface at compile time.
	var _ Installer = (*PortalInstaller)(nil)
	var _ Installer = (*NetflowInstaller)(nil)
	var _ Installer = (*FreeRADIUSInstaller)(nil)
	var _ Installer = (*PollerInstaller)(nil)
}

// TestInstallerRequirements verifies all installers return a non-empty
// list of requirements that includes OS and privilege information.
func TestInstallerRequirements(t *testing.T) {
	installers := []Installer{
		NewPortalInstaller(),
		NewNetflowInstaller(),
		NewFreeRADIUSInstaller(),
		NewPollerInstaller(),
	}

	for _, inst := range installers {
		t.Run(inst.Name(), func(t *testing.T) {
			reqs := inst.Requirements()
			if len(reqs) == 0 {
				t.Error("Requirements() should not be empty")
			}
			hasOS := false
			hasPriv := false
			for _, r := range reqs {
				if contains(r, "OS") {
					hasOS = true
				}
				if contains(r, "root") || contains(r, "sudo") {
					hasPriv = true
				}
			}
			if !hasOS {
				t.Error("Requirements should mention OS")
			}
			if !hasPriv {
				t.Error("Requirements should mention privileges")
			}
		})
	}
}

// TestPortalInstallValidation tests the Customer Portal install flow:
// missing URL, missing domain, and a valid config that runs all steps.
func TestPortalInstallValidation(t *testing.T) {
	p := NewPortalInstaller()
	ctx := context.Background()
	buf := &bytes.Buffer{}

	t.Run("missing sonar URL", func(t *testing.T) {
		err := p.Install(ctx, &Config{}, buf)
		if err == nil {
			t.Fatal("expected error for missing URL")
		}
		if !contains(err.Error(), "invalid config") {
			t.Errorf("error = %q, want to contain 'invalid config'", err.Error())
		}
	})

	t.Run("missing domain", func(t *testing.T) {
		err := p.Install(ctx, &Config{SonarURL: "https://myisp.sonar.software"}, buf)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
		if !contains(err.Error(), "domain is required") {
			t.Errorf("error = %q, want to contain 'domain is required'", err.Error())
		}
	})

	t.Run("valid config runs steps", func(t *testing.T) {
		buf.Reset()
		cfg := &Config{
			SonarURL: "https://myisp.sonar.software",
			Domain:   "portal.myisp.com",
		}
		err := p.Install(ctx, cfg, buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !contains(output, "Install prerequisites") {
			t.Error("expected output to contain step name 'Install prerequisites'")
		}
		if !contains(output, "Clone repository") {
			t.Error("expected output to contain step name 'Clone repository'")
		}
	})
}

// TestNetflowInstallValidation tests the Netflow install flow:
// missing API token, missing public IP, and a valid config.
func TestNetflowInstallValidation(t *testing.T) {
	n := NewNetflowInstaller()
	ctx := context.Background()
	buf := &bytes.Buffer{}

	t.Run("missing API token", func(t *testing.T) {
		err := n.Install(ctx, &Config{SonarURL: "https://myisp.sonar.software"}, buf)
		if err == nil {
			t.Fatal("expected error for missing API token")
		}
		if !contains(err.Error(), "API token is required") {
			t.Errorf("error = %q, want to contain 'API token is required'", err.Error())
		}
	})

	t.Run("missing public IP", func(t *testing.T) {
		err := n.Install(ctx, &Config{
			SonarURL: "https://myisp.sonar.software",
			APIToken: "test-token",
		}, buf)
		if err == nil {
			t.Fatal("expected error for missing public IP")
		}
		if !contains(err.Error(), "public IP is required") {
			t.Errorf("error = %q, want to contain 'public IP is required'", err.Error())
		}
	})

	t.Run("valid config runs steps", func(t *testing.T) {
		buf.Reset()
		cfg := &Config{
			SonarURL: "https://myisp.sonar.software",
			APIToken: "test-token",
			PublicIP: "1.2.3.4",
		}
		err := n.Install(ctx, cfg, buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(buf.String(), "Configure environment") {
			t.Error("expected output to contain step name")
		}
	})
}

// TestPollerInstallValidation tests the Poller install flow with a valid config.
func TestPollerInstallValidation(t *testing.T) {
	p := NewPollerInstaller()
	ctx := context.Background()
	buf := &bytes.Buffer{}

	t.Run("valid config runs steps", func(t *testing.T) {
		cfg := &Config{SonarURL: "https://myisp.sonar.software"}
		err := p.Install(ctx, cfg, buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(buf.String(), "Download setup script") {
			t.Error("expected output to contain step name")
		}
	})
}

// TestFreeRADIUSInstallValidation tests the FreeRADIUS install flow with a valid config.
func TestFreeRADIUSInstallValidation(t *testing.T) {
	f := NewFreeRADIUSInstaller()
	ctx := context.Background()
	buf := &bytes.Buffer{}

	t.Run("valid config runs steps", func(t *testing.T) {
		cfg := &Config{SonarURL: "https://myisp.sonar.software"}
		err := f.Install(ctx, cfg, buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(buf.String(), "Clone repository") {
			t.Error("expected output to contain step name")
		}
	})
}

// TestPortalSteps verifies the Customer Portal returns three ordered steps
// with non-nil actions.
func TestPortalSteps(t *testing.T) {
	p := NewPortalInstaller()
	steps := p.Steps()

	expected := []string{"Install prerequisites", "Clone repository", "Run install script"}
	if len(steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(steps))
	}
	for i, want := range expected {
		if steps[i].Name != want {
			t.Errorf("step[%d].Name = %q, want %q", i, steps[i].Name, want)
		}
		if steps[i].Action == nil {
			t.Errorf("step[%d].Action is nil", i)
		}
	}
}

// TestNetflowSteps verifies Netflow returns four ordered steps.
func TestNetflowSteps(t *testing.T) {
	n := NewNetflowInstaller()
	steps := n.Steps()

	expected := []string{"Install prerequisites", "Clone repository", "Configure environment", "Run install script"}
	if len(steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(steps))
	}
	for i, want := range expected {
		if steps[i].Name != want {
			t.Errorf("step[%d].Name = %q, want %q", i, steps[i].Name, want)
		}
	}
}

// TestFreeRADIUSSteps verifies FreeRADIUS returns three steps ending with "Run genie".
func TestFreeRADIUSSteps(t *testing.T) {
	f := NewFreeRADIUSInstaller()
	steps := f.Steps()

	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	if steps[2].Name != "Run genie" {
		t.Errorf("last step = %q, want 'Run genie'", steps[2].Name)
	}
}

// TestPollerSteps verifies the Poller returns two steps starting with "Download setup script".
func TestPollerSteps(t *testing.T) {
	p := NewPollerInstaller()
	steps := p.Steps()

	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0].Name != "Download setup script" {
		t.Errorf("first step = %q, want 'Download setup script'", steps[0].Name)
	}
}

// contains reports whether substr is found within s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

// searchSubstring performs a brute-force substring search.
func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
