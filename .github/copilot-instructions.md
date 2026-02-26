# STUI - Copilot Instructions

## Project Overview

**STUI** (pronounced "Stewie") is a **Sonar Terminal User Interface** — an ncurses-style TUI application built in Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea). It provides a unified interface for installing and managing applications from SonarSoftwareInc's GitHub repositories.

## Target Applications

STUI manages installation of these Sonar tools:

| App | Repo | Install Pattern |
|-----|------|----------------|
| Customer Portal | `SonarSoftwareInc/customer_portal` | `git clone` + `sudo ./install.sh` (Docker-based) |
| Netflow On-Prem | `SonarSoftwareInc/netflow-onprem` | `git clone` + configure `.env` + `sudo ./install.sh` (Docker-based) |
| FreeRADIUS Genie v3 | `SonarSoftwareInc/freeradius_genie-v3` | `git clone` + `./genie` CLI (native, PHP) |
| Poller | `SonarSoftwareInc/poller` | `wget setup.sh` + `sudo ./setup.sh` (native, PHP/composer) |

## Tech Stack

- **Language:** Go
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Model-View-Update architecture)
- **Components:** [Bubbles](https://github.com/charmbracelet/bubbles) (spinners, lists, text inputs, etc.)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Distribution:** Single static binary via GoReleaser, installable via `curl | sh`
- **Future:** GraphQL client for Sonar API queries/mutations

## Architecture

```
stui/
├── cmd/stui/              # CLI entrypoint
├── internal/
│   ├── tui/               # Bubble Tea models, views, styles
│   │   ├── app.go         # Root model
│   │   ├── menu.go        # App selection menu
│   │   ├── installer.go   # Installation progress view
│   │   └── styles.go      # Lip Gloss styles
│   ├── installer/         # Per-app install logic
│   │   ├── interface.go   # Common Installer interface
│   │   ├── portal.go      # Customer Portal installer
│   │   ├── netflow.go     # Netflow installer
│   │   ├── freeradius.go  # FreeRADIUS installer
│   │   └── poller.go      # Poller installer
│   ├── config/            # App config, saved state
│   └── graphql/           # Future: Sonar API client
├── scripts/
│   └── install.sh         # curl-friendly STUI installer
├── docs/                  # Project documentation
├── go.mod
└── README.md
```

## Design Principles

1. **Installer Interface** — Each Sonar app implements a common interface:
   - `PreflightCheck()` — detect OS/version, check root, check dependencies
   - `Configure()` — interactive prompts for `.env` values, API tokens, etc.
   - `Install()` — execute the actual installation steps
   - `Verify()` — confirm installation succeeded

2. **Stream Output** — Shell command output should be streamed into the TUI in real-time so users see progress from underlying install scripts.

3. **Preflight Safety** — Always detect OS, version, and prerequisites before starting installation. Fail fast with clear messages.

4. **Config Wizard** — Prompt for configuration values (Sonar URL, API tokens, domain names) interactively within the TUI rather than requiring manual file edits.

## Coding Standards

- Use `gofmt` / `goimports` for formatting
- Follow standard Go project layout conventions
- Error handling: wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use structured types over raw maps
- Keep TUI models focused — one model per screen/view
- Test installer logic independently from TUI rendering

## Issue Tracking

This project uses **bd (beads)** for issue tracking. Run `bd prime` for workflow context.

**Quick reference:**
- `bd ready` — Find unblocked work
- `bd create "Title" --type task --priority 2` — Create issue
- `bd close <id>` — Complete work
- `bd sync` — Sync with git (run at session end)

## Important Notes

- Target platform is **Ubuntu 24.04 LTS** (matching Sonar's supported OS)
- STUI itself should be cross-compilable but installers target Linux
- All Sonar apps require **root/sudo** for installation
- The customer_portal and netflow-onprem use **Docker**; poller and freeradius are native installs
