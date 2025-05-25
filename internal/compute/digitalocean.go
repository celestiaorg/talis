// Package compute provides infrastructure provider implementations.
// DigitalOcean provider for creating and managing DigitalOcean droplets
// https://github.com/digitalocean/godo/blob/main/droplets.go#L18
package compute

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	computeTypes "github.com/celestiaorg/talis/internal/compute/types"
	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/logger"
	talisTypes "github.com/celestiaorg/talis/internal/types"
)

// DefaultDOClient is the default implementation of DOClient
type DefaultDOClient struct {
	client *godo.Client
}

// ConfigureProvider configures the provider with the given stack
func (c *DefaultDOClient) ConfigureProvider(_ interface{}) error {
	return nil
}

// ValidateCredentials validates the provider credentials
func (c *DefaultDOClient) ValidateCredentials() error {
	_, _, err := c.client.Account.Get(context.Background())
	return err
}

// GetEnvironmentVars returns the environment variables needed for the provider
func (c *DefaultDOClient) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// CreateInstance creates a new instance
func (c *DefaultDOClient) CreateInstance(ctx context.Context, config *talisTypes.InstanceRequest) error {
	provider := &DigitalOceanProvider{doClient: c}
	return provider.CreateInstance(ctx, config)
}

// DeleteInstance deletes an instance
func (c *DefaultDOClient) DeleteInstance(ctx context.Context, dropletID int) error {
	provider := &DigitalOceanProvider{doClient: c}
	return provider.DeleteInstance(ctx, dropletID)
}

// Droplets returns the droplet service
func (c *DefaultDOClient) Droplets() computeTypes.DropletService {
	return &DefaultDropletService{service: c.client.Droplets}
}

// Keys returns the key service
func (c *DefaultDOClient) Keys() computeTypes.KeyService {
	return &DefaultKeyService{service: c.client.Keys}
}

// Storage returns the storage service
func (c *DefaultDOClient) Storage() computeTypes.StorageService {
	return &DefaultStorageService{
		service: c.client.Storage,
		actions: c.client.StorageActions,
	}
}

// NewDOClient creates a new DigitalOcean client
func NewDOClient(token string) computeTypes.DOClient {
	client := godo.NewFromToken(token)
	return &DefaultDOClient{
		client: client,
	}
}

// DefaultDropletService adapts godo.DropletService to our DropletService interface
type DefaultDropletService struct {
	service godo.DropletsService
}

// Create creates a new droplet
func (s *DefaultDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	return s.service.Create(ctx, createRequest)
}

// Get gets a droplet
func (s *DefaultDropletService) Get(ctx context.Context, dropletID int) (*godo.Droplet, *godo.Response, error) {
	return s.service.Get(ctx, dropletID)
}

// Delete deletes a droplet
func (s *DefaultDropletService) Delete(ctx context.Context, dropletID int) (*godo.Response, error) {
	return s.service.Delete(ctx, dropletID)
}

// List lists all droplets
func (s *DefaultDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DefaultKeyService adapts godo.KeyService to our KeyService interface
type DefaultKeyService struct {
	service godo.KeysService
}

// List lists all SSH keys
func (s *DefaultKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DefaultStorageService adapts godo.StorageService to our StorageService interface
type DefaultStorageService struct {
	service godo.StorageService
	actions godo.StorageActionsService
}

// CreateVolume creates a new volume
func (s *DefaultStorageService) CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
	return s.service.CreateVolume(ctx, request)
}

// DeleteVolume deletes a volume
func (s *DefaultStorageService) DeleteVolume(ctx context.Context, id string) (*godo.Response, error) {
	return s.service.DeleteVolume(ctx, id)
}

// ListVolumes lists all volumes
func (s *DefaultStorageService) ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
	return s.service.ListVolumes(ctx, opt)
}

// GetVolume gets a volume
func (s *DefaultStorageService) GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error) {
	return s.service.GetVolume(ctx, id)
}

// GetVolumeAction gets a volume action
func (s *DefaultStorageService) GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error) {
	return s.actions.Get(ctx, volumeID, actionID)
}

