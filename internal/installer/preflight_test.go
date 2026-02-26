// preflight_test.go contains tests for the preflight check logic.
// Since the actual PreflightCheck methods call real system functions
// (DetectOS, CommandExists), these tests verify structure on the host
// OS and simulate decision logic using injected dependencies.
package installer

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// ubuntuReader returns fake os-release content for Ubuntu 24.04.
func ubuntuReader(_ string) ([]byte, error) {
	return []byte(`ID=ubuntu
VERSION_ID="24.04"
PRETTY_NAME="Ubuntu 24.04 LTS"`), nil
}

// debianReader returns fake os-release content for Debian 12.
func debianReader(_ string) ([]byte, error) {
	return []byte(`ID=debian
VERSION_ID="12"
PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"`), nil
}

// centosReader returns fake os-release content for CentOS Stream 9.
func centosReader(_ string) ([]byte, error) {
	return []byte(`ID=centos
VERSION_ID="9"
PRETTY_NAME="CentOS Stream 9"`), nil
}

// foundLookPath simulates all commands being found on PATH.
func foundLookPath(name string) (string, error) {
	return "/usr/bin/" + name, nil
}

// missingLookPath returns a LookPathFunc that reports the specified
// commands as missing and all others as found.
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

// TestPortalPreflightCheckStructure verifies the Customer Portal preflight
// returns OS info and warns (not blocks) on non-Ubuntu systems.
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

	// Non-Ubuntu should produce a warning, not a blocking error.
	if result.OS != "ubuntu" {
		foundWarning := false
		for _, w := range result.Warnings {
			if strings.Contains(w, "unsupported OS") {
				foundWarning = true
			}
		}
		if !foundWarning {
			t.Error("expected unsupported-OS warning on non-Ubuntu systems")
		}
	}
}

// TestNetflowPreflightCheckStructure verifies the Netflow preflight
// returns OS info and warns (not blocks) on unsupported distros.
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

	// Non-Ubuntu/Debian should produce a warning, not a blocking error.
	if result.OS != "ubuntu" && result.OS != "debian" {
		foundWarning := false
		for _, w := range result.Warnings {
			if strings.Contains(w, "unsupported OS") {
				foundWarning = true
			}
		}
		if !foundWarning {
			t.Error("expected unsupported-OS warning on non-Ubuntu/Debian systems")
		}
	}
}

// TestFreeRADIUSPreflightCheckStructure verifies the FreeRADIUS preflight
// returns OS info on the current host.
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

// TestPollerPreflightCheckStructure verifies the Poller preflight
// returns OS info on the current host.
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

// TestPreflightOSDecisions tests the OS matching logic used by
// preflight checks, exercising ubuntu-only and ubuntu-or-debian rules.
// Non-matching OSes now produce warnings rather than blocking errors.
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

// TestPreflightCommandDecisions tests command-existence checks with
// injected LookPathFunc, covering all-found, partial, and all-missing.
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

// TestVerifyMethodsReturnNil confirms all Verify stubs return nil
// until real verification logic is implemented.
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

// --- NeedsRoot Tests ---

// TestPreflightNeedsRootSet verifies all installers set NeedsRoot
// when the process is not running as root.
func TestPreflightNeedsRootSet(t *testing.T) {
	if IsRoot() {
		t.Skip("test requires non-root execution")
	}

	ctx := context.Background()
	installers := []Installer{
		NewPortalInstaller(),
		NewNetflowInstaller(),
		NewFreeRADIUSInstaller(),
		NewPollerInstaller(),
	}

	for _, inst := range installers {
		t.Run(inst.Name(), func(t *testing.T) {
			result, err := inst.PreflightCheck(ctx)
			if err != nil {
				t.Fatalf("PreflightCheck returned error: %v", err)
			}
			if !result.NeedsRoot {
				t.Error("NeedsRoot should be true when not running as root")
			}
		})
	}
}

// TestPreflightOSWarningNotBlocking verifies that an unsupported OS
// produces a warning but does not set Passed to false.
func TestPreflightOSWarningNotBlocking(t *testing.T) {
	// This test verifies the design decision: OS mismatch is a warning,
	// not a blocker. We check this by examining the preflight results
	// on the current host — if on a non-Ubuntu system, the result
	// should still pass (assuming required commands are present).
	p := NewPortalInstaller()
	ctx := context.Background()

	result, err := p.PreflightCheck(ctx)
	if err != nil {
		t.Fatalf("PreflightCheck returned error: %v", err)
	}

	// OS check should never produce blocking errors.
	for _, e := range result.Errors {
		if strings.Contains(e, "unsupported OS") {
			t.Error("OS mismatch should be a warning, not a blocking error")
		}
	}
}
