package compute

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// DropletService defines the interface for droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
	ListByTag(ctx context.Context, tag string, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	ListByName(ctx context.Context, name string, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	ListWithGPUs(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	Actions(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Action, *godo.Response, error)
	Backups(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Image, *godo.Response, error)
	Kernels(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Kernel, *godo.Response, error)
	Snapshots(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Image, *godo.Response, error)
	Neighbors(ctx context.Context, dropletID int) ([]godo.Droplet, *godo.Response, error)
	ActionByID(ctx context.Context, dropletID int, actionID int) (*godo.Action, *godo.Response, error)
	ActionByTag(ctx context.Context, tag string, actionID int) ([]godo.Action, *godo.Response, error)
	DeleteByTag(ctx context.Context, tag string) (*godo.Response, error)
	GetBackupPolicy(ctx context.Context, dropletID int) (*godo.DropletBackupPolicy, *godo.Response, error)
	ListBackupPolicies(ctx context.Context, opt *godo.ListOptions) (map[int]*godo.DropletBackupPolicy, *godo.Response, error)
	ListSupportedBackupPolicies(ctx context.Context) ([]*godo.SupportedBackupPolicy, *godo.Response, error)
}

// KeyService defines the interface for SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Key, *godo.Response, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*godo.Key, *godo.Response, error)
	GetByID(ctx context.Context, id int) (*godo.Key, *godo.Response, error)
	Create(ctx context.Context, createRequest *godo.KeyCreateRequest) (*godo.Key, *godo.Response, error)
	UpdateByID(ctx context.Context, id int, updateRequest *godo.KeyUpdateRequest) (*godo.Key, *godo.Response, error)
	UpdateByFingerprint(ctx context.Context, fingerprint string, updateRequest *godo.KeyUpdateRequest) (*godo.Key, *godo.Response, error)
	DeleteByID(ctx context.Context, id int) (*godo.Response, error)
	DeleteByFingerprint(ctx context.Context, fingerprint string) (*godo.Response, error)
}

// DOClient defines the interface for DigitalOcean client operations
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
}

// mockDOClient implements the DOClient interface
type mockDOClient struct {
	mockDropletService *mockDropletService
	mockKeyService     *mockKeyService
}

// NewMockDOClient creates a new mock DigitalOcean client
func NewMockDOClient() DOClient {
	return &mockDOClient{
		mockDropletService: &mockDropletService{
			droplets: make(map[int]*godo.Droplet),
		},
		mockKeyService: &mockKeyService{
			keys: make(map[int]*godo.Key),
		},
	}
}

// Droplets returns the mock droplet service
func (c *mockDOClient) Droplets() DropletService {
	return c.mockDropletService
}

// Keys returns the mock key service
func (c *mockDOClient) Keys() KeyService {
	return c.mockKeyService
}

// mockDropletService implements the DropletService interface
type mockDropletService struct {
	droplets                        map[int]*godo.Droplet
	CreateFunc                      func(context.Context, *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	DeleteFunc                      func(context.Context, int) (*godo.Response, error)
	GetFunc                         func(context.Context, int) (*godo.Droplet, *godo.Response, error)
	ListFunc                        func(context.Context, *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	ListByNameFunc                  func(context.Context, string, *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	ListWithGPUsFunc                func(context.Context, *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	GetBackupPolicyFunc             func(context.Context, int) (*godo.DropletBackupPolicy, *godo.Response, error)
	ListBackupPoliciesFunc          func(context.Context, *godo.ListOptions) (map[int]*godo.DropletBackupPolicy, *godo.Response, error)
	ListSupportedBackupPoliciesFunc func(context.Context) ([]*godo.SupportedBackupPolicy, *godo.Response, error)
}

// Create mocks the Create method of DropletService
func (m *mockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, createRequest)
	}

	if m.droplets == nil {
		return nil, nil, fmt.Errorf("droplet service not initialized")
	}

	droplet := &godo.Droplet{
		ID:     12345, // Use the expected ID from the test
		Name:   createRequest.Name,
		Region: &godo.Region{Slug: createRequest.Region},
		Size:   &godo.Size{Slug: createRequest.Size},
		Image:  &godo.Image{Slug: createRequest.Image.Slug},
	}
	m.droplets[droplet.ID] = droplet
	return droplet, &godo.Response{}, nil
}

// CreateMultiple mocks the CreateMultiple method of DropletService
func (m *mockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	if m.droplets == nil {
		return nil, nil, fmt.Errorf("droplet service not initialized")
	}

	var droplets []godo.Droplet
	for i, name := range createRequest.Names {
		droplet := godo.Droplet{
			ID:     10000 + i,
			Name:   name,
			Region: &godo.Region{Slug: createRequest.Region},
			Size:   &godo.Size{Slug: createRequest.Size},
			Image:  &godo.Image{Slug: createRequest.Image.Slug},
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: fmt.Sprintf("192.0.2.%d", i+1),
					},
				},
			},
		}
		m.droplets[droplet.ID] = &droplet
		droplets = append(droplets, droplet)
	}
	return droplets, &godo.Response{}, nil
}