// waitForVolumeAction waits for a volume action to complete with retries
func (s *DefaultStorageService) waitForVolumeAction(
	ctx context.Context,
	volumeID string,
	actionID int,
	actionType string,
) (*godo.Response, error) {
	// Wait a few seconds before checking the action status
	time.Sleep(5 * time.Second)

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		action, resp, err := s.actions.Get(ctx, volumeID, actionID)
		if err != nil {
			// If we get a 404, wait and retry
			if resp != nil && resp.StatusCode == 404 {
				time.Sleep(5 * time.Second)
				continue
			}
			return resp, fmt.Errorf("failed to get volume action status: %w", err)
		}

		if action.Status == "completed" {
			return resp, nil
		}

		if action.Status == "errored" {
			return resp, fmt.Errorf("volume %s action errored", actionType)
		}

		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("volume %s action did not complete after %d retries", actionType, maxRetries)
}

// AttachVolume attaches a block storage volume to a droplet and waits for completion.
// The operation is considered complete when the volume is successfully attached
// or when it fails after maximum retries.
func (s *DefaultStorageService) AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	action, resp, err := s.actions.Attach(ctx, volumeID, dropletID)
	if err != nil {
		return resp, fmt.Errorf("failed to attach volume: %w", err)
	}

	return s.waitForVolumeAction(ctx, volumeID, action.ID, "attach")
}

// DetachVolume detaches a block storage volume from a droplet and waits for completion.
// The operation is considered complete when the volume is successfully detached
// or when it fails after maximum retries.
func (s *DefaultStorageService) DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	action, resp, err := s.actions.DetachByDropletID(ctx, volumeID, dropletID)
	if err != nil {
		return resp, fmt.Errorf("failed to detach volume: %w", err)
	}

	return s.waitForVolumeAction(ctx, volumeID, action.ID, "detach")
}

// DigitalOceanProvider implements the ComputeProvider interface
type DigitalOceanProvider struct {
	doClient computeTypes.DOClient
}

// SetClient sets the DO client for testing
func (p *DigitalOceanProvider) SetClient(client computeTypes.DOClient) {
	p.doClient = client
}

// ConfigureProvider is a no-op since we're not using Pulumi anymore
func (p *DigitalOceanProvider) ConfigureProvider(_ interface{}) error {
	return nil
}

// ValidateCredentials validates the DigitalOcean credentials
func (p *DigitalOceanProvider) ValidateCredentials() error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for the provider
func (p *DigitalOceanProvider) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// waitForPublicIP waits for a droplet to get a public IP address
func (p *DigitalOceanProvider) waitForPublicIP(ctx context.Context, dropletID int) (string, error) {
	if p.doClient == nil {
		return "", fmt.Errorf("client not initialized")
	}

	logger.Debug("⏳ Waiting for droplet to get an IP address...")
	maxRetries := 10
	interval := 10 * time.Second

	for i := 0; i < maxRetries; i++ {
		d, _, err := p.doClient.Droplets().Get(ctx, dropletID)
		if err != nil {
			logger.Errorf("❌ Failed to get droplet details: %v", err)
			time.Sleep(interval)
			continue
		}

		// Get the public IPv4 address
		for _, network := range d.Networks.V4 {
			if network.Type == "public" {
				ip := network.IPAddress
				logger.Debugf("📍 Found public IP for droplet: %s", ip)
				return ip, nil
			}
		}

		logger.Debugf("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...", i+1, maxRetries)
		time.Sleep(interval)
	}

	return "", fmt.Errorf("droplet created but no public IP found after %d retries", maxRetries)
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(
	ctx context.Context,
	config *talisTypes.InstanceRequest,
) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	// Log based on ProjectName as instance Name is removed
	logger.Debugf("🚀 Creating DigitalOcean droplet(s) for project: %s", config.ProjectName)
	logger.Debugf("  Region: %s", config.Region)
	logger.Debugf("  Size: %s", config.Size)
	logger.Debugf("  Image: %s", config.Image)
	logger.Debugf("  Number of instances: %d", config.NumberOfInstances)

	// Get SSH key ID from environment variable
	sshKeyID, err := p.getSSHKeyID(ctx, "")
	if err != nil {
		logger.Errorf("❌ Failed to get SSH key: %v", err)
		return fmt.Errorf("failed to get SSH key: %w", err)
	}

	// Create the droplet with the appropriate instance index
	// Note: The service layer only calls this method once per instance,
	// so we always use index 0 here. The instance index is used by the
	// createDropletRequest method to generate the appropriate name.
	return p.createSingleDroplet(ctx, config, sshKeyID)
}

