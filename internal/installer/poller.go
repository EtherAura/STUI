// poller.go implements the Sonar Poller installer.
// The Poller is a native (PHP/Composer) agent that performs network
// monitoring tasks (SNMP, ICMP, bandwidth) on behalf of Sonar.
// Repo: https://github.com/SonarSoftwareInc/poller
package installer

import (
	"context"
	"fmt"
	"io"
)

// PollerInstaller handles installation of the Sonar Poller.
type PollerInstaller struct{}

// NewPollerInstaller creates a new Poller installer.
func NewPollerInstaller() *PollerInstaller {
	return &PollerInstaller{}
}

// Name returns the display name for the Sonar Poller.
func (p *PollerInstaller) Name() string { return "Poller" }

// Description returns a short summary of what the Poller does.
func (p *PollerInstaller) Description() string { return "Network monitoring poller for Sonar" }

// PreflightCheck verifies the host meets Poller requirements:
// Ubuntu OS and root access.
func (p *PollerInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
	result := &PreflightResult{Passed: true}

	osInfo, err := DetectOS()
	if err != nil {
		return nil, fmt.Errorf("detecting OS: %w", err)
	}
	result.OS = osInfo.ID
	result.Version = osInfo.VersionID

	if osInfo.ID != "ubuntu" {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("unsupported OS: %s (requires Ubuntu)", osInfo.ID))
	}

	if !IsRoot() {
		result.Warnings = append(result.Warnings, "not running as root; sudo will be required")
	}

	return result, nil
}

// Steps returns the ordered installation steps for the Poller.
func (p *PollerInstaller) Steps() []Step {
	return []Step{
		{Name: "Download setup script", Action: p.downloadSetup},
		{Name: "Run setup script", Action: p.runSetup},
	}
}

// Install runs the full Poller installation. It validates the config,
// then executes each step sequentially.
func (p *PollerInstaller) Install(ctx context.Context, cfg *Config, output io.Writer) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	for _, step := range p.Steps() {
		_, _ = fmt.Fprintf(output, "==> %s\n", step.Name)
		if err := step.Action(ctx, cfg, output); err != nil {
			return fmt.Errorf("step %q failed: %w", step.Name, err)
		}
	}
	return nil
}

// Verify checks that the Poller installed successfully.
func (p *PollerInstaller) Verify(ctx context.Context) error {
	// TODO: Check supervisord process is running
	return nil
}

// downloadSetup fetches the Poller setup script from GitHub.
func (p *PollerInstaller) downloadSetup(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: wget https://raw.githubusercontent.com/SonarSoftwareInc/poller/master/setup.sh
	return nil
}

// runSetup executes the downloaded Poller setup script.
func (p *PollerInstaller) runSetup(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: chmod +x setup.sh && sudo ./setup.sh
	return nil
}
