// freeradius.go implements the FreeRADIUS Genie v3 installer.
// FreeRADIUS Genie is a native (PHP-based) tool that installs and
// configures FreeRADIUS for use with Sonar's RADIUS integration.
// Repo: https://github.com/SonarSoftwareInc/freeradius_genie-v3
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

// Name returns the display name for FreeRADIUS Genie v3.
func (f *FreeRADIUSInstaller) Name() string { return "FreeRADIUS Genie v3" }

// Description returns a short summary of what FreeRADIUS Genie does.
func (f *FreeRADIUSInstaller) Description() string {
	return "FreeRADIUS installer and configurator for Sonar"
}

// Requirements returns the system requirements for FreeRADIUS Genie.
func (f *FreeRADIUSInstaller) Requirements() []string {
	return []string{
		"OS: Ubuntu (recommended)",
		"Commands: git",
		"Privileges: root / sudo",
	}
}

// PreflightCheck verifies the host meets FreeRADIUS requirements:
// Ubuntu OS, git available, and root access.
func (f *FreeRADIUSInstaller) PreflightCheck(ctx context.Context) (*PreflightResult, error) {
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
	if !CommandExists("git") {
		result.Passed = false
		result.Errors = append(result.Errors, "required command not found: git")
	}

	// Check root — flag for sudo relaunch option.
	if !IsRoot() {
		result.NeedsRoot = true
		result.Warnings = append(result.Warnings, "not running as root; sudo will be required")
	}

	return result, nil
}

// Steps returns the ordered installation steps for FreeRADIUS Genie.
func (f *FreeRADIUSInstaller) Steps() []Step {
	return []Step{
		{Name: "Clone repository", Action: f.cloneRepo},
		{Name: "Configure environment", Action: f.configureEnv},
		{Name: "Run genie", Action: f.runGenie},
	}
}

// Install runs the full FreeRADIUS Genie installation. It validates
// the config, then executes each step sequentially.
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

// Verify checks that FreeRADIUS Genie installed successfully.
func (f *FreeRADIUSInstaller) Verify(ctx context.Context) error {
	// TODO: Check freeradius service is running
	return nil
}

// cloneRepo clones the freeradius_genie-v3 repository from GitHub.
func (f *FreeRADIUSInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: git clone https://github.com/SonarSoftwareInc/freeradius_genie-v3.git
	return nil
}

// configureEnv creates the .env file from .env.example and populates
// it with user-supplied values for the FreeRADIUS configuration.
func (f *FreeRADIUSInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: copy .env.example to .env and populate
	return nil
}

// runGenie executes the genie CLI to complete FreeRADIUS setup.
func (f *FreeRADIUSInstaller) runGenie(ctx context.Context, cfg *Config, output io.Writer) error {
	// TODO: ./genie
	return nil
}
