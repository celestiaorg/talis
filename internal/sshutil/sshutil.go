package sshutil

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// Constants for common PEM block types for private keys.
const (
	// pemBlockTypeRSA is the type for PKCS#1 RSA private keys.
	pemBlockTypeRSA = "RSA PRIVATE KEY"
	// pemBlockTypePKCS8 is the type for PKCS#8 encoded private keys (often used for Ed25519, ECDSA).
	pemBlockTypePKCS8 = "PRIVATE KEY"
	// pemBlockTypeOpenSSH is the type for OpenSSH's newer private key format.
	pemBlockTypeOpenSSH = "OPENSSH PRIVATE KEY"
)

// DerivePublicKeyFromPrivateKeyPEM takes a PEM-encoded private key and returns the corresponding public key in OpenSSH authorized_keys format.
// It supports RSA and Ed25519 keys.
func DerivePublicKeyFromPrivateKeyPEM(privateKeyPEM []byte) (string, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	var (
		key interface{}
		err error
	)

	switch block.Type {
	case pemBlockTypeRSA:
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case pemBlockTypePKCS8:
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case pemBlockTypeOpenSSH:
		key, err = ssh.ParseRawPrivateKey(privateKeyPEM)
	default:
		return "", fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	var sshPubKey ssh.PublicKey
	switch k := key.(type) {
	case *rsa.PrivateKey:
		sshPubKey, err = ssh.NewPublicKey(&k.PublicKey)
		if err != nil {
			return "", fmt.Errorf("failed to create SSH public key from RSA private key: %w", err)
		}
	case ed25519.PrivateKey:
		sshPubKey, err = ssh.NewPublicKey(k.Public())
		if err != nil {
			return "", fmt.Errorf("failed to create SSH public key from Ed25519 private key: %w", err)
		}
	// Add cases for other key types like ECDSA if needed
	default:
		return "", fmt.Errorf("unsupported parsed key type: %T", k)
	}

	// Marshal the public key into the authorized_keys format.
	authorizedKeyBytes := ssh.MarshalAuthorizedKey(sshPubKey)
	return string(authorizedKeyBytes), nil
}

// GenerateFingerprintSHA256 calculates the SHA256 fingerprint of an SSH public key.
// The input publicKeyStr should be the authorized_keys format (e.g., "ssh-ed25519 AAA...").
// Returns the fingerprint in the format "SHA256:base64encodedhash".
func GenerateFingerprintSHA256(publicKeyStr string) (string, error) {
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	return ssh.FingerprintSHA256(publicKey), nil
}
