// privilege_test.go contains unit tests for the privilege escalation
// detection logic, covering scenarios where doas, sudo, both, or
// neither are available.
package installer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"golang.org/x/sys/unix"
)

// TestDetectEscalationDoasOnly verifies that doas is detected when
// only doas is available.
func TestDetectEscalationDoasOnly(t *testing.T) {
	lookup := fakeLookPath(map[string]string{
		"doas": "/usr/bin/doas",
	})

	esc := DetectEscalationWith(lookup)
	if esc == nil {
		t.Fatal("expected escalation method, got nil")
	}
	if esc.Name != "doas" {
		t.Errorf("Name = %q, want %q", esc.Name, "doas")
	}
	if esc.Path != "/usr/bin/doas" {
		t.Errorf("Path = %q, want %q", esc.Path, "/usr/bin/doas")
	}
}

// TestDetectEscalationSudoOnly verifies that sudo is detected when
// only sudo is available.
func TestDetectEscalationSudoOnly(t *testing.T) {
	lookup := fakeLookPath(map[string]string{
		"sudo": "/usr/bin/sudo",
	})

	esc := DetectEscalationWith(lookup)
	if esc == nil {
		t.Fatal("expected escalation method, got nil")
	}
	if esc.Name != "sudo" {
		t.Errorf("Name = %q, want %q", esc.Name, "sudo")
	}
	if esc.Path != "/usr/bin/sudo" {
		t.Errorf("Path = %q, want %q", esc.Path, "/usr/bin/sudo")
	}
}

// TestDetectEscalationPrefersDoas verifies that doas is preferred over
// sudo when both are available.
func TestDetectEscalationPrefersDoas(t *testing.T) {
	lookup := fakeLookPath(map[string]string{
		"doas": "/usr/bin/doas",
		"sudo": "/usr/bin/sudo",
	})

	esc := DetectEscalationWith(lookup)
	if esc == nil {
		t.Fatal("expected escalation method, got nil")
	}
	if esc.Name != "doas" {
		t.Errorf("Name = %q, want %q — doas should be preferred", esc.Name, "doas")
	}
}

// TestDetectEscalationNoneFound verifies nil is returned when neither
// doas nor sudo is available.
func TestDetectEscalationNoneFound(t *testing.T) {
	lookup := fakeLookPath(map[string]string{})

	esc := DetectEscalationWith(lookup)
	if esc != nil {
		t.Errorf("expected nil, got %+v", esc)
	}
}

// TestEscalationMethodFields verifies the EscalationMethod struct
// stores the expected values.
func TestEscalationMethodFields(t *testing.T) {
	esc := EscalationMethod{Name: "sudo", Path: "/usr/bin/sudo"}
	if esc.Name != "sudo" {
		t.Errorf("Name = %q, want %q", esc.Name, "sudo")
	}
	if esc.Path != "/usr/bin/sudo" {
		t.Errorf("Path = %q, want %q", esc.Path, "/usr/bin/sudo")
	}
}

func TestPrivilegedCommand(t *testing.T) {
	tests := []struct {
		name    string
		target  Target
		isRoot  bool
		esc     *EscalationMethod
		command string
		want    string
		wantErr string
	}{
		{
			name:    "root runs command directly",
			target:  Target{Mode: TargetModeSSH, Host: "192.0.2.10", User: "root"},
			isRoot:  true,
			command: "apt-get update -y",
			want:    "apt-get update -y",
		},
		{
			name:    "remote sudo wraps command",
			target:  Target{Mode: TargetModeSSH, Host: "192.0.2.10", User: "ubuntu"},
			esc:     &EscalationMethod{Name: "sudo", Path: "/usr/bin/sudo"},
			command: "apt-get update -y",
			want:    "sudo -n sh -lc 'apt-get update -y'",
		},
		{
			name:    "remote doas wraps command",
			target:  Target{Mode: TargetModeSSH, Host: "192.0.2.10", User: "ubuntu"},
			esc:     &EscalationMethod{Name: "doas", Path: "/usr/bin/doas"},
			command: "apt-get update -y",
			want:    "doas -n sh -lc 'apt-get update -y'",
		},
		{
			name:    "local non-root errors",
			target:  Target{},
			command: "apt-get update -y",
			wantErr: "relaunch STUI locally",
		},
		{
			name:    "remote non-root without escalation errors",
			target:  Target{Mode: TargetModeSSH, Host: "192.0.2.10", User: "ubuntu"},
			command: "apt-get update -y",
			wantErr: "no sudo/doas command is available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := privilegedCommand(tt.target, tt.isRoot, tt.esc, tt.command)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

type stubSystem struct {
	isRoot  bool
	esc     *EscalationMethod
	runErr  error
	lastCmd string
}

func (s *stubSystem) ReadFile(path string) ([]byte, error) { return nil, errors.New("unused") }
func (s *stubSystem) LookPath(name string) (string, error) { return "", errors.New("unused") }
func (s *stubSystem) IsRoot() bool                         { return s.isRoot }
func (s *stubSystem) DetectEscalation() *EscalationMethod  { return s.esc }
func (s *stubSystem) NumCPU() int                          { return 0 }
func (s *stubSystem) Statfs(path string, buf *unix.Statfs_t) error {
	return errors.New("unused")
}
func (s *stubSystem) RunCmd(ctx context.Context, command string, output io.Writer) error {
	s.lastCmd = command
	return s.runErr
}

func TestRunPrivilegedCmdRemoteErrorMessage(t *testing.T) {
	sys := &stubSystem{
		esc:    &EscalationMethod{Name: "sudo", Path: "/usr/bin/sudo"},
		runErr: errors.New("sudo: a password is required"),
	}

	err := RunPrivilegedCmd(
		context.Background(),
		Target{Mode: TargetModeSSH, Host: "192.0.2.10", User: "ubuntu"},
		sys,
		"apt-get update -y",
		io.Discard,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "connect as root or configure passwordless sudo") {
		t.Fatalf("error = %q, want remote guidance", err.Error())
	}
	if sys.lastCmd != "sudo -n sh -lc 'apt-get update -y'" {
		t.Fatalf("RunCmd command = %q", sys.lastCmd)
	}
}

// fakeLookPath returns a LookPathFunc that resolves commands from the
// provided map. Commands not in the map return an error.
func fakeLookPath(available map[string]string) LookPathFunc {
	return func(name string) (string, error) {
		if path, ok := available[name]; ok {
			return path, nil
		}
		return "", fmt.Errorf("command %q not found", name)
	}
}
