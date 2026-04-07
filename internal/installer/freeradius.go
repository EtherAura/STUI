// freeradius.go implements the FreeRADIUS Genie v3 installer.
// FreeRADIUS Genie is a native (PHP-based) tool that installs and
// configures FreeRADIUS for use with Sonar's RADIUS integration.
// Repo: https://github.com/SonarSoftwareInc/freeradius_genie-v3
package installer

import (
	"context"
	"fmt"
	"io"
	"strings"
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
		"CPU: 1+ core",
		"RAM: 1 GB+",
		"Disk: 10 GB+ free",
	}
}

// HardwareRequirements returns the Sonar-recommended minimums for
// FreeRADIUS Genie: 1 core, 1 GB RAM, 10 GB disk.
func (f *FreeRADIUSInstaller) HardwareRequirements() HardwareReqs {
	return HardwareReqs{
		MinCPUCores: 1,
		MinRAMMB:    1024,
		MinDiskGB:   10,
	}
}

// PreflightCheck verifies the host meets FreeRADIUS requirements:
// Ubuntu OS, git available, and root access.
func (f *FreeRADIUSInstaller) PreflightCheck(ctx context.Context, target Target) (*PreflightResult, error) {
	target.Normalize()
	result := &PreflightResult{Passed: true}
	system, err := SystemForTarget(target)
	if err != nil {
		return nil, fmt.Errorf("resolving target system: %w", err)
	}

	osInfo, err := DetectOSOn(system)
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
	if !CommandExistsOn(system, "git") {
		result.Passed = false
		result.Errors = append(result.Errors, "required command not found: git")
	}

	// Check hardware against Sonar recommendations.
	reqs := &HardwareReqs{
		MinCPUCores: 1,
		MinRAMMB:    1024,
		MinDiskGB:   10,
	}
	hw, hwErr := DetectHardwareOn(system)
	if hwErr == nil {
		result.Hardware = hw
		result.HardwareReqs = reqs
		result.Warnings = append(result.Warnings, CheckHardware(hw, reqs)...)
	}

	// Local installs may relaunch STUI with local privilege escalation.
	// Remote targets should not trigger a local sudo/doas restart.
	if target.Mode == TargetModeLocal && !system.IsRoot() {
		result.NeedsRoot = true
		result.Escalation = system.DetectEscalation()
		result.Warnings = append(result.Warnings, "not running as root; elevated privileges are required")
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
// If the directory already exists it pulls the latest changes instead.
func (f *FreeRADIUSInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "freeradius_genie-v3")
	cmd := fmt.Sprintf(
		`if [ -d %[1]s/.git ]; then echo "Repository exists, pulling latest..." && cd %[1]s && git pull; else git clone https://github.com/SonarSoftwareInc/freeradius_genie-v3.git %[1]s; fi`,
		shellQuote(dir),
	)
	return sys.RunCmd(ctx, cmd, output)
}

// configureEnv creates the .env file from .env.example and populates
// it with user-supplied values for the FreeRADIUS configuration.
func (f *FreeRADIUSInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "freeradius_genie-v3")

	var env strings.Builder
	fmt.Fprintf(&env, "SONAR_URL=%s\n", cfg.SonarURL)

	envPath := dir + "/.env"
	cmd := fmt.Sprintf("printf '%%s' %s > %s", shellQuote(env.String()), shellQuote(envPath))
	_, _ = fmt.Fprintf(output, "  Writing %s\n", envPath)
	return sys.RunCmd(ctx, cmd, output)
}

// runGenie executes the genie CLI to complete FreeRADIUS setup.
func (f *FreeRADIUSInstaller) runGenie(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "freeradius_genie-v3")
	return sys.RunCmd(ctx, fmt.Sprintf("cd %s && chmod +x genie && ./genie", shellQuote(dir)), output)
}
