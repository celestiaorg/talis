package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/test"
)

func TestSSHKeyRPCMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Test parameters
	ownerID := models.AdminID // Use admin ID for testing
	testKey1 := handlers.SSHKeyCreateParams{
		Name:      "test-key-1",
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDEtest",
		OwnerID:   ownerID,
	}
	testKey2 := handlers.SSHKeyCreateParams{
		Name:      "test-key-2",
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDEtest2",
		OwnerID:   ownerID,
	}

	t.Run("ValidateSSHPublicKey", func(t *testing.T) {
		// Test valid keys
		validKeys := []string{
			"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDEtest",
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMHKBCIlkZhv8uEHRyW9n9NExJCPE1mHT2a6gVrUfSbR",
			"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBAHJjABF9xRBMB7dZ0y2xc/OyX9bMHPLQJgXHEQwCVyRNNhCHv32hV9m87VjfJK5lm0dPcIpYEtUZbC/Ot+MZEk=",
		}

		for _, key := range validKeys {
			err := handlers.ValidateSSHPublicKey(key)
			require.NoError(t, err, "Valid key should not return an error: %s", key)
		}

		// Test invalid keys
		invalidKeys := []string{
			"invalid-key",
			"rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDEtest",
			"ssh-invalid AAAAB3NzaC1yc2EAAAADAQABAAABgQDEtest",
			"just some random text",
			"", // Empty key
		}

		for _, key := range invalidKeys {
			err := handlers.ValidateSSHPublicKey(key)
			require.Error(t, err, "Invalid key should return an error: %s", key)
		}
	})

	t.Run("CreateSSHKey_Success", func(t *testing.T) {
		// Create a new SSH key
		key, err := suite.APIClient.CreateSSHKey(suite.Context(), testKey1)
		require.NoError(t, err)
		require.Equal(t, testKey1.Name, key.Name)
		require.Equal(t, testKey1.PublicKey, key.PublicKey)
		require.Equal(t, testKey1.OwnerID, key.OwnerID)
	})

	t.Run("CreateSSHKey_InvalidFormat", func(t *testing.T) {
		// Try to create a key with invalid format
		invalidKey := handlers.SSHKeyCreateParams{
			Name:      "invalid-key",
			PublicKey: "invalid-key-format",
			OwnerID:   ownerID,
		}
		_, err := suite.APIClient.CreateSSHKey(suite.Context(), invalidKey)
		require.Error(t, err, "Creating key with invalid format should fail")
		require.Contains(t, err.Error(), "invalid SSH public key format")
	})

	t.Run("ListSSHKeys_Success", func(t *testing.T) {
		// Create a second key to test listing multiple keys
		_, err := suite.APIClient.CreateSSHKey(suite.Context(), testKey2)
		require.NoError(t, err)

		// List SSH keys
		keys, err := suite.APIClient.ListSSHKeys(suite.Context(), handlers.SSHKeyListParams{
			OwnerID: ownerID,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(keys), 2)

		// Verify both keys are in the list
		foundKey1 := false
		foundKey2 := false
		for _, key := range keys {
			if key.Name == testKey1.Name {
				foundKey1 = true
				require.Equal(t, testKey1.PublicKey, key.PublicKey)
			}
			if key.Name == testKey2.Name {
				foundKey2 = true
				require.Equal(t, testKey2.PublicKey, key.PublicKey)
			}
		}
		require.True(t, foundKey1, "First key not found in list")
		require.True(t, foundKey2, "Second key not found in list")
	})

	t.Run("DeleteSSHKey_Success", func(t *testing.T) {
		// Delete the first key
		err := suite.APIClient.DeleteSSHKey(suite.Context(), handlers.SSHKeyDeleteParams{
			Name:    testKey1.Name,
			OwnerID: ownerID,
		})
		require.NoError(t, err)

		// List keys again to verify it was deleted
		keys, err := suite.APIClient.ListSSHKeys(suite.Context(), handlers.SSHKeyListParams{
			OwnerID: ownerID,
		})
		require.NoError(t, err)

		// Check key is no longer in the list
		for _, key := range keys {
			require.NotEqual(t, testKey1.Name, key.Name, "Key should have been deleted but is still in the list")
		}
	})

	t.Run("DeleteSSHKey_NotFound", func(t *testing.T) {
		// Try to delete a non-existent key
		nonExistentKeyName := "non-existent-key"
		err := suite.APIClient.DeleteSSHKey(suite.Context(), handlers.SSHKeyDeleteParams{
			Name:    nonExistentKeyName,
			OwnerID: ownerID,
		})
		require.Error(t, err)
		// Check for the presence of key components in the error message
		errMsg := err.Error()
		require.Contains(t, errMsg, "not found", "Error should indicate the key was not found")
		require.Contains(t, errMsg, nonExistentKeyName, "Error should mention the key name")
	})

	t.Run("CreateSSHKey_DuplicateName", func(t *testing.T) {
		// Create a key with the same name as an existing key
		duplicateKey := handlers.SSHKeyCreateParams{
			Name:      testKey2.Name, // Already exists
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDEduplicate",
			OwnerID:   ownerID,
		}
		_, err := suite.APIClient.CreateSSHKey(suite.Context(), duplicateKey)
		require.Error(t, err)
	})
}
