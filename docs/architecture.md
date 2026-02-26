# Architecture

## Overview

STUI is a Go TUI application built on Bubble Tea's Model-View-Update (MVU) pattern. It wraps SonarSoftwareInc's existing install scripts behind an interactive terminal interface.

## Core Components

### TUI Layer (`internal/tui/`)

Uses the Elm architecture via Bubble Tea:
- **Model** — application state
- **Update** — handles messages (keypresses, command results)
- **View** — renders the current state to a string

Each screen (menu, installer, config wizard) is its own Bubble Tea model.

### Installer Layer (`internal/installer/`)

Each Sonar application has a dedicated installer struct implementing a common interface:

```go
type Installer interface {
    Name() string
    Description() string
    PreflightCheck() error
    Configure(prompts ConfigPrompter) (*Config, error)
    Install(ctx context.Context, cfg *Config, output io.Writer) error
    Verify() error
}
```

Installers are decoupled from the TUI — they can be tested and run independently.

### Config Layer (`internal/config/`)

Manages persistent configuration and state. Stores things like:
- Previously installed applications
- Saved Sonar instance URLs
- Installation logs

## Supported Applications

| Application | Type | Requires Docker | Requires Root |
|------------|------|----------------|---------------|
| Customer Portal | Docker-based | Yes | Yes |
| Netflow On-Prem | Docker-based | Yes | Yes |
| FreeRADIUS Genie v3 | Native | No | Yes |
| Poller | Native | No | Yes |

## Distribution

STUI is distributed as a single static Go binary. Users install via:

```bash
curl -sSL https://raw.githubusercontent.com/YOURORG/stui/main/scripts/install.sh | sudo bash
```

The install script detects architecture and downloads the correct binary from GitHub Releases.
