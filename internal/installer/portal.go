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

func (p *PortalInstaller) Name() string { return "Customer Portal" }
func (p *PortalInstaller) Description() string {
	return "A prebuilt customer portal for Sonar (Docker-based)"
}

func (p *PortalInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
	result := &PreflightResult{Passed: true}

	osInfo, err := DetectOS()
	if err != nil {
		return nil, fmt.Errorf("detecting OS: %w", err)
	}
	result.OS = osInfo.ID
	result.Version = osInfo.VersionID

	// Check supported OS
	if osInfo.ID != "ubuntu" {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("unsupported OS: %s (requires Ubuntu)", osInfo.ID))
	}

	// Check required commands
	for _, cmd := range []string{"git", "curl"} {
		if !CommandExists(cmd) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("required command not found: %s", cmd))
		}
	}

	// Check root
	if !IsRoot() {
		result.Warnings = append(result.Warnings, "not running as root; sudo will be required")
	}

	return result, nil
}

func (p *PortalInstaller) Steps() []Step {
	return []Step{
		{Name: "Install prerequisites", Action: p.installPrereqs},
		{Name: "Clone repository", Action: p.cloneRepo},
		{Name: "Run install script", Action: p.runInstall},
	}
}

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

func (p *PortalInstaller) Verify(ctx context.Context) error {
	// TODO: Check that Docker containers are running
	return nil
}

func (p *PortalInstaller) installPrereqs(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: apt-get install git unzip
	return nil
}

func (p *PortalInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: git clone https://github.com/SonarSoftwareInc/customer_portal.git
	return nil
}

func (p *PortalInstaller) runInstall(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: sudo ./install.sh
	return nil
}