// Get mocks the Get method of DropletService
func (m *mockDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	if droplet, ok := m.droplets[id]; ok {
		return droplet, &godo.Response{}, nil
	}
	return nil, &godo.Response{}, fmt.Errorf("droplet not found")
}

// Delete mocks the Delete method of DropletService
func (m *mockDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	delete(m.droplets, id)
	return &godo.Response{}, nil
}

// List mocks the List method of DropletService
func (m *mockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	if m.droplets == nil {
		return nil, nil, fmt.Errorf("droplet service not initialized")
	}

	var droplets []godo.Droplet
	for _, droplet := range m.droplets {
		droplets = append(droplets, *droplet)
	}
	return droplets, &godo.Response{}, nil
}

// ListByTag mocks the ListByTag method of DropletService
func (m *mockDropletService) ListByTag(ctx context.Context, tag string, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// ListByName mocks the ListByName method of DropletService
func (m *mockDropletService) ListByName(ctx context.Context, name string, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	if m.ListByNameFunc != nil {
		return m.ListByNameFunc(ctx, name, opt)
	}
	return nil, nil, nil
}

// ListWithGPUs mocks the ListWithGPUs method of DropletService
func (m *mockDropletService) ListWithGPUs(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	if m.ListWithGPUsFunc != nil {
		return m.ListWithGPUsFunc(ctx, opt)
	}
	return nil, nil, nil
}

// Actions mocks the Actions method of DropletService
func (m *mockDropletService) Actions(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Action, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// Backups mocks the Backups method of DropletService
func (m *mockDropletService) Backups(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Image, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// Kernels mocks the Kernels method of DropletService
func (m *mockDropletService) Kernels(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Kernel, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// Snapshots mocks the Snapshots method of DropletService
func (m *mockDropletService) Snapshots(ctx context.Context, dropletID int, opt *godo.ListOptions) ([]godo.Image, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// Neighbors mocks the Neighbors method of DropletService
func (m *mockDropletService) Neighbors(ctx context.Context, dropletID int) ([]godo.Droplet, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// ActionByID mocks the ActionByID method of DropletService
func (m *mockDropletService) ActionByID(ctx context.Context, dropletID int, actionID int) (*godo.Action, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// ActionByTag mocks the ActionByTag method of DropletService
func (m *mockDropletService) ActionByTag(ctx context.Context, tag string, actionID int) ([]godo.Action, *godo.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// DeleteByTag mocks the DeleteByTag method of DropletService
func (m *mockDropletService) DeleteByTag(ctx context.Context, tag string) (*godo.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetBackupPolicy mocks the GetBackupPolicy method of DropletService
func (m *mockDropletService) GetBackupPolicy(ctx context.Context, dropletID int) (*godo.DropletBackupPolicy, *godo.Response, error) {
	if m.GetBackupPolicyFunc != nil {
		return m.GetBackupPolicyFunc(ctx, dropletID)
	}
	return nil, nil, nil
}

// ListBackupPolicies mocks the ListBackupPolicies method of DropletService
func (m *mockDropletService) ListBackupPolicies(ctx context.Context, opt *godo.ListOptions) (map[int]*godo.DropletBackupPolicy, *godo.Response, error) {
	if m.ListBackupPoliciesFunc != nil {
		return m.ListBackupPoliciesFunc(ctx, opt)
	}
	return nil, nil, nil
}

// ListSupportedBackupPolicies mocks the ListSupportedBackupPolicies method of DropletService
func (m *mockDropletService) ListSupportedBackupPolicies(ctx context.Context) ([]*godo.SupportedBackupPolicy, *godo.Response, error) {
	if m.ListSupportedBackupPoliciesFunc != nil {
		return m.ListSupportedBackupPoliciesFunc(ctx)
	}
	return nil, nil, nil
}

// mockKeyService implements the KeyService interface
type mockKeyService struct {
	keys map[int]*godo.Key
}

// List mocks the List method of KeyService
func (m *mockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	if m.keys == nil {
		return nil, nil, fmt.Errorf("key service not initialized")
	}

	var keys []godo.Key
	for _, key := range m.keys {
		keys = append(keys, *key)
	}
	return keys, &godo.Response{}, nil
}

// GetByID mocks the GetByID method of KeyService
func (m *mockKeyService) GetByID(ctx context.Context, id int) (*godo.Key, *godo.Response, error) {
	if key, ok := m.keys[id]; ok {
		return key, &godo.Response{}, nil
	}
	return nil, nil, fmt.Errorf("key not found")
}

// GetByFingerprint mocks the GetByFingerprint method of KeyService
func (m *mockKeyService) GetByFingerprint(ctx context.Context, fingerprint string) (*godo.Key, *godo.Response, error) {
	for _, key := range m.keys {
		if key.Fingerprint == fingerprint {
			return key, &godo.Response{}, nil
		}
	}
	return nil, nil, fmt.Errorf("key not found")
}

// Create mocks the Create method of KeyService
func (m *mockKeyService) Create(ctx context.Context, createRequest *godo.KeyCreateRequest) (*godo.Key, *godo.Response, error) {
	key := &godo.Key{
		ID:          len(m.keys) + 1,
		Name:        createRequest.Name,
		PublicKey:   createRequest.PublicKey,
		Fingerprint: "mock-fingerprint",
	}
	m.keys[key.ID] = key
	return key, &godo.Response{}, nil
}

// UpdateByID mocks the UpdateByID method of KeyService
func (m *mockKeyService) UpdateByID(ctx context.Context, id int, updateRequest *godo.KeyUpdateRequest) (*godo.Key, *godo.Response, error) {
	if key, ok := m.keys[id]; ok {
		key.Name = updateRequest.Name
		return key, &godo.Response{}, nil
	}
	return nil, nil, fmt.Errorf("key not found")
}

// UpdateByFingerprint mocks the UpdateByFingerprint method of KeyService
func (m *mockKeyService) UpdateByFingerprint(ctx context.Context, fingerprint string, updateRequest *godo.KeyUpdateRequest) (*godo.Key, *godo.Response, error) {
	for _, key := range m.keys {
		if key.Fingerprint == fingerprint {
			key.Name = updateRequest.Name
			return key, &godo.Response{}, nil
		}
	}
	return nil, nil, fmt.Errorf("key not found")
}

// DeleteByID mocks the DeleteByID method of KeyService
func (m *mockKeyService) DeleteByID(ctx context.Context, id int) (*godo.Response, error) {
	delete(m.keys, id)
	return &godo.Response{}, nil
}

// DeleteByFingerprint mocks the DeleteByFingerprint method of KeyService
func (m *mockKeyService) DeleteByFingerprint(ctx context.Context, fingerprint string) (*godo.Response, error) {
	for id, key := range m.keys {
		if key.Fingerprint == fingerprint {
			delete(m.keys, id)
			return &godo.Response{}, nil
		}
	}
	return nil, fmt.Errorf("key not found")
}

// Get mocks the Get method of KeyService
func (m *mockKeyService) Get(ctx context.Context, id int) (*godo.Key, *godo.Response, error) {
	return m.GetByID(ctx, id)
}
