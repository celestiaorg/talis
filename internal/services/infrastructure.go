// Package services provides business logic implementation for the API
package services

import (
	"fmt"
	"os"

	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/logger"
)

// GetTalisServerSSHKey returns the SSH key content from the environment variable
func GetTalisServerSSHKey() (string, error) {
	// Get the key content from the environment variable
	keyContent := os.Getenv(constants.EnvTalisSSHKey)
	if keyContent == "" {
		return "", fmt.Errorf("SSH key not found in environment variable %s", constants.EnvTalisSSHKey)
	}

	logger.Debug("Using SSH key content from environment variable")
	return keyContent, nil
}

/*
// inferSSHKeyType examines the key content to determine its type
// This can be used if we need to infer the key type from the key content
func inferSSHKeyType(keyContent string) string {
	keyContent = strings.TrimSpace(keyContent)

	// Check key type based on content
	if strings.HasPrefix(keyContent, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		// Modern OpenSSH format, need to check inside content
		if strings.Contains(keyContent, "ssh-ed25519") {
			return "ed25519"
		} else if strings.Contains(keyContent, "ecdsa") {
			return "ecdsa"
		}
		// Default to RSA if can't determine specifically
		return "rsa"
	} else if strings.HasPrefix(keyContent, "-----BEGIN RSA PRIVATE KEY-----") {
		return "rsa"
	} else if strings.HasPrefix(keyContent, "-----BEGIN EC PRIVATE KEY-----") {
		return "ecdsa"
	} else if strings.HasPrefix(keyContent, "-----BEGIN DSA PRIVATE KEY-----") {
		return "dsa"
	}

	// Default to RSA if we can't determine
	return "rsa"
}
*/
