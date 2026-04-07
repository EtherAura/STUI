package installer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
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
	// RunCmd executes a shell command, streaming combined stdout/stderr
	// to output. The command is passed to sh -c for local targets or
	// sh -lc for SSH targets. Context cancellation terminates the command.
	RunCmd(ctx context.Context, command string, output io.Writer) error
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

// RunCmd executes a shell command locally via sh -c, streaming
// combined stdout and stderr to the provided writer.
func (localSystem) RunCmd(ctx context.Context, command string, output io.Writer) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = output
	cmd.Stderr = output
	return cmd.Run()
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

// RunCmd executes a shell command on the remote host over SSH,
// streaming combined stdout and stderr to the provided writer.
func (s sshSystem) RunCmd(ctx context.Context, command string, output io.Writer) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("starting ssh session: %w", err)
	}
	defer func() { _ = session.Close() }()

	session.Stdout = output
	session.Stderr = output

	done := make(chan error, 1)
	go func() {
		done <- session.Run("sh -lc " + shellQuote(command))
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		_ = session.Close()
		return ctx.Err()
	case err := <-done:
		return err
	}
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
	client, err := s.client()
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("starting ssh session: %w", err)
	}
	defer func() { _ = session.Close() }()

	out, err := session.CombinedOutput("sh -lc " + shellQuote(script))
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, msg)
	}
	return out, nil
}

func (s sshSystem) client() (*ssh.Client, error) {
	config, err := s.clientConfig()
	if err != nil {
		return nil, err
	}

	address := net.JoinHostPort(s.target.Host, strconv.Itoa(s.target.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connecting to ssh target: %w", err)
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ssh handshake failed: %w", err)
	}

	return ssh.NewClient(clientConn, chans, reqs), nil
}

func (s sshSystem) clientConfig() (*ssh.ClientConfig, error) {
	auth, err := s.authMethods()
	if err != nil {
		return nil, err
	}
	if len(auth) == 0 {
		return nil, fmt.Errorf("no ssh authentication method available")
	}

	return &ssh.ClientConfig{
		User:            s.target.User,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback(),
		Timeout:         10 * time.Second,
	}, nil
}

func (s sshSystem) authMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if s.target.Password != "" {
		methods = append(methods, ssh.Password(s.target.Password))
	}

	if s.target.KeyPath != "" {
		signer, err := signerFromPath(s.target.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("loading ssh key: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if len(methods) > 0 {
		return methods, nil
	}

	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	for _, candidate := range []string{"~/.ssh/id_ed25519", "~/.ssh/id_rsa"} {
		signer, err := signerFromPath(candidate)
		if err == nil {
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	return methods, nil
}

func signerFromPath(path string) (ssh.Signer, error) {
	expanded, err := expandPath(path)
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(expanded)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(key)
}

func hostKeyCallback() ssh.HostKeyCallback {
	home, err := os.UserHomeDir()
	if err == nil {
		knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
		if _, statErr := os.Stat(knownHostsPath); statErr == nil {
			if callback, callbackErr := knownHostsCallback(knownHostsPath); callbackErr == nil {
				return callback
			}
		}
	}
	return ssh.InsecureIgnoreHostKey()
}

func knownHostsCallback(path string) (ssh.HostKeyCallback, error) {
	callback, err := knownhosts.New(path)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := callback(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			return nil
		}

		return err
	}, nil
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
