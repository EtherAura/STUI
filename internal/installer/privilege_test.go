// privilege_test.go contains unit tests for the privilege escalation
// detection logic, covering scenarios where doas, sudo, both, or
// neither are available.
package installer

import (
	"fmt"
	"testing"
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
