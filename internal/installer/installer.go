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

// Config holds the configuration values collected from the user
// during the interactive TUI wizard before installation begins.
type Config struct {
	// SonarURL is the base URL of the Sonar instance (e.g., "https://myisp.sonar.software").
	SonarURL string
	// APIToken is the Sonar API bearer token for authenticated requests.
	APIToken string

	// APIUsername is the Sonar API username (Customer Portal only).
	APIUsername string
	// APIPassword is the Sonar API password (Customer Portal only).
	APIPassword string
	// Domain is the FQDN where the portal will be served (Customer Portal only).
	Domain string
	// Email is the admin contact email for TLS certificates (Customer Portal only).
	Email string

	// NetflowName is the collector display name in Sonar (Netflow only).
	NetflowName string
	// PublicIP is the server's public IP for flow data ingestion (Netflow only).
	PublicIP string
	// DBPassword is the database password for the Netflow backend (Netflow only).
	DBPassword string
	// MaxLife is the maximum retention period for flow data (Netflow only).
	MaxLife string
	// MaxSize is the maximum disk size for flow data storage (Netflow only).
	MaxSize string

	// PollerAPIKey is the Sonar API key used by the Poller agent (Poller only).
	PollerAPIKey string

	// Extra holds any additional key-value pairs for future or custom config.
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

// PreflightResult contains the results of preflight checks run
// before installation to verify system compatibility.
type PreflightResult struct {
	// Passed is true if all checks succeeded and installation can proceed.
	Passed bool
	// OS is the detected operating system ID (e.g., "ubuntu").
	OS string
	// Version is the detected OS version (e.g., "24.04").
	Version string
	// Errors lists blocking issues that prevent installation.
	Errors []string
	// Warnings lists non-blocking issues the user should know about.
	Warnings []string
	// NeedsRoot is true when the process is not running as root and
	// the installer requires elevated privileges.
	NeedsRoot bool
	// Escalation is the detected privilege escalation method (sudo/doas),
	// or nil if none is available. Only set when NeedsRoot is true.
	Escalation *EscalationMethod
	// Hardware holds the detected system hardware specs, or nil if
	// detection failed.
	Hardware *HardwareInfo
	// HardwareReqs holds the recommended hardware for this installer,
	// used by the view to render per-metric pass/fail lines.
	HardwareReqs *HardwareReqs
}

// Installer is the interface that all Sonar application installers implement.
type Installer interface {
	// Name returns the display name of the application.
	Name() string

	// Description returns a short description of the application.
	Description() string

	// Requirements returns a human-readable list of system requirements
	// (e.g., supported OS, required commands, root access).
	Requirements() []string

	// HardwareRequirements returns the minimum recommended hardware
	// specifications for the application.
	HardwareRequirements() HardwareReqs

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

// AppID constants identify each supported Sonar application.
const (
	// AppCustomerPortal is the registry key for the Customer Portal installer.
	AppCustomerPortal = "customer-portal"
	// AppNetflowOnPrem is the registry key for the Netflow On-Prem installer.
	AppNetflowOnPrem = "netflow-onprem"
	// AppFreeRADIUS is the registry key for the FreeRADIUS Genie installer.
	AppFreeRADIUS = "freeradius-genie"
	// AppPoller is the registry key for the Poller installer.
	AppPoller = "poller"
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
