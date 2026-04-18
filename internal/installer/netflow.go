// netflow.go implements the Netflow On-Prem installer.
// Netflow On-Prem is a Docker-based flow data processor that receives
// NetFlow/sFlow data and forwards it to Sonar for network monitoring.
// Repo: https://github.com/SonarSoftwareInc/netflow-onprem
package installer

import (
	"context"
	"fmt"
	"io"
	"strings"
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
		"Commands: git, make, unzip (Docker installed by install.sh)",
		"Privileges: root / sudo",
		"CPU: 2+ cores",
		"RAM: 4 GB+",
		"Disk: 50 GB+ free",
	}
}

// HardwareRequirements returns the Sonar-recommended minimums for
// Netflow On-Prem: 2 cores, 4 GB RAM, 50 GB disk for flow storage.
func (n *NetflowInstaller) HardwareRequirements() HardwareReqs {
	return HardwareReqs{
		MinCPUCores: 2,
		MinRAMMB:    4096,
		MinDiskGB:   50,
	}
}

// PreflightCheck verifies the host meets Netflow requirements:
// Ubuntu or Debian OS, git/make/unzip available, and root access.
func (n *NetflowInstaller) PreflightCheck(ctx context.Context, cfg *Config) (*PreflightResult, error) {
	target := cfg.Target
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

	// Check supported OS — warn but do not block on non-Ubuntu/Debian.
	if osInfo.ID != "ubuntu" && osInfo.ID != "debian" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("unsupported OS: %s (officially supports Ubuntu or Debian)", osInfo.ID))
	}

	// Check commands needed before the installer can bootstrap the host.
	for _, cmd := range []string{"git", "make", "unzip"} {
		if !CommandExistsOn(system, cmd) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("required command not found: %s", cmd))
		}
	}

	// Docker is installed by the upstream install.sh, so missing it should not
	// block preflight. Surface it as an informational warning instead.
	if !CommandExistsOn(system, "docker") {
		result.Warnings = append(result.Warnings, "docker not found; upstream install.sh is expected to install Docker")
	}

	// Check hardware against Sonar recommendations.
	reqs := &HardwareReqs{
		MinCPUCores: 2,
		MinRAMMB:    4096,
		MinDiskGB:   50,
	}
	hw, hwErr := DetectHardwareOn(system)
	if hwErr == nil {
		result.Hardware = hw
		result.HardwareReqs = reqs
		result.Warnings = append(result.Warnings, CheckHardware(hw, reqs)...)
	}

	// Local installs may relaunch STUI with local privilege escalation.
	// Remote targets should not trigger a local sudo/doas restart, but they
	// still need a usable remote privilege escalation path.
	if target.Mode == TargetModeLocal && !system.IsRoot() {
		result.NeedsRoot = true
		result.Escalation = system.DetectEscalation()
		result.Warnings = append(result.Warnings, "not running as root; elevated privileges are required")
	} else if target.Mode == TargetModeSSH && !system.IsRoot() {
		result.Escalation = system.DetectEscalation()
		if result.Escalation == nil {
			result.Passed = false
			result.Errors = append(result.Errors,
				"remote target is not root and no sudo/doas command is available; connect as root or install sudo/doas")
		} else {
			msg := fmt.Sprintf("remote target is not root; privileged commands will use %s", result.Escalation.Name)
			if result.Escalation.Name == "sudo" && target.SudoPassword == "" {
				msg += " and may require a sudo password"
			}
			result.Warnings = append(result.Warnings, msg)
		}
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
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	return RunPrivilegedCmd(ctx, cfg.Target, sys, "apt-get update -y && apt-get install -y git make unzip", output)
}

// cloneRepo clones the netflow-onprem repository from GitHub.
// If the directory already exists it pulls the latest changes instead.
func (n *NetflowInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "netflow-onprem")
	cmd := fmt.Sprintf(
		`if [ -d %[1]s/.git ]; then echo "Repository exists, pulling latest..." && cd %[1]s && git pull; else git clone https://github.com/SonarSoftwareInc/netflow-onprem.git %[1]s; fi`,
		shellQuote(dir),
	)
	return RunPrivilegedCmd(ctx, cfg.Target, sys, cmd, output)
}

// configureEnv creates the .env file from .env.example and populates
// it with user-supplied values (API token, public IP, etc.).
func (n *NetflowInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "netflow-onprem")

	var env strings.Builder
	fmt.Fprintf(&env, "SONAR_URL=%s\n", cfg.SonarURL)
	fmt.Fprintf(&env, "API_TOKEN=%s\n", cfg.APIToken)
	if cfg.NetflowName != "" {
		fmt.Fprintf(&env, "NAME=%s\n", cfg.NetflowName)
	}
	fmt.Fprintf(&env, "PUBLIC_IP=%s\n", cfg.PublicIP)
	if cfg.DBPassword != "" {
		fmt.Fprintf(&env, "DB_PASSWORD=%s\n", cfg.DBPassword)
	}
	if cfg.MaxLife != "" {
		fmt.Fprintf(&env, "MAX_LIFE=%s\n", cfg.MaxLife)
	}
	if cfg.MaxSize != "" {
		fmt.Fprintf(&env, "MAX_SIZE=%s\n", cfg.MaxSize)
	}

	envPath := dir + "/.env"
	cmd := fmt.Sprintf("printf '%%s' %s > %s", shellQuote(env.String()), shellQuote(envPath))
	_, _ = fmt.Fprintf(output, "  Writing %s\n", envPath)
	return RunPrivilegedCmd(ctx, cfg.Target, sys, cmd, output)
}

// runInstall executes the netflow Docker-based install script.
func (n *NetflowInstaller) runInstall(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "netflow-onprem")
	return RunPrivilegedCmd(ctx, cfg.Target, sys, fmt.Sprintf("cd %s && chmod +x install.sh && ./install.sh", shellQuote(dir)), output)
}
