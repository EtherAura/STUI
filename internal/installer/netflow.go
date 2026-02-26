// netflow.go implements the Netflow On-Prem installer.
// Netflow On-Prem is a Docker-based flow data processor that receives
// NetFlow/sFlow data and forwards it to Sonar for network monitoring.
// Repo: https://github.com/SonarSoftwareInc/netflow-onprem
package installer

import (
	"context"
	"fmt"
	"io"
)

// NetflowInstaller handles installation of Netflow On-Prem.
type NetflowInstaller struct{}

// NewNetflowInstaller creates a new Netflow On-Prem installer.
func NewNetflowInstaller() *NetflowInstaller {
	return &NetflowInstaller{}
}

// Name returns the display name for Netflow On-Prem.
func (n *NetflowInstaller) Name() string { return "Netflow On-Prem" }

// Description returns a short summary of what Netflow On-Prem does.
func (n *NetflowInstaller) Description() string { return "Netflow on-premise processor (Docker-based)" }

// Requirements returns the system requirements for Netflow On-Prem.
func (n *NetflowInstaller) Requirements() []string {
	return []string{
		"OS: Ubuntu or Debian (recommended)",
		"Commands: git, make, unzip",
		"Privileges: root / sudo",
	}
}

// PreflightCheck verifies the host meets Netflow requirements:
// Ubuntu or Debian OS, git/make/unzip available, and root access.
func (n *NetflowInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
	result := &PreflightResult{Passed: true}

	osInfo, err := DetectOS()
	if err != nil {
		return nil, fmt.Errorf("detecting OS: %w", err)
	}
	result.OS = osInfo.ID
	result.Version = osInfo.VersionID

	// Check supported OS — warn but do not block on non-Ubuntu/Debian.
	if osInfo.ID != "ubuntu" && osInfo.ID != "debian" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("unsupported OS: %s (officially supports Ubuntu or Debian)", osInfo.ID))
	}

	// Check required commands.
	for _, cmd := range []string{"git", "make", "unzip"} {
		if !CommandExists(cmd) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("required command not found: %s", cmd))
		}
	}

	// Check root — flag for sudo/doas relaunch option.
	if !IsRoot() {
		result.NeedsRoot = true
		result.Escalation = DetectEscalation()
		result.Warnings = append(result.Warnings, "not running as root; elevated privileges are required")
	}

	return result, nil
}

// Steps returns the ordered installation steps for Netflow On-Prem.
func (n *NetflowInstaller) Steps() []Step {
	return []Step{
		{Name: "Install prerequisites", Action: n.installPrereqs},
		{Name: "Clone repository", Action: n.cloneRepo},
		{Name: "Configure environment", Action: n.configureEnv},
		{Name: "Run install script", Action: n.runInstall},
	}
}

// Install runs the full Netflow On-Prem installation. It validates the
// config (requires APIToken and PublicIP), then executes each step.
func (n *NetflowInstaller) Install(ctx context.Context, cfg *Config, output io.Writer) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.APIToken == "" {
		return fmt.Errorf("API token is required for netflow")
	}
	if cfg.PublicIP == "" {
		return fmt.Errorf("public IP is required for netflow")
	}

	for _, step := range n.Steps() {
		_, _ = fmt.Fprintf(output, "==> %s\n", step.Name)
		if err := step.Action(ctx, cfg, output); err != nil {
			return fmt.Errorf("step %q failed: %w", step.Name, err)
		}
	}
	return nil
}

// Verify checks that Netflow On-Prem installed successfully.
func (n *NetflowInstaller) Verify(ctx context.Context) error {
	// TODO: Check Docker containers are running
	return nil
}

// installPrereqs installs system packages needed by netflow (git, make, unzip).
func (n *NetflowInstaller) installPrereqs(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: apt-get install git make unzip
	return nil
}

// cloneRepo clones the netflow-onprem repository from GitHub.
func (n *NetflowInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: git clone https://github.com/SonarSoftwareInc/netflow-onprem.git
	return nil
}

// configureEnv creates the .env file from .env.example and populates
// it with user-supplied values (API token, public IP, etc.).
func (n *NetflowInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: copy .env.example to .env and populate values
	return nil
}

// runInstall executes the netflow Docker-based install script.
func (n *NetflowInstaller) runInstall(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: chmod +x ./install.sh && sudo ./install.sh
	return nil
}
