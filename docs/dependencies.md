# Dependencies

This document describes every dependency used by STUI, why it's needed, and what version is pinned in `go.mod`.

---

## Runtime Dependencies

These are the libraries compiled into the STUI binary.

### Bubble Tea — `github.com/charmbracelet/bubbletea` v1.3.10

**The TUI framework.** Bubble Tea implements the Elm Architecture (Model-View-Update) for Go terminal applications. Every screen in STUI — the app selection menu, config wizards, install progress views — is a Bubble Tea `Model` with three methods:

- `Init()` — returns an initial command (e.g., start a spinner)
- `Update(msg)` — handles input events and command results, returns new state
- `View()` — renders the current state to a string

Bubble Tea handles raw terminal mode, input parsing, alternate screen buffer, mouse events, and framerate-based rendering. It's the foundation everything else sits on.

- **Repository:** https://github.com/charmbracelet/bubbletea
- **License:** MIT

### Bubbles — `github.com/charmbracelet/bubbles` v1.0.0

**Pre-built TUI components.** A library of composable, ready-made Bubble Tea models for common UI elements:

| Component | What STUI uses it for |
|-----------|----------------------|
| `list` | Application selection menu |
| `spinner` | Installation progress indicators |
| `textinput` | Config wizard prompts (URLs, API tokens, passwords) |
| `viewport` | Scrollable output from install scripts |
| `progress` | Installation step progress bars |
| `help` | Keybinding help footer |
| `key` | Keybinding definitions |

Each component follows the same Model-View-Update pattern and can be nested inside STUI's own models.

- **Repository:** https://github.com/charmbracelet/bubbles
- **License:** MIT

### Lip Gloss — `github.com/charmbracelet/lipgloss` v1.1.0

**Terminal styling.** Lip Gloss provides a CSS-like API for styling terminal output — colors, bold, italic, padding, margins, borders, alignment. It's how STUI gets consistent, branded visual styling without ANSI escape code wrangling.

```go
var titleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#7D56F4")).
    PaddingLeft(2)
```

Lip Gloss is terminal-aware — it detects color profile support (true color, 256 color, ANSI, no color) and degrades gracefully.

- **Repository:** https://github.com/charmbracelet/lipgloss
- **License:** MIT

### Charm Log — `github.com/charmbracelet/log` v0.4.2

**Structured logging.** A pretty-printing structured logger that integrates visually with Charm's styling ecosystem. Used for debug/verbose output and logging install steps.

```go
log.Info("Installing", "app", "poller", "step", 3)
```

Supports log levels, key-value pairs, colorized output, and can be silenced in non-verbose mode.

- **Repository:** https://github.com/charmbracelet/log
- **License:** MIT

---

## Indirect Dependencies

These are pulled in transitively by the libraries above. You don't import them directly, but they're compiled into the binary.

| Package | Purpose |
|---------|---------|
| `charmbracelet/colorprofile` v0.4.1 | Detects terminal color capabilities (true color, 256, ANSI, none) |
| `charmbracelet/x/ansi` v0.11.6 | ANSI escape sequence parsing and generation |
| `charmbracelet/x/cellbuf` v0.0.15 | Cell-based terminal buffer for efficient rendering |
| `charmbracelet/x/term` v0.2.2 | Terminal size detection and raw mode management |
| `aymanbagabas/go-osc52` v2.0.1 | OSC52 clipboard support (copy to clipboard from TUI) |
| `lucasb-eyer/go-colorful` v1.3.0 | Color space conversions (used by Lip Gloss) |
| `mattn/go-isatty` v0.0.20 | Detects whether stdout is a terminal |
| `mattn/go-localereader` v0.0.1 | Locale-aware input reading |
| `mattn/go-runewidth` v0.0.19 | Calculates display width of Unicode runes (CJK, emoji) |
| `muesli/ansi` v0.0.0 | ANSI sequence detection and stripping |
| `muesli/cancelreader` v0.2.2 | Cancelable stdin reader for graceful shutdown |
| `muesli/termenv` v0.16.0 | Terminal feature detection and environment adaptation |
| `rivo/uniseg` v0.4.7 | Unicode text segmentation (grapheme clusters) |
| `xo/terminfo` v0.0.0 | Reads terminfo database for terminal capabilities |
| `erikgeiser/coninput` v0.0.0 | Windows console input handling |
| `clipperhouse/uax29` v2.5.0 | Unicode word/sentence breaking (UAX #29) |
| `clipperhouse/displaywidth` v0.9.0 | Display width calculations |
| `clipperhouse/stringish` v0.1.1 | String-like type abstractions |
| `golang.org/x/sys` v0.38.0 | Low-level OS calls (terminal ioctls, signal handling) |
| `golang.org/x/text` v0.3.8 | Unicode text processing |

---

## Development Tools

These are **not** compiled into the binary. They're installed separately as Go tools for development workflow.

### golangci-lint v2.10.1

**Linter aggregator.** Runs 50+ Go linters in parallel with a single command. Catches bugs, style issues, security problems, and performance anti-patterns that `go vet` alone won't find.

```bash
golangci-lint run ./...
```

- **Install:** `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
- **Repository:** https://github.com/golangci/golangci-lint

### GoReleaser v2.8.2

**Release automation.** Builds cross-compiled binaries, creates GitHub Releases, generates changelogs, and packages archives. STUI uses it to produce `linux/amd64` and `linux/arm64` binaries that users download via the `scripts/install.sh` curl wrapper.

```bash
goreleaser release --snapshot --clean   # local test
goreleaser release                       # actual release (in CI)
```

- **Install:** `go install github.com/goreleaser/goreleaser/v2@v2.8.2`
- **Repository:** https://github.com/goreleaser/goreleaser
- **Note:** v2.8.2 pinned because v2.14+ requires Go 1.26

---

## System Requirements

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25.5+ | Build and run STUI |
| git | 2.x | Version control, beads sync |
| dolt | 1.82.6 | SQL backend for beads issue tracking |
| bd | 0.56.1 | Beads issue tracker CLI |

---

## Adding Dependencies

When adding a new dependency:

1. `go get github.com/example/pkg@latest`
2. Import and use it in code
3. `go mod tidy` to clean up
4. Update this document with the package name, version, purpose, and license
5. Commit `go.mod`, `go.sum`, and this doc together
