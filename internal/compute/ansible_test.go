package compute

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// mockInstanceRepository is a mock implementation of InstanceRepository for testing
type mockInstanceRepository struct {
	instances []models.Instance
	err       error
}

func (m *mockInstanceRepository) GetByJobID(_ context.Context, _ uint, _ uint) ([]models.Instance, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.instances, nil
}

func TestCreateInventory(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ansible_test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	// Change to temp directory for test
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		if err != nil {
			t.Errorf("failed to change back to original directory: %v", err)
		}
	}()

	// Create ansible/inventory directory
	err = os.MkdirAll(filepath.Join("ansible", "inventory"), 0750)
	require.NoError(t, err)

	// Create test instances
	testInstances := []models.Instance{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name:     "test-1",
			PublicIP: "192.168.1.1",
			JobID:    123,
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Name:     "test-2",
			PublicIP: "192.168.1.2",
			JobID:    123,
		},
	}

	tests := []struct {
		name          string
		jobID         string
		ownerID       uint
		instances     map[string]string
		dbInstances   []models.Instance
		mockErr       error
		wantErr       bool
		useRepository bool
	}{
		{
			name:  "successful inventory creation from map",
			jobID: "job-123",
			instances: map[string]string{
				"test-1": "192.168.1.1",
				"test-2": "192.168.1.2",
			},
			wantErr: false,
		},
		{
			name:          "successful inventory creation from database",
			jobID:         "job-123",
			ownerID:       1,
			dbInstances:   testInstances,
			useRepository: true,
			wantErr:       false,
		},
		{
			name:          "database error falls back to map",
			jobID:         "job-123",
			ownerID:       1,
			instances:     map[string]string{"test-1": "192.168.1.1"},
			useRepository: true,
			mockErr:       assert.AnError,
			wantErr:       false,
		},
		{
			name:          "no instances provided and database error",
			jobID:         "job-123",
			ownerID:       1,
			instances:     nil,
			useRepository: true,
			mockErr:       assert.AnError,
			wantErr:       true,
		},
		{
			name:      "no instances provided without database",
			jobID:     "job-123",
			instances: nil,
			wantErr:   true,
		},
		{
			name:          "skip deleted instances from database",
			jobID:         "job-123",
			ownerID:       1,
			useRepository: true,
			dbInstances: []models.Instance{
				{
					Model: gorm.Model{
						ID:        1,
						DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
					},
					Name:     "test-1",
					PublicIP: "192.168.1.1",
					JobID:    123,
				},
				{
					Model: gorm.Model{
						ID: 2,
					},
					Name:     "test-2",
					PublicIP: "192.168.1.2",
					JobID:    123,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockRepo *mockInstanceRepository
			if tt.useRepository {
				mockRepo = &mockInstanceRepository{
					instances: tt.dbInstances,
					err:       tt.mockErr,
				}
			}

			// Create configurator
			cfg := NewAnsibleConfigurator(tt.jobID, mockRepo)
			if tt.useRepository {
				cfg.SetOwnerID(tt.ownerID)
			}

			// Create inventory
			err := cfg.CreateInventory(tt.instances, "")

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Check inventory file exists
			inventoryPath := filepath.Join("ansible", "inventory", "inventory_"+tt.jobID+"_ansible.ini")
			assert.FileExists(t, inventoryPath)

			// Check file permissions
			info, err := os.Stat(inventoryPath)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

			// Read and verify file contents
			content, err := os.ReadFile(filepath.Clean(inventoryPath))
			require.NoError(t, err)

			// Verify header
			assert.Contains(t, string(content), "[all:vars]")
			assert.Contains(t, string(content), "ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'")
			assert.Contains(t, string(content), "[all]")

			// Verify instances
			if tt.useRepository && tt.mockErr == nil {
				for _, instance := range tt.dbInstances {
					if !instance.DeletedAt.Valid {
						expectedLine := instance.Name + " ansible_host=" + instance.PublicIP
						assert.Contains(t, string(content), expectedLine)
					}
				}
			} else if tt.instances != nil {
				for name, ip := range tt.instances {
					expectedLine := name + " ansible_host=" + ip
					assert.Contains(t, string(content), expectedLine)
				}
			}
		})
	}
}
