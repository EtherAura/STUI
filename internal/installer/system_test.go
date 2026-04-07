package installer

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestKnownHostsCallbackAcceptsUnknownHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	callback, err := knownHostsCallback(path)
	if err != nil {
		t.Fatalf("knownHostsCallback() error = %v", err)
	}

	key := testPublicKey(t)
	err = callback("192.0.2.10:22", &net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 22}, key)
	if err != nil {
		t.Fatalf("callback() error = %v, want nil for unknown host", err)
	}
}

func TestKnownHostsCallbackRejectsMismatchedHostKey(t *testing.T) {
	knownKey := testPublicKey(t)
	path := filepath.Join(t.TempDir(), "known_hosts")
	line := knownhosts.Line([]string{"192.0.2.10:22"}, knownKey) + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	callback, err := knownHostsCallback(path)
	if err != nil {
		t.Fatalf("knownHostsCallback() error = %v", err)
	}

	err = callback("192.0.2.10:22", &net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 22}, testPublicKey(t))
	if err == nil {
		t.Fatal("callback() error = nil, want mismatch failure")
	}

	var keyErr *knownhosts.KeyError
	if !errors.As(err, &keyErr) || len(keyErr.Want) == 0 {
		t.Fatalf("callback() error = %v, want known_hosts mismatch", err)
	}
}

func testPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("NewSignerFromKey() error = %v", err)
	}

	return signer.PublicKey()
}
