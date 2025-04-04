package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserRole(t *testing.T) {
	tests := []struct {
		name          string
		role          UserRole
		stringValue   string
		validForParse bool
		roleIndex     int
	}{
		{
			name:          "User role",
			role:          UserRoleUser,
			stringValue:   "user",
			validForParse: true,
			roleIndex:     0,
		},
		{
			name:          "Admin role",
			role:          UserRoleAdmin,
			stringValue:   "admin",
			validForParse: true,
			roleIndex:     1,
		},
		{
			name:          "Invalid role",
			stringValue:   "invalid_role",
			validForParse: false,
			roleIndex:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String() method if we have a valid role
			if tt.roleIndex >= 0 {
				assert.Equal(t, tt.stringValue, tt.role.String(), "String() method failed")
				// Verify the role index matches the iota value
				assert.Equal(t, tt.roleIndex, int(tt.role), "Role index does not match expected iota value")
			}

			// Test ParseUserRole
			parsedRole, err := ParseUserRole(tt.stringValue)
			if tt.validForParse {
				assert.NoError(t, err, "ParseUserRole should not return error")
				assert.Equal(t, tt.role, parsedRole, "ParseUserRole returned wrong role")
			} else {
				assert.Error(t, err, "ParseUserRole should return error for invalid role")
				assert.Equal(t, UserRoleUser, parsedRole, "Invalid role should return UserRoleUser")
			}
		})
	}
}

func TestUser_Validation(t *testing.T) {
	timeStr := "2024-03-25T12:00:00Z"
	validUser := User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         UserRoleAdmin,
		PublicSSHKey: "ssh-rsa AAAAB3NzaC1yc2E...",
		CreatedAt:    timeStr,
		UpdatedAt:    timeStr,
	}

	t.Run("Valid user", func(t *testing.T) {
		jsonData, err := json.Marshal(validUser)
		assert.NoError(t, err)

		var unmarshaledUser User
		err = json.Unmarshal(jsonData, &unmarshaledUser)
		assert.NoError(t, err)

		// Verify fields were correctly marshaled/unmarshaled
		// Don't check ID as it's not part of the JSON output
		assert.Equal(t, validUser.Username, unmarshaledUser.Username)
		assert.Equal(t, validUser.Email, unmarshaledUser.Email)
		assert.Equal(t, validUser.Role, unmarshaledUser.Role)
		assert.Equal(t, validUser.PublicSSHKey, unmarshaledUser.PublicSSHKey)
		assert.Equal(t, validUser.CreatedAt, unmarshaledUser.CreatedAt)
		assert.Equal(t, validUser.UpdatedAt, unmarshaledUser.UpdatedAt)
	})

	t.Run("User role validation", func(t *testing.T) {
		// Test that roles are within valid range
		assert.GreaterOrEqual(t, int(UserRoleUser), 0, "User role should be non-negative")
		assert.GreaterOrEqual(t, int(UserRoleAdmin), 0, "Admin role should be non-negative")
		assert.Less(t, int(UserRoleUser), 2, "User role should be less than number of roles")
		assert.Less(t, int(UserRoleAdmin), 2, "Admin role should be less than number of roles")

		// Test that roles are ordered correctly (User = 0, Admin = 1)
		assert.Equal(t, 0, int(UserRoleUser), "UserRoleUser should be 0")
		assert.Equal(t, 1, int(UserRoleAdmin), "UserRoleAdmin should be 1")
	})

	t.Run("Username validation", func(t *testing.T) {
		// Test empty username
		invalidUser := validUser
		invalidUser.Username = ""
		_, err := json.Marshal(invalidUser)
		assert.NoError(t, err, "Should be able to marshal user with empty username")

		// Test very long username
		invalidUser.Username = "very_long_username_that_might_exceed_database_limits_very_long_username_that_might_exceed_database_limits"
		_, err = json.Marshal(invalidUser)
		assert.NoError(t, err, "Should be able to marshal user with long username")
	})

	t.Run("Email validation", func(t *testing.T) {
		// Test empty email
		invalidUser := validUser
		invalidUser.Email = ""
		_, err := json.Marshal(invalidUser)
		assert.NoError(t, err, "Should be able to marshal user with empty email")

		// Test invalid email format
		invalidUser.Email = "invalid_email"
		_, err = json.Marshal(invalidUser)
		assert.NoError(t, err, "Should be able to marshal user with invalid email format")
	})

	t.Run("SSH key validation", func(t *testing.T) {
		// Test empty SSH key
		invalidUser := validUser
		invalidUser.PublicSSHKey = ""
		_, err := json.Marshal(invalidUser)
		assert.NoError(t, err, "Should be able to marshal user with empty SSH key")

		// Test invalid SSH key format
		invalidUser.PublicSSHKey = "invalid_ssh_key"
		_, err = json.Marshal(invalidUser)
		assert.NoError(t, err, "Should be able to marshal user with invalid SSH key format")
	})
}
