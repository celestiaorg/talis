package provisioner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/celestiaorg/talis/internal/provisioner/config"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Also add a test for SSH failure scenario
func TestAnsibleProvisioner_Configure_SSHFailure(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ansible_test")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	// Create test config
	cfg := &config.AnsibleConfig{
		JobID:             "test-job-ssh-fail",
		OwnerID:           1,
		SSHUser:           DefaultSSHUser,
		SSHKeyPath:        DefaultSSHKeyPath,
		PlaybookPath:      filepath.Join(tmpDir, "playbook.yml"),
		InventoryBasePath: tmpDir,
	}

	// Create provisioner with mock SSH checker
	provisioner, err := NewAnsibleProvisioner(cfg)
	require.NoError(t, err)

	// Replace the default SSH checker with our mock that simulates failure
	mockSSHChecker := &MockSSHChecker{
		WaitForSSHFunc: func(host string) error {
			return fmt.Errorf("simulated SSH failure for host %s", host)
		},
	}
	provisioner.sshChecker = mockSSHChecker

	// Test instances
	instances := []types.InstanceInfo{
		{
			Name:     "test-instance-1",
			PublicIP: "1.2.3.4",
		},
	}

	// Configure instances - should fail due to SSH error
	err = provisioner.Configure(context.Background(), instances)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated SSH failure for host 1.2.3.4")
}
