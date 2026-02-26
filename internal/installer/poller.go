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

func (p *PollerInstaller) Name() string        { return "Poller" }
func (p *PollerInstaller) Description() string { return "Network monitoring poller for Sonar" }

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

func (p *PollerInstaller) Steps() []Step {
	return []Step{
		{Name: "Download setup script", Action: p.downloadSetup},
		{Name: "Run setup script", Action: p.runSetup},
	}
}

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

func (p *PollerInstaller) Verify(ctx context.Context) error {
	// TODO: Check supervisord process is running
	return nil
}

func (p *PollerInstaller) downloadSetup(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: wget https://raw.githubusercontent.com/SonarSoftwareInc/poller/master/setup.sh
	return nil
}

func (p *PollerInstaller) runSetup(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: chmod +x setup.sh && sudo ./setup.sh
	return nil
}
