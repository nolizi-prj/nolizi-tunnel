package relay

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// GenerateHostKeyPEM makes a fresh ed25519 host key for the relay's ssh
// ingress. ed25519 because every current ssh client accepts it and the key is
// small; the relay only ever needs one.
func GenerateHostKeyPEM() ([]byte, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	block, err := ssh.MarshalPrivateKey(priv, "pumasi-relay ssh host key")
	if err != nil {
		return nil, fmt.Errorf("relay: marshalling host key: %w", err)
	}
	return pem.EncodeToMemory(block), nil
}
