// portal.go implements the Customer Portal installer.
// The Customer Portal is a Docker-based, prebuilt web portal that
// ISPs deploy to let their subscribers manage their own accounts.
// Repo: https://github.com/SonarSoftwareInc/customer_portal
package installer

import (
	"context"
	"fmt"
	"io"
	"strings"
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
// Source: https://github.com/SonarSoftwareInc/customer_portal#quick-start
func (p *PortalInstaller) Requirements() []string {
	return []string{
		"OS: Ubuntu 18.04 or 22.04 x64 (Debian-based; other versions unsupported)",
		"Commands: git, unzip (Docker installed by install.sh)",
		"Privileges: root / sudo",
		"CPU: 2+ vCPUs",
		"RAM: 2 GB+",
		"Disk: 25 GB+ free",
	}
}

// HardwareRequirements returns the Sonar-recommended minimums for
// the Customer Portal: 2 vCPUs, 2 GB RAM, 25 GB disk.
// Source: https://github.com/SonarSoftwareInc/customer_portal#quick-start
func (p *PortalInstaller) HardwareRequirements() HardwareReqs {
	return HardwareReqs{
		MinCPUCores: 2,
		MinRAMMB:    2048,
		MinDiskGB:   25,
	}
}

// PreflightCheck verifies the host meets Customer Portal requirements:
// Ubuntu OS, git and unzip available, root access, and domain DNS resolution.
func (p *PortalInstaller) PreflightCheck(ctx context.Context, cfg *Config) (*PreflightResult, error) {
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

	// Check supported OS — warn but do not block on non-Ubuntu.
	// Source: https://github.com/SonarSoftwareInc/customer_portal#quick-start
	if osInfo.ID != "ubuntu" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("unsupported OS: %s (officially supports Ubuntu 18.04/22.04 only)", osInfo.ID))
	} else if osInfo.VersionID != "18.04" && osInfo.VersionID != "22.04" {
		// The upstream install.sh uses lsb_release -cs for the Docker APT
		// repository. Docker is unsupported on Ubuntu 19.x and blocking issues
		// have been encountered on Ubuntu 24.04.
		result.Passed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("unsupported Ubuntu version: %s (only 18.04 and 22.04 are supported)", osInfo.VersionID))
	}

	// Check commands needed before the installer can bootstrap the host.
	for _, cmd := range []string{"git", "unzip"} {
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
		MinRAMMB:    2048,
		MinDiskGB:   25,
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

	// Verify domain DNS when a domain has been configured. Certbot
	// needs the domain to resolve to a public IP for HTTP-01 challenges.
	if cfg.Domain != "" {
		dns := ResolveDomain(cfg.Domain)
		// Check port reachability on public IPs when DNS resolves.
		if dns.OK() {
			dns.CheckPorts(80, 443)
		}
		result.DNS = dns
		result.Warnings = append(result.Warnings, dns.Warnings...)
		if !dns.OK() {
			result.Passed = false
			result.Errors = append(result.Errors, dns.Errors...)
		}
	}

	return result, nil
}

// Steps returns the ordered installation steps for the Customer Portal.
func (p *PortalInstaller) Steps() []Step {
	return []Step{
		{Name: "Install prerequisites", Action: p.installPrereqs},
		{Name: "Clone repository", Action: p.cloneRepo},
		{Name: "Configure environment", Action: p.configureEnv},
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
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	return RunPrivilegedCmd(ctx, cfg.Target, sys, "apt-get update -y && apt-get install -y git unzip", output)
}

// cloneRepo clones the customer_portal repository from GitHub.
// If the directory already exists it pulls the latest changes instead.
func (p *PortalInstaller) cloneRepo(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "customer_portal")
	cmd := fmt.Sprintf(
		`if [ -d %[1]s/.git ]; then echo "Repository exists, pulling latest..." && cd %[1]s && git pull; else git clone https://github.com/SonarSoftwareInc/customer_portal.git %[1]s; fi`,
		shellQuote(dir),
	)
	return RunPrivilegedCmd(ctx, cfg.Target, sys, cmd, output)
}

// configureEnv writes the .env file directly to the target host so
// that the install commands can source it. This bypasses install.sh's
// fragile interactive prompts entirely.
func (p *PortalInstaller) configureEnv(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "customer_portal")
	trimmedURL := strings.TrimRight(cfg.SonarURL, "/")

	// Stop any running containers before writing a fresh .env.
	// The APP_KEY is generated on the target so every install gets a
	// unique key, matching install.sh behaviour. Each config value is
	// shell-quoted to prevent injection through user-supplied inputs.
	cmd := fmt.Sprintf(
		"cd %s && "+
			"(docker compose stop 2>/dev/null || true) && "+
			"APP_KEY=\"base64:$(head -c32 /dev/urandom | base64)\" && { "+
			"echo \"APP_KEY=$APP_KEY\"; "+
			"echo %s; "+
			"echo %s; "+
			"echo %s; "+
			"echo %s; "+
			"echo %s; "+
			"} > .env",
		shellQuote(dir),
		shellQuote("NGINX_HOST="+cfg.Domain),
		shellQuote("API_USERNAME="+cfg.APIUsername),
		shellQuote("API_PASSWORD="+cfg.APIPassword),
		shellQuote("SONAR_URL="+trimmedURL),
		shellQuote("EMAIL_ADDRESS="+cfg.Email),
	)
	_, _ = io.WriteString(output, "  Writing .env to target...\n")
	return RunPrivilegedCmd(ctx, cfg.Target, sys, cmd, output)
}

// runInstall executes the portal's Docker-based install script.
// configureEnv has already written .env, so install.sh will find it
// and prompt "Set it up again? [y/N]" via read -n 1. That prompt
// reads exactly one byte without consuming the trailing newline, so
// we must NOT place a \n between the "y" answer and the subsequent
// values — otherwise the stale \n gets consumed by the next read as
// an empty line. After accepting "y", install.sh runs docker compose
// stop (freeing ports 80/443), so the port-in-use prompt should not
// appear. The remaining read -ep prompts each consume one \n-delimited
// line: domain, username, password, URL, email.
func (p *PortalInstaller) runInstall(ctx context.Context, cfg *Config, output io.Writer) error {
	sys, err := SystemForTarget(cfg.Target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	dir := repoDir(cfg, "customer_portal")

	// Stdin answers for install.sh:
	//   "y"        — .env exists, set up again? (read -n 1, consumes 1 byte)
	//   domain\n   — portal domain           (read -ep, consumes line)
	//   user\n     — API username             (read -ep, consumes line)
	//   pass\n     — API password             (read -esp, consumes line)
	//   url\n      — Sonar instance URL       (read -ep, consumes line)
	//   email\n    — email address            (read -ep, consumes line)
	//
	// IMPORTANT: No \n after "y". read -n 1 consumes the single 'y'
	// byte and returns immediately. The next read -ep then sees the
	// domain value as its first input line.
	answers := "y" +
		cfg.Domain + "\n" +
		cfg.APIUsername + "\n" +
		cfg.APIPassword + "\n" +
		cfg.SonarURL + "\n" +
		cfg.Email + "\n"

	cmd := fmt.Sprintf("cd %s && chmod +x install.sh && ./install.sh", shellQuote(dir))
	return RunPrivilegedCmdInput(ctx, cfg.Target, sys, cmd, answers, output)
}
