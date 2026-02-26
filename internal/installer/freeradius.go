package installer

import (
	"context"
	"fmt"
	"io"
)

// FreeRADIUSInstaller handles installation of FreeRADIUS Genie v3.
type FreeRADIUSInstaller struct{}

// NewFreeRADIUSInstaller creates a new FreeRADIUS Genie installer.
func NewFreeRADIUSInstaller() *FreeRADIUSInstaller {
	return &FreeRADIUSInstaller{}
}

func (f *FreeRADIUSInstaller) Name() string { return "FreeRADIUS Genie v3" }
func (f *FreeRADIUSInstaller) Description() string {
	return "FreeRADIUS installer and configurator for Sonar"
}

func (f *FreeRADIUSInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
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

	if !CommandExists("git") {
		result.Passed = false
		result.Errors = append(result.Errors, "required command not found: git")
	}

	if !IsRoot() {
		result.Warnings = append(result.Warnings, "not running as root; sudo will be required")
	}

	return result, nil
}

func (f *FreeRADIUSInstaller) Steps() []Step {
	return []Step{
		{Name: "Clone repository", Action: f.cloneRepo},
		{Name: "Configure environment", Action: f.configureEnv},
		{Name: "Run genie", Action: f.runGenie},
	}
}

func (f *FreeRADIUSInstaller) Install(ctx context.Context, cfg *Config, output io.Writer) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	for _, step := range f.Steps() {
		_, _ = fmt.Fprintf(output, "==> %s\n", step.Name)
		if err := step.Action(ctx, cfg, output); err != nil {
			return fmt.Errorf("step %q failed: %w", step.Name, err)
		}
	}
	return nil
}

func (f *FreeRADIUSInstaller) Verify(ctx context.Context) error {
	// TODO: Check freeradius service is running
	return nil
}

func (f *FreeRADIUSInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: git clone https://github.com/SonarSoftwareInc/freeradius_genie-v3.git
	return nil
}

func (f *FreeRADIUSInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: copy .env.example to .env and populate
	return nil
}

func (f *FreeRADIUSInstaller) runGenie(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: ./genie
	return nil
}
