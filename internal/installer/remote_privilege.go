package installer

import (
	"context"
	"fmt"
	"io"
)

// privilegedCommand wraps a command so it runs with elevated privileges when
// the target is remote and the SSH user is not already root.
func privilegedCommand(target Target, isRoot bool, esc *EscalationMethod, command string) (string, error) {
	target.Normalize()

	if isRoot {
		return command, nil
	}

	if target.Mode != TargetModeSSH {
		return "", fmt.Errorf("elevated privileges are required; relaunch STUI locally with sudo or doas")
	}

	if esc == nil {
		return "", fmt.Errorf("remote target user is not root and no sudo/doas command is available")
	}

	switch esc.Name {
	case "sudo":
		return "sudo -n sh -lc " + shellQuote(command), nil
	case "doas":
		return "doas -n sh -lc " + shellQuote(command), nil
	default:
		return "", fmt.Errorf("unsupported privilege escalation command %q", esc.Name)
	}
}

// RunPrivilegedCmd executes a command with the privileges required by the
// current target. Remote installs use non-interactive sudo/doas when needed.
func RunPrivilegedCmd(ctx context.Context, target Target, system System, command string, output io.Writer) error {
	privileged, err := privilegedCommand(target, system.IsRoot(), system.DetectEscalation(), command)
	if err != nil {
		return err
	}

	if err := system.RunCmd(ctx, privileged, output); err != nil {
		esc := system.DetectEscalation()
		if target.Mode == TargetModeSSH && !system.IsRoot() && esc != nil {
			return fmt.Errorf("%w; connect as root or configure passwordless %s on the remote host", err, esc.Name)
		}
		return err
	}

	return nil
}
