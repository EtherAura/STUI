// portal.go implements the Customer Portal installer.
// The Customer Portal is a Docker-based, prebuilt web portal that
// ISPs deploy to let their subscribers manage their own accounts.
// Repo: https://github.com/SonarSoftwareInc/customer_portal
package installer

import (
	"context"
	"fmt"
	"io"
)

// PortalInstaller handles installation of the Sonar Customer Portal.
type PortalInstaller struct{}

// NewPortalInstaller creates a new Customer Portal installer.
func NewPortalInstaller() *PortalInstaller {
	return &PortalInstaller{}
}

// Name returns the display name for the Customer Portal.
func (p *PortalInstaller) Name() string { return "Customer Portal" }

// Description returns a short summary of what the Customer Portal does.
func (p *PortalInstaller) Description() string {
	return "A prebuilt customer portal for Sonar (Docker-based)"
}

// Requirements returns the system requirements for the Customer Portal.
func (p *PortalInstaller) Requirements() []string {
	return []string{
		"OS: Ubuntu (recommended)",
		"Commands: git, curl, docker",
		"Privileges: root / sudo",
		"CPU: 2+ cores",
		"RAM: 2 GB+",
		"Disk: 10 GB+ free",
	}
}

// HardwareRequirements returns the Sonar-recommended minimums for
// the Customer Portal: 2 cores, 2 GB RAM, 10 GB disk.
func (p *PortalInstaller) HardwareRequirements() HardwareReqs {
	return HardwareReqs{
		MinCPUCores: 2,
		MinRAMMB:    2048,
		MinDiskGB:   10,
	}
}

// PreflightCheck verifies the host meets Customer Portal requirements:
// Ubuntu OS, git and curl available, and root access.
func (p *PortalInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
	result := &PreflightResult{Passed: true}

	osInfo, err := DetectOS()
	if err != nil {
		return nil, fmt.Errorf("detecting OS: %w", err)
	}
	result.OS = osInfo.ID
	result.Version = osInfo.VersionID

	// Check supported OS — warn but do not block on non-Ubuntu.
	if osInfo.ID != "ubuntu" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("unsupported OS: %s (officially supports Ubuntu only)", osInfo.ID))
	}

	// Check required commands.
	for _, cmd := range []string{"git", "curl", "docker"} {
		if !CommandExists(cmd) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("required command not found: %s", cmd))
		}
	}

	// Check hardware against Sonar recommendations.
	reqs := &HardwareReqs{
		MinCPUCores: 2,
		MinRAMMB:    2048,
		MinDiskGB:   10,
	}
	hw, hwErr := DetectHardware()
	if hwErr == nil {
		result.Hardware = hw
		result.HardwareReqs = reqs
		result.Warnings = append(result.Warnings, CheckHardware(hw, reqs)...)
	}

	// Check root — flag for sudo/doas relaunch option.
	if !IsRoot() {
		result.NeedsRoot = true
		result.Escalation = DetectEscalation()
		result.Warnings = append(result.Warnings, "not running as root; elevated privileges are required")
	}

	return result, nil
}

// Steps returns the ordered installation steps for the Customer Portal.
func (p *PortalInstaller) Steps() []Step {
	return []Step{
		{Name: "Install prerequisites", Action: p.installPrereqs},
		{Name: "Clone repository", Action: p.cloneRepo},
		{Name: "Run install script", Action: p.runInstall},
	}
}

// Install runs the full Customer Portal installation. It validates the
// config (requires Domain), then executes each step sequentially.
func (p *PortalInstaller) Install(ctx context.Context, cfg *Config, output io.Writer) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.Domain == "" {
		return fmt.Errorf("domain is required for customer portal")
	}

	for _, step := range p.Steps() {
		_, _ = fmt.Fprintf(output, "==> %s\n", step.Name)
		if err := step.Action(ctx, cfg, output); err != nil {
			return fmt.Errorf("step %q failed: %w", step.Name, err)
		}
	}
	return nil
}

// Verify checks that the Customer Portal installed successfully.
func (p *PortalInstaller) Verify(ctx context.Context) error {
	// TODO: Check that Docker containers are running
	return nil
}

// installPrereqs installs system packages needed by the portal (git, unzip).
func (p *PortalInstaller) installPrereqs(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: apt-get install git unzip
	return nil
}

// cloneRepo clones the customer_portal repository from GitHub.
func (p *PortalInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: git clone https://github.com/SonarSoftwareInc/customer_portal.git
	return nil
}

// runInstall executes the portal's Docker-based install script.
func (p *PortalInstaller) runInstall(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: sudo ./install.sh
	return nil
}
