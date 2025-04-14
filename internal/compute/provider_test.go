package compute

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnsibleProvisioner(t *testing.T) {
	// Create temporary test directories
	tmpDir := t.TempDir()

	// Create ansible directory structure
	ansibleDir := filepath.Join(tmpDir, "ansible")
	require.NoError(t, os.MkdirAll(ansibleDir, 0750))

	// Create dummy playbook
	playbookPath := filepath.Join(ansibleDir, "main.yml")
	require.NoError(t, os.WriteFile(playbookPath, []byte("---\n"), 0600))

	// Create inventory directory
	inventoryDir := filepath.Join(ansibleDir, "inventory")
	require.NoError(t, os.MkdirAll(inventoryDir, 0750))

	// Create dummy SSH key
	sshDir := filepath.Join(tmpDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("dummy key"), 0600))

	// Set up environment for testing
	t.Setenv("HOME", tmpDir)
	t.Setenv("ANSIBLE_PLAYBOOK_PATH", playbookPath)
	t.Setenv("ANSIBLE_INVENTORY_PATH", inventoryDir)

	tests := []struct {
		name    string
		jobID   string
		wantErr bool
	}{
		{
			name:    "valid job ID",
			jobID:   "job-20250411-134559",
			wantErr: false,
		},
		{
			name:    "empty job ID",
			jobID:   "",
			wantErr: true,
		},
		{
			name:    "invalid job ID format",
			jobID:   "invalid-format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provisioner, err := NewAnsibleProvisioner(tt.jobID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provisioner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provisioner)
			}
		})
	}
}
