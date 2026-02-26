# Usage

## Installing STUI

STUI is distributed as a single static binary. Install it with one command:

```bash
curl -sSL https://raw.githubusercontent.com/EtherAura/stui/main/scripts/install.sh | sudo bash
```

This detects your architecture, downloads the correct binary from GitHub Releases, and places it in `/usr/local/bin/stui`.

### Manual Install

Download a release binary directly from [GitHub Releases](https://github.com/EtherAura/stui/releases), extract it, and move it to a directory on your PATH:

```bash
tar xzf stui_linux_amd64.tar.gz
sudo mv stui /usr/local/bin/
```

## Running STUI

```bash
stui
```

STUI launches a full-screen terminal interface. No flags or arguments are needed for normal use.

### Requirements

- **OS:** Ubuntu 24.04 LTS (primary target). May work on other Debian-based distributions.
- **Privileges:** Root/sudo is required for installing Sonar applications. STUI will prompt for elevation when needed.
- **Terminal:** Any terminal emulator that supports 256 colors. True color is supported when available.

## Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Select / confirm |
| `Esc` | Go back |
| `q` | Quit |
| `?` | Toggle help |

## Workflow

### 1. Select an Application

The main menu displays the available Sonar applications:

- **Customer Portal** — Self-hosted customer portal (Docker-based)
- **Netflow On-Prem** — Netflow on-premise processor (Docker-based)
- **FreeRADIUS Genie v3** — FreeRADIUS installer and configurator (native)
- **Poller** — Network monitoring poller (native)

Use arrow keys or `j`/`k` to highlight an application, then press `Enter`.

### 2. Preflight Checks

STUI automatically verifies your system before proceeding:

- Operating system and version detection
- Root/sudo availability
- Required dependencies (git, Docker, etc.)
- Network connectivity to GitHub

If any check fails, STUI shows a clear error message explaining what's needed and how to fix it.

### 3. Configuration Wizard

Each application prompts for the values it needs. For example:

**Customer Portal:**
- Sonar instance URL (e.g., `https://myisp.sonar.software`)
- API username
- API password
- Domain name (e.g., `portal.myisp.com`)
- Email address for SSL certificates

**Poller:**
- Sonar instance URL
- Poller API key (from Sonar Settings → Monitoring → Pollers)

STUI validates inputs as you type. Passwords are masked. You can go back to change previous answers.

### 4. Installation

After configuration, STUI starts the installation process. You'll see:

- Real-time output from the underlying install scripts
- A progress indicator showing the current step
- Clear success/failure status for each step

Installation times vary by application:
- **Customer Portal:** 5–15 minutes (downloads Docker images)
- **Netflow On-Prem:** 15–30 minutes (builds Docker images)
- **FreeRADIUS Genie v3:** 5–10 minutes
- **Poller:** 5–10 minutes

### 5. Verification

After installation completes, STUI runs verification checks to confirm everything is working:

- Service is running
- Ports are listening
- Web interface is accessible (where applicable)

## Logging

STUI writes logs to `~/.local/share/stui/logs/`. Each installation creates a timestamped log file with the full output of all commands that were run.

## Uninstalling STUI

To remove STUI itself:

```bash
sudo rm /usr/local/bin/stui
rm -rf ~/.local/share/stui
```

This does **not** uninstall any Sonar applications. Each application has its own removal process — consult the application's documentation.