// createDropletRequest creates a DropletCreateRequest with common configuration
func (p *DigitalOceanProvider) createDropletRequest(
	config *talisTypes.InstanceRequest,
	sshKeyID int,
) *godo.DropletCreateRequest {
	// Generate a name for the DO droplet
	var dropletName string
	if config.Name != "" {
		// If user provided a name and we're creating multiple instances, add index suffix
		if config.NumberOfInstances > 1 {
			dropletName = fmt.Sprintf("%s-%d", config.Name, config.InstanceIndex+1)
		} else {
			dropletName = config.Name
		}
	} else {
		// Fallback to project name with random suffix if no name provided
		dropletName = fmt.Sprintf("%s-%s", config.ProjectName, generateRandomSuffix())
	}

	return &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyID},
		},
		Tags: append([]string{dropletName}, config.Tags...),
		UserData: fmt.Sprintf(`#!/bin/bash
apt-get update
apt-get install -y python3

# Mount volumes if specified
%s
`, p.generateVolumeMountScript(config.Volumes)),
	}
}

// createSingleDroplet creates a single droplet
func (p *DigitalOceanProvider) createSingleDroplet(
	ctx context.Context,
	config *talisTypes.InstanceRequest,
	sshKeyID int,
) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	// Create droplet
	createRequest := p.createDropletRequest(config, sshKeyID)
	logger.Debugf("  Sending droplet creation request: %+v", createRequest)
	logger.Debugf("📝 Initiated instance creation for project: %s", config.ProjectName)

	droplet, _, err := p.doClient.Droplets().Create(ctx, createRequest)
	if err != nil {
		logger.Errorf("❌ Failed to create droplet: %v", err)
		return fmt.Errorf("failed to create droplet: %w", err)
	}

	// Initialize instance info
	config.ProviderInstanceID = droplet.ID
	config.VolumeIDs = []string{}
	config.VolumeDetails = []talisTypes.VolumeDetails{}

	// Wait for public IP
	ip, err := p.waitForPublicIP(ctx, droplet.ID)
	if err != nil {
		errMsg := fmt.Errorf("❌ Failed to get public IP for droplet %s: %w", droplet.Name, err)
		logger.Error(errMsg)
		return errMsg
	}
	config.PublicIP = ip

	// Create and attach volumes if specified
	if len(config.Volumes) > 0 {
		volumeIDs, volumeDetails, err := p.createAndAttachVolumes(ctx, droplet.ID, config)
		if err != nil {
			logger.Errorf("❌ Failed to create/attach volumes for droplet %s: %v", droplet.Name, err)
			return fmt.Errorf("failed to create/attach volumes for droplet %s: %w", droplet.Name, err)
		}
		logger.Debugf("📦 Setting volumes: %v and details: %+v", volumeIDs, volumeDetails)
		config.VolumeIDs = volumeIDs
		config.VolumeDetails = volumeDetails
	}

	logger.Infof("✅ Droplet '%s' (ID: %d) created successfully. Waiting for IP...", createRequest.Name, droplet.ID)
	return nil
}

// generateVolumeMountScript generates a bash script to mount volumes
func (p *DigitalOceanProvider) generateVolumeMountScript(volumes []talisTypes.VolumeConfig) string {
	if len(volumes) == 0 {
		return ""
	}

	var script strings.Builder
	script.WriteString("\n# Mount volumes\n")

	for _, vol := range volumes {
		if vol.MountPoint == "" {
			continue
		}

		script.WriteString(fmt.Sprintf(`
# Mount volume %s
mkdir -p %s
if [ -n "$(blkid | grep /dev/disk/by-id/*%s)" ]; then
    device=$(readlink -f /dev/disk/by-id/*%s)
    if [ -n "$device" ]; then
        if [ -n "%s" ]; then
            mkfs.%s "$device" || true
        fi
        echo "$device %s ext4 defaults,nofail 0 2" >> /etc/fstab
        mount %s || true
    fi
fi
`, vol.Name, vol.MountPoint, vol.Name, vol.Name, vol.FileSystem, vol.FileSystem, vol.MountPoint, vol.MountPoint))
	}

	return script.String()
}

