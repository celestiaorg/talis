// Package services provides business logic implementation for the API
package services

import (
	"os"
	"strings"

	"github.com/celestiaorg/talis/internal/types"
)

// getAnsibleSSHKeyPath determines the appropriate SSH private key path for Ansible
// based on the instance requests, prioritizing custom paths, then key types,
// and falling back to the default RSA key path.
// It assumes the key configuration from the first request applies to the whole job.
func getAnsibleSSHKeyPath(instanceRequest types.InstanceRequest) string {
	sshKeyPath := "$HOME/.ssh/id_rsa"     // Default
	if instanceRequest.SSHKeyPath != "" { // Priority 1: Custom path
		sshKeyPath = instanceRequest.SSHKeyPath
	} else if instanceRequest.SSHKeyType != "" { // Priority 2: Key type
		switch strings.ToLower(instanceRequest.SSHKeyType) {
		case "ed25519":
			sshKeyPath = "$HOME/.ssh/id_ed25519"
		case "ecdsa":
			sshKeyPath = "$HOME/.ssh/id_ecdsa"
			// Add other types if needed
		}
	}
	// Expand environment variables like $HOME (Ansible handles this, but doing it here is safe)
	return os.ExpandEnv(sshKeyPath)
}
