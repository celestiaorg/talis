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

func TestAnsibleProvisioner_Configure(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ansible_test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create test config
	cfg := &config.AnsibleConfig{
		JobID:   "test-job",
		OwnerID: 1,
	}

	// Create provisioner
	provisioner, err := NewAnsibleProvisioner(cfg)
	require.NoError(t, err)

	// Test instances
	instances := []types.InstanceInfo{
		{
			Name:     "test-instance-1",
			PublicIP: "1.2.3.4",
		},
		{
			Name:     "test-instance-2",
			PublicIP: "5.6.7.8",
		},
	}

	// Configure instances
	err = provisioner.Configure(context.Background(), instances)
	assert.NoError(t, err)

	// Verify inventory file was created
	inventoryPath := filepath.Join("ansible", "inventory", fmt.Sprintf("inventory_%s_ansible.ini", cfg.JobID))
	_, err = os.Stat(inventoryPath)
	assert.NoError(t, err)
}