// getSSHKeyID gets the ID of an SSH key from the environment variable
func (p *DigitalOceanProvider) getSSHKeyID(ctx context.Context, _ string) (int, error) {
	if p.doClient == nil {
		return 0, fmt.Errorf("client not initialized")
	}

	// Get SSH key name from environment variable
	keyName := os.Getenv(constants.EnvTalisSSHKeyName)
	if keyName != "" {
		logger.Debugf("🔑 Using SSH key name from environment variable: %s", keyName)
	} else {
		// If not set, use the default test key
		keyName = "test-key"
		logger.Warnf("🔑 Environment variable %s not set, using default test key: %s", constants.EnvTalisSSHKeyName, keyName)
	}

	logger.Debugf("🔑 Looking up SSH key: %s", keyName)

	// List all SSH keys
	keys, _, err := p.doClient.Keys().List(ctx, &godo.ListOptions{})
	if err != nil {
		logger.Errorf("❌ Failed to list SSH keys: %v", err)
		return 0, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	// Find the key by name
	for _, key := range keys {
		if key.Name == keyName {
			logger.Debugf("✅ Found SSH key '%s' with ID: %d", keyName, key.ID)
			return key.ID, nil
		}
	}

	// If we get here, print available keys to help with diagnosis
	if len(keys) > 0 {
		logger.Debugf("Available SSH keys:")
		for _, key := range keys {
			logger.Debugf("  - %s (ID: %d)", key.Name, key.ID)
		}
	}

	return 0, fmt.Errorf("SSH key '%s' not found", keyName)
}

// createAndAttachVolumes creates and attaches volumes to a droplet
func (p *DigitalOceanProvider) createAndAttachVolumes(
	ctx context.Context,
	dropletID int,
	config *talisTypes.InstanceRequest,
) ([]string, []talisTypes.VolumeDetails, error) {
	var volumeIDs []string
	var volumeDetails []talisTypes.VolumeDetails

	// If no volumes specified, return empty lists
	if len(config.Volumes) == 0 {
		return []string{}, []talisTypes.VolumeDetails{}, nil
	}

	logger.Debugf("📦 Creating and attaching volumes for droplet %d", dropletID)

	for _, volConfig := range config.Volumes {
		// Generate a name for the DO volume using ProjectName and VolumeConfig Name
		volName := fmt.Sprintf("%s-%s-%s", config.ProjectName, volConfig.Name, generateRandomSuffix())
		createRequest := &godo.VolumeCreateRequest{
			Name:          volName,
			Region:        config.Region,
			SizeGigaBytes: int64(volConfig.SizeGB),
			Description:   fmt.Sprintf("Volume for project %s", config.ProjectName),
		}

		logger.Debugf("  Sending volume creation request: %+v", createRequest)
		volume, resp, err := p.doClient.Storage().CreateVolume(ctx, createRequest)
		if err != nil {
			// Check if error is due to volume already existing
			if resp != nil && resp.StatusCode == http.StatusConflict {
				logger.Warnf("⚠️ Volume name conflict, retrying with new suffix")
				continue
			}
			logger.Errorf("❌ Failed to create volume: %v (Response: %+v)", err, resp)
			return nil, nil, fmt.Errorf("failed to create volume %s: %w", volName, err)
		}
		logger.Debugf("✅ Volume created successfully: %s (ID: %s)", volName, volume.ID)

		// Store volume details
		volumeDetail := talisTypes.VolumeDetails{
			ID:         volume.ID,
			Name:       volume.Name,
			Region:     volume.Region.Slug,
			SizeGB:     volConfig.SizeGB,
			MountPoint: volConfig.MountPoint,
		}
		volumeDetails = append(volumeDetails, volumeDetail)
		volumeIDs = append(volumeIDs, volume.ID)

		// Wait for volume to be ready
		logger.Debugf("⏳ Waiting for volume to be ready...")
		time.Sleep(10 * time.Second)

		// Verify volume exists and is ready
		vol, _, err := p.doClient.Storage().GetVolume(ctx, volume.ID)
		if err != nil {
			logger.Errorf("❌ Failed to verify volume status: %v", err)
			return nil, nil, fmt.Errorf("failed to verify volume status: %w", err)
		}
		logger.Debugf("✅ Volume is ready: %s", vol.ID)

		logger.Debugf("📦 Attaching volume %s to droplet %d", volume.ID, dropletID)
		resp, err = p.doClient.Storage().AttachVolume(ctx, volume.ID, dropletID)
		if err != nil {
			logger.Errorf("❌ Failed to attach volume: %v (Response: %+v)", err, resp)
			// Try to clean up the volume if attachment fails
			logger.Debugf("🗑️ Attempting to delete volume %s after attachment failure", volume.ID)
			deleteResp, deleteErr := p.doClient.Storage().DeleteVolume(ctx, volume.ID)
			if deleteErr != nil {
				logger.Warnf("⚠️ Warning: Failed to delete volume %s after attachment failure: %v (Response: %+v)", volume.ID, deleteErr, deleteResp)
			} else {
				logger.Debugf("✅ Successfully deleted volume %s after attachment failure", volume.ID)
			}
			return nil, nil, fmt.Errorf("failed to attach volume %s: %w", volName, err)
		}

		// Wait for volume to be attached
		logger.Debugf("⏳ Waiting for volume to be attached...")
		time.Sleep(10 * time.Second)

		logger.Debugf("✅ Successfully created and attached volume %s (%s)", vol.Name, volume.ID)
	}

	logger.Debugf("📦 Returning volumes: %v and details: %+v", volumeIDs, volumeDetails)
	return volumeIDs, volumeDetails, nil
}

// generateRandomSuffix generates a random 6-character string
func generateRandomSuffix() string {
	bytes := make([]byte, 3) // 3 bytes = 6 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// If random fails, use timestamp as fallback
		return hex.EncodeToString([]byte(time.Now().Format("150405"))[:3])
	}
	return hex.EncodeToString(bytes)
}

