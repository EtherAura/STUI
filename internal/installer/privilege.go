// privilege.go detects available privilege escalation methods (sudo, doas)
// on the host system. The detected method is used by the TUI to offer
// the user a relaunch with elevated privileges when not running as root.
package installer

import "os/exec"

// EscalationMethod identifies a privilege escalation command.
type EscalationMethod struct {
	// Name is the human-readable name (e.g., "sudo", "doas").
	Name string
	// Path is the absolute path to the escalation binary.
	Path string
}

// DetectEscalation probes the system for available privilege escalation
// commands, preferring doas over sudo. Returns nil if neither is found.
func DetectEscalation() *EscalationMethod {
	return DetectEscalationWith(exec.LookPath)
}

// DetectEscalationWith uses the provided lookup function to find a
// privilege escalation command. Allows dependency injection for testing.
func DetectEscalationWith(lookPath LookPathFunc) *EscalationMethod {
	// Prefer doas (simpler, growing adoption on BSDs and some Linux distros).
	for _, cmd := range []string{"doas", "sudo"} {
		if path, err := lookPath(cmd); err == nil {
			return &EscalationMethod{Name: cmd, Path: path}
		}
	}
	return nil
}
