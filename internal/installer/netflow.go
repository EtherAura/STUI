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

func (n *NetflowInstaller) Name() string        { return "Netflow On-Prem" }
func (n *NetflowInstaller) Description() string { return "Netflow on-premise processor (Docker-based)" }

func (n *NetflowInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
	result := &PreflightResult{Passed: true}

	osInfo, err := DetectOS()
	if err != nil {
		return nil, fmt.Errorf("detecting OS: %w", err)
	}
	result.OS = osInfo.ID
	result.Version = osInfo.VersionID

	if osInfo.ID != "ubuntu" && osInfo.ID != "debian" {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("unsupported OS: %s (requires Ubuntu or Debian)", osInfo.ID))
	}

	for _, cmd := range []string{"git", "make", "unzip"} {
		if !CommandExists(cmd) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("required command not found: %s", cmd))
		}
	}

	if !IsRoot() {
		result.Warnings = append(result.Warnings, "not running as root; sudo will be required")
	}

	return result, nil
}

func (n *NetflowInstaller) Steps() []Step {
	return []Step{
		{Name: "Install prerequisites", Action: n.installPrereqs},
		{Name: "Clone repository", Action: n.cloneRepo},
		{Name: "Configure environment", Action: n.configureEnv},
		{Name: "Run install script", Action: n.runInstall},
	}
}

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

func (n *NetflowInstaller) Verify(ctx context.Context) error {
	// TODO: Check Docker containers are running
	return nil
}

func (n *NetflowInstaller) installPrereqs(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: apt-get install git make unzip
	return nil
}

func (n *NetflowInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: git clone https://github.com/SonarSoftwareInc/netflow-onprem.git
	return nil
}

func (n *NetflowInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: copy .env.example to .env and populate values
	return nil
}

func (n *NetflowInstaller) runInstall(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: chmod +x ./install.sh && sudo ./install.sh
	return nil
}
