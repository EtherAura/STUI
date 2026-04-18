package installer

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// SudoPasswordRequiredError indicates the remote host needs a sudo password
// before a privileged command can continue.
type SudoPasswordRequiredError struct{}

func (e *SudoPasswordRequiredError) Error() string {
	return "remote sudo password is required"
}

// privilegedCommand wraps a command so it runs with elevated privileges when
// the target is remote and the SSH user is not already root.
func privilegedCommand(target Target, isRoot bool, esc *EscalationMethod, command string) (string, string, error) {
	target.Normalize()

	if isRoot {
		return command, "", nil
	}

	if target.Mode != TargetModeSSH {
		return "", "", fmt.Errorf("elevated privileges are required; relaunch STUI locally with sudo or doas")
	}

	if esc == nil {
		return "", "", fmt.Errorf("remote target user is not root and no sudo/doas command is available")
	}

	switch esc.Name {
	case "sudo":
		if target.SudoPassword != "" {
			return "sudo -S -p '' sh -lc " + shellQuote(command), target.SudoPassword + "\n", nil
		}
		return "", "", &SudoPasswordRequiredError{}
	case "doas":
		return "doas -n sh -lc " + shellQuote(command), "", nil
	default:
		return "", "", fmt.Errorf("unsupported privilege escalation command %q", esc.Name)
	}
}

// RunPrivilegedCmd executes a command with the privileges required by the
// current target. Remote installs use non-interactive sudo/doas when needed.
func RunPrivilegedCmd(ctx context.Context, target Target, system System, command string, output io.Writer) error {
	return RunPrivilegedCmdInput(ctx, target, system, command, "", output)
}

// RunPrivilegedCmdInput executes a privileged command and forwards extra stdin
// to the child process after any required sudo password.
func RunPrivilegedCmdInput(ctx context.Context, target Target, system System, command string, extraInput string, output io.Writer) error {
	privileged, input, err := privilegedCommand(target, system.IsRoot(), system.DetectEscalation(), command)
	if err != nil {
		return err
	}
	input += extraInput

	if err := system.RunCmdInput(ctx, privileged, input, output); err != nil {
		esc := system.DetectEscalation()
		if target.Mode == TargetModeSSH && !system.IsRoot() && esc != nil {
			if esc.Name == "sudo" && target.SudoPassword != "" {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				return fmt.Errorf("%w; the remote sudo password may be incorrect", err)
			}
			return fmt.Errorf("%w; connect as root or verify the remote %s password/permissions", err, esc.Name)
		}
		return err
	}

	return nil
}
