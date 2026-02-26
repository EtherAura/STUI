package installer

import (
	"context"
	"fmt"
	"testing"
)

func ubuntuReader(_ string) ([]byte, error) {
	return []byte(`ID=ubuntu
VERSION_ID="24.04"
PRETTY_NAME="Ubuntu 24.04 LTS"`), nil
}

func debianReader(_ string) ([]byte, error) {
	return []byte(`ID=debian
VERSION_ID="12"
PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"`), nil
}

func centosReader(_ string) ([]byte, error) {
	return []byte(`ID=centos
VERSION_ID="9"
PRETTY_NAME="CentOS Stream 9"`), nil
}

func foundLookPath(name string) (string, error) {
	return "/usr/bin/" + name, nil
}

func missingLookPath(cmds ...string) LookPathFunc {
	missing := make(map[string]bool)
	for _, c := range cmds {
		missing[c] = true
	}
	return func(name string) (string, error) {
		if missing[name] {
			return "", fmt.Errorf("not found: %s", name)
		}
		return "/usr/bin/" + name, nil
	}
}

// --- Portal Preflight Tests ---
// Note: The current PreflightCheck implementations call the real DetectOS() and
// CommandExists() functions directly, making them hard to unit test without
// actually being on Ubuntu. These tests document the expected behavior and
// serve as integration tests on Ubuntu systems.
//
// For now, we test the detectable logic indirectly through DetectOSWith and
// CommandExistsWith, and verify the preflight check structure.

func TestPortalPreflightCheckStructure(t *testing.T) {
	p := NewPortalInstaller()
	ctx := context.Background()

	result, err := p.PreflightCheck(ctx)
	if err != nil {
		t.Fatalf("PreflightCheck returned error: %v", err)
	}

	// On any Linux system, we should get OS info
	if result.OS == "" {
		t.Error("OS should be populated")
	}
	if result.Version == "" {
		t.Error("Version should be populated")
	}

	// If not on Ubuntu, expect errors
	if result.OS != "ubuntu" && result.Passed {
		t.Error("expected Passed=false on non-Ubuntu systems")
	}
}

func TestNetflowPreflightCheckStructure(t *testing.T) {
	n := NewNetflowInstaller()
	ctx := context.Background()

	result, err := n.PreflightCheck(ctx)
	if err != nil {
		t.Fatalf("PreflightCheck returned error: %v", err)
	}

	if result.OS == "" {
		t.Error("OS should be populated")
	}

	// Netflow accepts ubuntu OR debian
	if result.OS != "ubuntu" && result.OS != "debian" && result.Passed {
		t.Error("expected Passed=false on non-Ubuntu/Debian systems")
	}
}

func TestFreeRADIUSPreflightCheckStructure(t *testing.T) {
	f := NewFreeRADIUSInstaller()
	ctx := context.Background()

	result, err := f.PreflightCheck(ctx)
	if err != nil {
		t.Fatalf("PreflightCheck returned error: %v", err)
	}

	if result.OS == "" {
		t.Error("OS should be populated")
	}
}

func TestPollerPreflightCheckStructure(t *testing.T) {
	p := NewPollerInstaller()
	ctx := context.Background()

	result, err := p.PreflightCheck(ctx)
	if err != nil {
		t.Fatalf("PreflightCheck returned error: %v", err)
	}

	if result.OS == "" {
		t.Error("OS should be populated")
	}
}

// --- Simulated Preflight Logic Tests ---
// These test the decision logic that PreflightCheck uses, with injected dependencies.

func TestPreflightOSDecisions(t *testing.T) {
	tests := []struct {
		name       string
		readFile   ReadFileFunc
		acceptedOS []string
		wantPass   bool
	}{
		{
			name:       "ubuntu accepted where ubuntu required",
			readFile:   ubuntuReader,
			acceptedOS: []string{"ubuntu"},
			wantPass:   true,
		},
		{
			name:       "debian rejected where ubuntu required",
			readFile:   debianReader,
			acceptedOS: []string{"ubuntu"},
			wantPass:   false,
		},
		{
			name:       "debian accepted where ubuntu or debian required",
			readFile:   debianReader,
			acceptedOS: []string{"ubuntu", "debian"},
			wantPass:   true,
		},
		{
			name:       "centos rejected everywhere",
			readFile:   centosReader,
			acceptedOS: []string{"ubuntu"},
			wantPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osInfo, err := DetectOSWith(tt.readFile)
			if err != nil {
				t.Fatalf("DetectOSWith error: %v", err)
			}

			accepted := false
			for _, os := range tt.acceptedOS {
				if osInfo.ID == os {
					accepted = true
					break
				}
			}

			if accepted != tt.wantPass {
				t.Errorf("OS %q accepted=%v, want %v", osInfo.ID, accepted, tt.wantPass)
			}
		})
	}
}

func TestPreflightCommandDecisions(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		lookPath LookPathFunc
		wantAll  bool
	}{
		{
			name:     "all commands found",
			commands: []string{"git", "curl"},
			lookPath: foundLookPath,
			wantAll:  true,
		},
		{
			name:     "git missing",
			commands: []string{"git", "curl"},
			lookPath: missingLookPath("git"),
			wantAll:  false,
		},
		{
			name:     "all missing",
			commands: []string{"git", "curl", "make"},
			lookPath: missingLookPath("git", "curl", "make"),
			wantAll:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allFound := true
			for _, cmd := range tt.commands {
				if !CommandExistsWith(cmd, tt.lookPath) {
					allFound = false
				}
			}
			if allFound != tt.wantAll {
				t.Errorf("allFound=%v, want %v", allFound, tt.wantAll)
			}
		})
	}
}

// --- Verify Tests ---

func TestVerifyMethodsReturnNil(t *testing.T) {
	// All Verify() implementations are stubs returning nil for now.
	ctx := context.Background()

	installers := []Installer{
		NewPortalInstaller(),
		NewNetflowInstaller(),
		NewFreeRADIUSInstaller(),
		NewPollerInstaller(),
	}

	for _, inst := range installers {
		t.Run(inst.Name(), func(t *testing.T) {
			if err := inst.Verify(ctx); err != nil {
				t.Errorf("Verify() returned unexpected error: %v", err)
			}
		})
	}
}
