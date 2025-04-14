package provisioner

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/celestiaorg/talis/internal/logger"
)

// SSHChecker defines the interface for checking SSH connectivity
type SSHChecker interface {
	WaitForSSH(host string) error
}

// DefaultSSHChecker implements the default SSH checking behavior
type DefaultSSHChecker struct {
	sshUser    string
	sshKeyPath string
	maxRetries int
	retryDelay time.Duration
}

// NewDefaultSSHChecker creates a new DefaultSSHChecker
func NewDefaultSSHChecker(sshUser, sshKeyPath string) *DefaultSSHChecker {
	return &DefaultSSHChecker{
		sshUser:    sshUser,
		sshKeyPath: sshKeyPath,
		maxRetries: 30,
		retryDelay: 10 * time.Second,
	}
}

// WaitForSSH implements SSHChecker.WaitForSSH
func (c *DefaultSSHChecker) WaitForSSH(host string) error {
	logger.Infof("⏳ Waiting for SSH to be available on %s...", host)

	for i := 0; i < c.maxRetries; i++ {
		args := []string{
			"-i", c.sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("%s@%s", c.sshUser, host),
			"echo 'SSH is ready'",
		}

		// #nosec G204 -- command arguments are constructed from validated inputs
		checkCmd := exec.Command("ssh", args...)
		if err := checkCmd.Run(); err == nil {
			logger.Infof("✅ SSH connection established to %s", host)
			return nil
		}

		if i == c.maxRetries-1 {
			return fmt.Errorf("timeout waiting for SSH to be ready on %s after 5 minutes", host)
		}

		logger.Infof("  Retrying SSH connection to %s in 10 seconds... (%d/%d)", host, i+1, c.maxRetries)
		time.Sleep(c.retryDelay)
	}

	return nil
}
