package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// System abstracts host-level reads and probes so installers can target
// either the local machine or a remote host in the future.
type System interface {
	ReadFile(path string) ([]byte, error)
	LookPath(name string) (string, error)
	IsRoot() bool
	DetectEscalation() *EscalationMethod
	NumCPU() int
	Statfs(path string, buf *unix.Statfs_t) error
}

type localSystem struct{}

// NewLocalSystem returns a System backed by the current machine.
func NewLocalSystem() System {
	return localSystem{}
}

func (localSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (localSystem) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (localSystem) IsRoot() bool {
	return IsRoot()
}

func (localSystem) DetectEscalation() *EscalationMethod {
	return DetectEscalation()
}

func (localSystem) NumCPU() int {
	return runtime.NumCPU()
}

func (localSystem) Statfs(path string, buf *unix.Statfs_t) error {
	return unix.Statfs(path, buf)
}

// SystemForTarget returns a host abstraction for the requested target.
func SystemForTarget(target Target) (System, error) {
	target.Normalize()
	if err := target.Validate(); err != nil {
		return nil, err
	}

	switch target.Mode {
	case TargetModeLocal:
		return NewLocalSystem(), nil
	case TargetModeSSH:
		return sshSystem{target: target}, nil
	default:
		return nil, fmt.Errorf("unsupported target mode %q", target.Mode)
	}
}

type sshSystem struct {
	target Target
}

func (s sshSystem) ReadFile(path string) ([]byte, error) {
	return s.run("cat " + shellQuote(path))
}

func (s sshSystem) LookPath(name string) (string, error) {
	out, err := s.run("command -v " + shellQuote(name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (s sshSystem) IsRoot() bool {
	_, err := s.run(`test "$(id -u)" = 0`)
	return err == nil
}

func (s sshSystem) DetectEscalation() *EscalationMethod {
	return DetectEscalationWith(s.LookPath)
}

func (s sshSystem) NumCPU() int {
	out, err := s.run("getconf _NPROCESSORS_ONLN 2>/dev/null || nproc")
	if err != nil {
		return 0
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0
	}
	return count
}

func (s sshSystem) Statfs(path string, buf *unix.Statfs_t) error {
	out, err := s.run("stat -fc '%a %S' " + shellQuote(path))
	if err != nil {
		return err
	}

	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 {
		return fmt.Errorf("unexpected statfs output: %q", strings.TrimSpace(string(out)))
	}

	bavail, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing available blocks: %w", err)
	}
	bsize, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing block size: %w", err)
	}

	buf.Bavail = bavail
	buf.Bsize = bsize
	return nil
}

func (s sshSystem) run(script string) ([]byte, error) {
	args := []string{
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		"-p", strconv.Itoa(s.target.Port),
		fmt.Sprintf("%s@%s", s.target.User, s.target.Host),
		"sh", "-lc", script,
	}

	out, err := exec.Command("ssh", args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%v: %s", err, msg)
	}
	return out, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