// DeleteInstance deletes a DigitalOcean droplet and its associated volumes
func (p *DigitalOceanProvider) DeleteInstance(ctx context.Context, dropletID int) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	logger.Debugf("🗑️ Deleting DigitalOcean droplet: %d", dropletID)

	// Get droplet details to find attached volumes
	droplet, _, err := p.doClient.Droplets().Get(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to get droplet details: %w", err)
	}

	// List all volumes to find those attached to this droplet
	volumes, _, err := p.doClient.Storage().ListVolumes(ctx, &godo.ListVolumeParams{
		Region: droplet.Region.Slug,
	})
	if err != nil {
		logger.Warnf("⚠️ Warning: Failed to list volumes: %v", err)
	} else {
		// Delete volumes that are attached to this droplet
		for _, volume := range volumes {
			for _, id := range volume.DropletIDs {
				if id == droplet.ID {
					logger.Debugf("🗑️ Detaching and deleting volume: %s", volume.ID)

					// Try to detach the volume first
					_, err := p.doClient.Storage().DetachVolume(ctx, volume.ID, droplet.ID)
					if err != nil {
						logger.Warnf("⚠️ Warning: Failed to detach volume %s: %v", volume.ID, err)
					}

					// Delete the volume
					_, err = p.doClient.Storage().DeleteVolume(ctx, volume.ID)
					if err != nil {
						logger.Warnf("⚠️ Warning: Failed to delete volume %s: %v", volume.ID, err)
					}
					break
				}
			}
		}
	}

	// Delete the droplet
	logger.Debugf("🗑️ Deleting droplet with ID: %d", dropletID)
	_, err = p.doClient.Droplets().Delete(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	logger.Debugf("✅ Deleted droplet: %d", dropletID)
	return nil
}

// NewDigitalOceanProvider creates a new DigitalOcean provider instance
func NewDigitalOceanProvider() (*DigitalOceanProvider, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	// Create DigitalOcean API client
	doClient := NewDOClient(token)

	return &DigitalOceanProvider{
		doClient: doClient,
	}, nil
}
