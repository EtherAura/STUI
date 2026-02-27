// main_test.go contains unit tests for the CLI entrypoint helpers,
// including environment sanitization for privilege escalation.
package main

import (
	"testing"
)

// TestSanitizeEnv verifies that sanitizeEnv filters environment variables
// to the allowlist, blocking sensitive entries like cloud credentials.
func TestSanitizeEnv(t *testing.T) {
	input := []string{
		"PATH=/usr/bin:/bin",
		"HOME=/home/user",
		"TERM=xterm-256color",
		"LANG=en_US.UTF-8",
		"LC_ALL=en_US.UTF-8",
		"AWS_SECRET_ACCESS_KEY=supersecret",
		"GITHUB_TOKEN=ghp_xxx",
		"DATABASE_URL=postgres://user:pass@host/db",
		"SHELL=/bin/bash",
		"DISPLAY=:0",
		"XDG_RUNTIME_DIR=/run/user/1000",
		"COLORTERM=truecolor",
		"LOGNAME=carl",
		"USER=carl",
	}

	result := sanitizeEnv(input)

	// Build a set of the filtered results for easy lookup.
	filtered := make(map[string]bool)
	for _, e := range result {
		filtered[e] = true
	}

	// These should be kept.
	kept := []string{
		"PATH=/usr/bin:/bin",
		"HOME=/home/user",
		"TERM=xterm-256color",
		"LANG=en_US.UTF-8",
		"LC_ALL=en_US.UTF-8",
		"SHELL=/bin/bash",
		"DISPLAY=:0",
		"XDG_RUNTIME_DIR=/run/user/1000",
		"COLORTERM=truecolor",
		"LOGNAME=carl",
		"USER=carl",
	}
	for _, k := range kept {
		if !filtered[k] {
			t.Errorf("expected %q to be kept, but it was filtered out", k)
		}
	}

	// These should be removed.
	removed := []string{
		"AWS_SECRET_ACCESS_KEY=supersecret",
		"GITHUB_TOKEN=ghp_xxx",
		"DATABASE_URL=postgres://user:pass@host/db",
	}
	for _, r := range removed {
		if filtered[r] {
			t.Errorf("expected %q to be removed, but it was kept", r)
		}
	}
}

// TestSanitizeEnvEmpty verifies sanitizeEnv handles an empty input.
func TestSanitizeEnvEmpty(t *testing.T) {
	result := sanitizeEnv(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}
