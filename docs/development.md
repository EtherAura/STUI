# Development

## Prerequisites

- Go 1.25+ (`go version` to check)
- git
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

## Building

```bash
go build -o stui ./cmd/stui
```

This produces a `./stui` binary in the project root. You can also run without building:

```bash
go run ./cmd/stui
```

## Running

```bash
./stui
```

## Testing

Run all tests:

```bash
go test ./...
```

Run with verbose output:

```bash
go test ./... -v
```

Generate a coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Linting

```bash
golangci-lint run ./...
```

Configuration lives in `.golangci.yml` at the project root.

## Project Structure

```
cmd/stui/          — CLI entrypoint
internal/tui/      — Bubble Tea models and views
internal/installer/ — Per-app install logic
internal/config/   — Configuration management
scripts/           — Distribution install script
docs/              — This documentation
```

## Adding a New Installer

1. Create a new file in `internal/installer/` (e.g., `newapp.go`)
2. Implement the `Installer` interface
3. Register it in the app selection menu in `internal/tui/menu.go`
4. Add documentation in `docs/installers.md`
5. Create a beads issue: `bd create "Add newapp installer" -t feature -p 2`

## Release Process

Releases are built via GoReleaser and published to GitHub Releases. The `scripts/install.sh` script downloads the correct binary for the user's platform.

## Code Style

- Run `gofmt` before committing
- Use `golangci-lint run ./...` for additional checks
- Wrap errors with context: `fmt.Errorf("doing X: %%w", err)`
- Follow the **Comment Standards** in `.github/copilot-instructions.md` — every function, type, and file must be commented
