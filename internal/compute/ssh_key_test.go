package compute

import (
	"context"
	"testing"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of repos.UserRepository
type MockUserRepository struct {
	mock.Mock
}

var _ repos.UserRepository = (*MockUserRepository)(nil) // Ensure interface compliance

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CreateBatch(ctx context.Context, users []*models.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockUserRepository) GetUsers(ctx context.Context, opts *models.ListOptions) ([]models.User, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestSSHKeyManager_GetSSHKey(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		keyName    string
		mockUser   *models.User
		mockError  error
		wantKey    string
		wantErrMsg string
	}{
		{
			name:    "successful key retrieval",
			userID:  1,
			keyName: "test-key",
			mockUser: &models.User{
				Username:     "test-user",
				PublicSSHKey: "ssh-rsa AAAA...",
			},
			mockError:  nil,
			wantKey:    "ssh-rsa AAAA...",
			wantErrMsg: "",
		},
		{
			name:    "user has no key",
			userID:  1,
			keyName: "test-key",
			mockUser: &models.User{
				Username:     "test-user",
				PublicSSHKey: "",
			},
			mockError:  nil,
			wantKey:    "",
			wantErrMsg: "user has no SSH key configured",
		},
		{
			name:       "user not found",
			userID:     1,
			keyName:    "test-key",
			mockUser:   nil,
			mockError:  gorm.ErrRecordNotFound,
			wantKey:    "",
			wantErrMsg: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockRepo.On("GetUserByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockError)

			manager := NewSSHKeyManager(mockRepo)
			key, err := manager.GetSSHKey(context.Background(), tt.userID, tt.keyName)

			if tt.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantKey, key)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSSHKeyManager_GetSSHKeys(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		mockUser   *models.User
		mockError  error
		wantKeys   []string
		wantErrMsg string
	}{
		{
			name:   "successful keys retrieval",
			userID: 1,
			mockUser: &models.User{
				Username:     "test-user",
				PublicSSHKey: "ssh-rsa AAAA...",
			},
			mockError:  nil,
			wantKeys:   []string{"ssh-rsa AAAA..."},
			wantErrMsg: "",
		},
		{
			name:   "user has no keys",
			userID: 1,
			mockUser: &models.User{
				Username:     "test-user",
				PublicSSHKey: "",
			},
			mockError:  nil,
			wantKeys:   []string{},
			wantErrMsg: "",
		},
		{
			name:       "user not found",
			userID:     1,
			mockUser:   nil,
			mockError:  gorm.ErrRecordNotFound,
			wantKeys:   nil,
			wantErrMsg: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockRepo.On("GetUserByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockError)

			manager := NewSSHKeyManager(mockRepo)
			keys, err := manager.GetSSHKeys(context.Background(), tt.userID)

			if tt.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantKeys, keys)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSSHKeyManager_ValidateSSHKey(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		keyName    string
		mockUser   *models.User
		mockError  error
		wantErrMsg string
	}{
		{
			name:    "valid key",
			userID:  1,
			keyName: "test-key",
			mockUser: &models.User{
				Username:     "test-user",
				PublicSSHKey: "ssh-rsa AAAA...",
			},
			mockError:  nil,
			wantErrMsg: "",
		},
		{
			name:    "user has no key",
			userID:  1,
			keyName: "test-key",
			mockUser: &models.User{
				Username:     "test-user",
				PublicSSHKey: "",
			},
			mockError:  nil,
			wantErrMsg: "user has no SSH key configured",
		},
		{
			name:       "user not found",
			userID:     1,
			keyName:    "test-key",
			mockUser:   nil,
			mockError:  gorm.ErrRecordNotFound,
			wantErrMsg: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockRepo.On("GetUserByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockError)

			manager := NewSSHKeyManager(mockRepo)
			err := manager.ValidateSSHKey(context.Background(), tt.userID, tt.keyName)

			if tt.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
