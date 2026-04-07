package installer

import (
	"os"
	"os/exec"
	"runtime"

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
