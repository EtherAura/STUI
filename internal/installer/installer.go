// Package installer defines the interface and types for Sonar application installers.
package installer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Step represents a single installation step with a name and action.
type Step struct {
	Name   string
	Action func(ctx context.Context, cfg *Config, output io.Writer) error
}

// Config holds the configuration values collected from the user.
type Config struct {
	// Common fields
	SonarURL string
	APIToken string

	// Customer Portal specific
	APIUsername string
	APIPassword string
	Domain      string
	Email       string

	// Netflow specific
	NetflowName string
	PublicIP    string
	DBPassword  string
	MaxLife     string
	MaxSize     string

	// Poller specific
	PollerAPIKey string

	// Extra holds any additional key-value pairs.
	Extra map[string]string
}

// Validate checks that required common fields are set.
func (c *Config) Validate() error {
	if c.SonarURL == "" {
		return fmt.Errorf("sonar URL is required")
	}
	if !strings.HasPrefix(c.SonarURL, "https://") {
		return fmt.Errorf("sonar URL must start with https://")
	}
	if strings.HasSuffix(c.SonarURL, "/") {
		return fmt.Errorf("sonar URL must not have a trailing slash")
	}
	return nil
}

// PreflightResult contains the results of preflight checks.
type PreflightResult struct {
	Passed   bool
	OS       string
	Version  string
	Errors   []string
	Warnings []string
}

// Installer is the interface that all Sonar application installers implement.
type Installer interface {
	// Name returns the display name of the application.
	Name() string

	// Description returns a short description of the application.
	Description() string

	// PreflightCheck verifies the system meets requirements.
	PreflightCheck(ctx context.Context) (*PreflightResult, error)

	// Steps returns the ordered installation steps.
	Steps() []Step

	// Install executes the full installation with the given config.
	// Output from commands is written to the provided writer.
	Install(ctx context.Context, cfg *Config, output io.Writer) error

	// Verify checks that the installation was successful.
	Verify(ctx context.Context) error
}

// OSInfo holds detected operating system information.
type OSInfo struct {
	ID         string // e.g., "ubuntu"
	VersionID  string // e.g., "24.04"
	PrettyName string // e.g., "Ubuntu 24.04 LTS"
	Arch       string // e.g., "amd64"
}

// ReadFileFunc is a function that reads a file and returns its contents.
// Allows dependency injection for testing.
type ReadFileFunc func(path string) ([]byte, error)

// DetectOS reads /etc/os-release to determine the operating system.
func DetectOS() (*OSInfo, error) {
	return DetectOSWith(os.ReadFile)
}

// DetectOSWith reads /etc/os-release using the provided file reader.
// This allows tests to inject fake file contents.
func DetectOSWith(readFile ReadFileFunc) (*OSInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("unsupported operating system: %s (only linux is supported)", runtime.GOOS)
	}

	info := &OSInfo{
		Arch: runtime.GOARCH,
	}

	data, err := readFile("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("reading os-release: %w", err)
	}

	info.ID, info.VersionID, info.PrettyName = ParseOSRelease(string(data))

	if info.ID == "" {
		return nil, fmt.Errorf("could not determine OS from /etc/os-release")
	}

	return info, nil
}

// ParseOSRelease extracts ID, VERSION_ID, and PRETTY_NAME from os-release content.
func ParseOSRelease(data string) (id, versionID, prettyName string) {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(value, `"`)
		switch key {
		case "ID":
			id = value
		case "VERSION_ID":
			versionID = value
		case "PRETTY_NAME":
			prettyName = value
		}
	}
	return
}

// IsRoot returns true if the current process is running as root.
func IsRoot() bool {
	return os.Geteuid() == 0
}

// LookPathFunc is a function that looks up a command on PATH.
// Allows dependency injection for testing.
type LookPathFunc func(name string) (string, error)

// CommandExists checks if a command is available on the system PATH.
func CommandExists(name string) bool {
	return CommandExistsWith(name, exec.LookPath)
}

// CommandExistsWith checks if a command exists using the provided lookup function.
func CommandExistsWith(name string, lookPath LookPathFunc) bool {
	_, err := lookPath(name)
	return err == nil
}

// Registry maps application identifiers to their installer constructors.
type Registry map[string]func() Installer

// AppID constants for the supported applications.
const (
	AppCustomerPortal = "customer-portal"
	AppNetflowOnPrem  = "netflow-onprem"
	AppFreeRADIUS     = "freeradius-genie"
	AppPoller         = "poller"
)

// NewRegistry returns a registry with all supported installers.
func NewRegistry() Registry {
	return Registry{
		AppCustomerPortal: func() Installer { return NewPortalInstaller() },
		AppNetflowOnPrem:  func() Installer { return NewNetflowInstaller() },
		AppFreeRADIUS:     func() Installer { return NewFreeRADIUSInstaller() },
		AppPoller:         func() Installer { return NewPollerInstaller() },
	}
}

// List returns the app IDs in display order.
func (r Registry) List() []string {
	return []string{
		AppCustomerPortal,
		AppNetflowOnPrem,
		AppFreeRADIUS,
		AppPoller,
	}
}
