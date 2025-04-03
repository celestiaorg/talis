// DigitalOcean provider for creating and managing DigitalOcean droplets
// https://github.com/digitalocean/godo/blob/main/droplets.go#L18
package compute

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// DefaultDOClient is the standard implementation of DOClient using godo
type DefaultDOClient struct {
	client *godo.Client
}

// ConfigureProvider configures the provider with the given stack
func (c *DefaultDOClient) ConfigureProvider(stack interface{}) error {
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
func (c *DefaultDOClient) CreateInstance(ctx context.Context, name string, config types.InstanceConfig) ([]types.InstanceInfo, error) {
	provider := &DigitalOceanProvider{doClient: c}
	return provider.CreateInstance(ctx, name, config)
}

// DeleteInstance deletes an instance
func (c *DefaultDOClient) DeleteInstance(ctx context.Context, name string, region string) error {
	provider := &DigitalOceanProvider{doClient: c}
	return provider.DeleteInstance(ctx, name, region)
}

// Droplets returns the droplet service
func (c *DefaultDOClient) Droplets() types.DropletService {
	return &DefaultDropletService{service: c.client.Droplets}
}

// Keys returns the key service
func (c *DefaultDOClient) Keys() types.KeyService {
	return &DefaultKeyService{service: c.client.Keys}
}

// Storage returns the storage service
func (c *DefaultDOClient) Storage() types.StorageService {
	return &DefaultStorageService{
		service: c.client.Storage,
		actions: c.client.StorageActions,
	}
}

// NewDOClient creates a new DefaultDOClient
func NewDOClient(token string) types.DOClient {
	return &DefaultDOClient{
		client: godo.NewFromToken(token),
	}
}

// DefaultDropletService implements DropletService using godo
type DefaultDropletService struct {
	service godo.DropletsService
}

// Create creates a new droplet
func (s *DefaultDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	return s.service.Create(ctx, createRequest)
}

// CreateMultiple creates multiple droplets
func (s *DefaultDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	return s.service.CreateMultiple(ctx, createRequest)
}

// Delete deletes a droplet
func (s *DefaultDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	return s.service.Delete(ctx, id)
}

// Get retrieves a droplet by ID
func (s *DefaultDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	return s.service.Get(ctx, id)
}

// List lists all droplets
func (s *DefaultDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DefaultKeyService implements KeyService using godo
type DefaultKeyService struct {
	service godo.KeysService
}

// List lists all SSH keys
func (s *DefaultKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DefaultStorageService implements StorageService using godo
type DefaultStorageService struct {
	service godo.StorageService
	actions godo.StorageActionsService
}

// CreateVolume creates a new block storage volume in DigitalOcean.
// It directly maps to the DigitalOcean API for volume creation.
func (s *DefaultStorageService) CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
	return s.service.CreateVolume(ctx, request)
}

// DeleteVolume deletes a block storage volume from DigitalOcean by its ID.
func (s *DefaultStorageService) DeleteVolume(ctx context.Context, id string) (*godo.Response, error) {
	return s.service.DeleteVolume(ctx, id)
}

// ListVolumes retrieves all block storage volumes associated with the account.
func (s *DefaultStorageService) ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
	return s.service.ListVolumes(ctx, opt)
}

// GetVolume retrieves information about a specific block storage volume.
func (s *DefaultStorageService) GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error) {
	return s.service.GetVolume(ctx, id)
}

// GetVolumeAction retrieves the status of a volume action.
func (s *DefaultStorageService) GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error) {
	return s.actions.Get(ctx, volumeID, actionID)
}

// AttachVolume attaches a block storage volume to a droplet and waits for completion.
// The operation is considered complete when the volume is successfully attached
// or when it fails after maximum retries.
func (s *DefaultStorageService) AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	if s.actions == nil {
		return nil, fmt.Errorf("storage actions service not initialized")
	}

	action, resp, err := s.actions.Attach(ctx, volumeID, dropletID)
	if err != nil {
		return resp, fmt.Errorf("failed to attach volume: %w", err)
	}

	// Wait a few seconds before checking the action status
	time.Sleep(5 * time.Second)

	// Wait for the attach action to complete
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		action, resp, err = s.actions.Get(ctx, volumeID, action.ID)
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
			return resp, fmt.Errorf("volume attach action errored")
		}

		time.Sleep(5 * time.Second)
	}

	return resp, fmt.Errorf("volume attach action did not complete after %d retries", maxRetries)
}

// DetachVolume detaches a block storage volume from a droplet and waits for completion.
// The operation is considered complete when the volume is successfully detached
// or when it fails after maximum retries.
func (s *DefaultStorageService) DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	if s.actions == nil {
		return nil, fmt.Errorf("storage actions service not initialized")
	}

	action, resp, err := s.actions.DetachByDropletID(ctx, volumeID, dropletID)
	if err != nil {
		return resp, fmt.Errorf("failed to detach volume: %w", err)
	}

	// Wait for the detach action to complete
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		action, resp, err = s.actions.Get(ctx, volumeID, action.ID)
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
			return resp, fmt.Errorf("volume detach action errored")
		}

		time.Sleep(5 * time.Second)
	}

	return resp, fmt.Errorf("volume detach action did not complete after %d retries", maxRetries)
}

// DigitalOceanProvider implements the ComputeProvider interface
type DigitalOceanProvider struct {
	doClient types.DOClient
}

// SetClient sets the DO client for testing
func (p *DigitalOceanProvider) SetClient(client types.DOClient) {
	p.doClient = client
}

// ConfigureProvider is a no-op since we're not using Pulumi anymore
func (p *DigitalOceanProvider) ConfigureProvider(stack interface{}) error {
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

	logger.Info("⏳ Waiting for droplet to get an IP address...")
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
				logger.Infof("📍 Found public IP for droplet: %s", ip)
				return ip, nil
			}
		}

		logger.Infof("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...", i+1, maxRetries)
		time.Sleep(interval)
	}

	return "", fmt.Errorf("droplet created but no public IP found after %d retries", maxRetries)
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(
	ctx context.Context,
	name string,
	config types.InstanceConfig,
) ([]types.InstanceInfo, error) {
	if p.doClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	logger.Infof("🚀 Creating DigitalOcean droplet(s): %s", name)
	logger.Infof("  Region: %s", config.Region)
	logger.Infof("  Size: %s", config.Size)
	logger.Infof("  Image: %s", config.Image)
	logger.Infof("  Number of instances: %d", config.NumberOfInstances)

	// Get SSH key ID
	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		logger.Errorf("❌ Failed to get SSH key: %v", err)
		return nil, fmt.Errorf("failed to get SSH key: %w", err)
	}

	// Create single or multiple droplets based on configuration
	if config.NumberOfInstances > 1 {
		return p.createMultipleDroplets(ctx, name, config, sshKeyID)
	}

	// For single instance, wrap the result in a slice for consistent interface
	instance, err := p.createSingleDroplet(ctx, name, config, sshKeyID)
	if err != nil {
		return nil, err
	}
	return []types.InstanceInfo{instance}, nil
}

// createDropletRequest creates a DropletCreateRequest with common configuration
func (p *DigitalOceanProvider) createDropletRequest(
	name string,
	config types.InstanceConfig,
	sshKeyID int,
) *godo.DropletCreateRequest {
	return &godo.DropletCreateRequest{
		Name:   name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyID},
		},
		Tags: append([]string{name}, config.Tags...),
		UserData: fmt.Sprintf(`#!/bin/bash
apt-get update
apt-get install -y python3

# Mount volumes if specified
%s
`, p.generateVolumeMountScript(config.Volumes)),
	}
}

// createMultipleDroplets creates multiple droplets using the CreateMultiple API
func (p *DigitalOceanProvider) createMultipleDroplets(
	ctx context.Context,
	name string,
	config types.InstanceConfig,
	sshKeyID int,
) ([]types.InstanceInfo, error) {
	if p.doClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	const maxDropletsPerBatch = 10
	var allInstances []types.InstanceInfo
	remainingInstances := config.NumberOfInstances
	batchNumber := 0

	for remainingInstances > 0 {
		// Calculate how many instances to create in this batch
		batchSize := remainingInstances
		if batchSize > maxDropletsPerBatch {
			batchSize = maxDropletsPerBatch
		}

		// Create names for this batch
		names := make([]string, batchSize)
		startIndex := batchNumber * maxDropletsPerBatch
		for i := 0; i < batchSize; i++ {
			if config.CustomName != "" {
				names[i] = fmt.Sprintf("%s-%d", config.CustomName, startIndex+i)
			} else {
				names[i] = fmt.Sprintf("%s-%d", name, startIndex+i)
			}
		}

		createRequest := &godo.DropletMultiCreateRequest{
			Names:  names,
			Region: config.Region,
			Size:   config.Size,
			Image: godo.DropletCreateImage{
				Slug: config.Image,
			},
			SSHKeys: []godo.DropletCreateSSHKey{
				{ID: sshKeyID},
			},
			Tags: append([]string{name}, config.Tags...),
			UserData: fmt.Sprintf(`#!/bin/bash
apt-get update
apt-get install -y python3

# Mount volumes if specified
%s
`, p.generateVolumeMountScript(config.Volumes)),
		}

		logger.Infof("🚀 Creating batch %d of droplets (%d instances)...", batchNumber+1, batchSize)

		droplets, _, err := p.doClient.Droplets().CreateMultiple(ctx, createRequest)
		if err != nil {
			logger.Errorf("❌ Failed to create droplets: %v", err)
			return nil, fmt.Errorf("failed to create droplets: %w", err)
		}

		// Wait for droplets to be ready and get their public IPs
		for _, droplet := range droplets {
			instance := types.InstanceInfo{
				ID:       fmt.Sprintf("%d", droplet.ID),
				Name:     droplet.Name,
				Provider: "do",
				Region:   droplet.Region.Slug,
				Size:     droplet.Size.Slug,
			}

			// Wait for public IP
			ip, err := p.waitForPublicIP(ctx, droplet.ID)
			if err != nil {
				logger.Errorf("❌ Failed to get public IP for droplet %s: %v", droplet.Name, err)
				return nil, fmt.Errorf("failed to get public IP for droplet %s: %w", droplet.Name, err)
			}
			instance.PublicIP = ip

			// Create and attach volumes if specified
			if len(config.Volumes) > 0 {
				volumeIDs, err := p.createAndAttachVolumes(ctx, droplet.ID, droplet.Name, config)
				if err != nil {
					logger.Errorf("❌ Failed to create/attach volumes for droplet %s: %v", droplet.Name, err)
					return nil, fmt.Errorf("failed to create/attach volumes for droplet %s: %w", droplet.Name, err)
				}
				instance.Volumes = volumeIDs
			}

			allInstances = append(allInstances, instance)
		}

		remainingInstances -= batchSize
		batchNumber++
	}

	return allInstances, nil
}

// createSingleDroplet creates a single droplet
func (p *DigitalOceanProvider) createSingleDroplet(
	ctx context.Context,
	name string,
	config types.InstanceConfig,
	sshKeyID int,
) (types.InstanceInfo, error) {
	if p.doClient == nil {
		return types.InstanceInfo{}, fmt.Errorf("client not initialized")
	}

	// Create droplet
	createRequest := p.createDropletRequest(name, config, sshKeyID)
	droplet, _, err := p.doClient.Droplets().Create(ctx, createRequest)
	if err != nil {
		logger.Errorf("❌ Failed to create droplet: %v", err)
		return types.InstanceInfo{}, fmt.Errorf("failed to create droplet: %w", err)
	}

	// Initialize instance info
	instance := types.InstanceInfo{
		ID:       fmt.Sprintf("%d", droplet.ID),
		Name:     droplet.Name,
		Provider: "do",
		Region:   droplet.Region.Slug,
		Size:     droplet.Size.Slug,
	}

	// Wait for public IP
	ip, err := p.waitForPublicIP(ctx, droplet.ID)
	if err != nil {
		logger.Errorf("❌ Failed to get public IP for droplet %s: %v", droplet.Name, err)
		return types.InstanceInfo{}, fmt.Errorf("failed to get public IP for droplet %s: %w", droplet.Name, err)
	}
	instance.PublicIP = ip

	// Create and attach volumes if specified
	if len(config.Volumes) > 0 {
		volumeIDs, err := p.createAndAttachVolumes(ctx, droplet.ID, droplet.Name, config)
		if err != nil {
			logger.Errorf("❌ Failed to create/attach volumes for droplet %s: %v", droplet.Name, err)
			return types.InstanceInfo{}, fmt.Errorf("failed to create/attach volumes for droplet %s: %w", droplet.Name, err)
		}
		instance.Volumes = volumeIDs
	}

	return instance, nil
}

// generateVolumeMountScript generates a bash script to mount volumes
func (p *DigitalOceanProvider) generateVolumeMountScript(volumes []types.VolumeConfig) string {
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

// getSSHKeyID gets the ID of an SSH key by its name
func (p *DigitalOceanProvider) getSSHKeyID(ctx context.Context, keyName string) (int, error) {
	if p.doClient == nil {
		return 0, fmt.Errorf("client not initialized")
	}

	logger.Infof("🔑 Looking up SSH key: %s", keyName)

	// List all SSH keys
	keys, _, err := p.doClient.Keys().List(ctx, &godo.ListOptions{})
	if err != nil {
		logger.Errorf("❌ Failed to list SSH keys: %v", err)
		return 0, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	// Find the key by name
	for _, key := range keys {
		if key.Name == keyName {
			logger.Infof("✅ Found SSH key '%s' with ID: %d", keyName, key.ID)
			return key.ID, nil
		}
	}

	// If we get here, print available keys to help with diagnosis
	if len(keys) > 0 {
		logger.Infof("Available SSH keys:")
		for _, key := range keys {
			logger.Infof("  - %s (ID: %d)", key.Name, key.ID)
		}
	}

	return 0, fmt.Errorf("SSH key '%s' not found", keyName)
}

// createAndAttachVolumes creates and attaches volumes to a droplet
func (p *DigitalOceanProvider) createAndAttachVolumes(
	ctx context.Context,
	dropletID int,
	name string,
	config types.InstanceConfig,
) ([]string, error) {
	var volumeIDs []string

	// If no volumes specified, return empty list
	if len(config.Volumes) == 0 {
		return volumeIDs, nil
	}

	logger.Infof("📦 Creating and attaching volumes for droplet %d", dropletID)

	for _, vol := range config.Volumes {
		logger.Infof("📦 Processing volume configuration: name=%s, size=%dGB, region=%s", vol.Name, vol.SizeGB, vol.Region)

		// Generate volume name with random suffix
		suffix := generateRandomSuffix()
		volumeName := fmt.Sprintf("%s-%s", name, suffix)
		volumeRequest := &godo.VolumeCreateRequest{
			Name:          volumeName,
			SizeGigaBytes: int64(vol.SizeGB),
			Region:        vol.Region,
		}

		logger.Infof("📦 Creating volume %s with size %dGB in region %s", volumeName, vol.SizeGB, vol.Region)
		volume, resp, err := p.doClient.Storage().CreateVolume(ctx, volumeRequest)
		if err != nil {
			logger.Errorf("❌ Failed to create volume: %v (Response: %+v)", err, resp)
			return nil, fmt.Errorf("failed to create volume %s: %w", volumeName, err)
		}
		logger.Infof("✅ Volume created successfully: %s (ID: %s)", volumeName, volume.ID)

		// Wait for volume to be ready
		logger.Infof("⏳ Waiting for volume to be ready...")
		time.Sleep(10 * time.Second)

		// Verify volume exists and is ready
		vol, _, err := p.doClient.Storage().GetVolume(ctx, volume.ID)
		if err != nil {
			logger.Errorf("❌ Failed to verify volume status: %v", err)
			return nil, fmt.Errorf("failed to verify volume status: %w", err)
		}
		logger.Infof("✅ Volume is ready: %s", vol.ID)

		logger.Infof("📦 Attaching volume %s to droplet %d", volume.ID, dropletID)
		resp, err = p.doClient.Storage().AttachVolume(ctx, volume.ID, dropletID)
		if err != nil {
			logger.Errorf("❌ Failed to attach volume: %v (Response: %+v)", err, resp)
			// Try to clean up the volume if attachment fails
			logger.Infof("🗑️ Attempting to delete volume %s after attachment failure", volume.ID)
			deleteResp, deleteErr := p.doClient.Storage().DeleteVolume(ctx, volume.ID)
			if deleteErr != nil {
				logger.Warnf("⚠️ Warning: Failed to delete volume %s after attachment failure: %v (Response: %+v)", volume.ID, deleteErr, deleteResp)
			} else {
				logger.Infof("✅ Successfully deleted volume %s after attachment failure", volume.ID)
			}
			return nil, fmt.Errorf("failed to attach volume %s: %w", volumeName, err)
		}

		// Wait for volume to be attached
		logger.Infof("⏳ Waiting for volume to be attached...")
		time.Sleep(10 * time.Second)

		volumeIDs = append(volumeIDs, volume.ID)
		logger.Infof("✅ Successfully created and attached volume %s (%s)", volumeName, volume.ID)
	}

	return volumeIDs, nil
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
func (p *DigitalOceanProvider) DeleteInstance(ctx context.Context, name string, region string) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	logger.Infof("🗑️ Deleting DigitalOcean droplet: %s in region %s", name, region)

	// List all droplets to find the one with our name in the specific region
	droplets, _, err := p.doClient.Droplets().List(ctx, &godo.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list droplets: %w", err)
	}

	// Find the droplet by name and region
	var dropletID int
	var found bool
	for _, d := range droplets {
		if d.Name == name && d.Region.Slug == region {
			dropletID = d.ID
			found = true
			break
		}
	}

	if !found {
		logger.Infof("⚠️ Droplet %s in region %s not found, nothing to delete", name, region)
		return nil
	}

	// Get droplet details to find attached volumes
	droplet, _, err := p.doClient.Droplets().Get(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to get droplet details: %w", err)
	}

	// List all volumes to find those attached to this droplet
	volumes, _, err := p.doClient.Storage().ListVolumes(ctx, &godo.ListVolumeParams{
		Region: region,
	})
	if err != nil {
		logger.Warnf("⚠️ Warning: Failed to list volumes: %v", err)
	} else {
		// Delete volumes that are attached to this droplet
		for _, volume := range volumes {
			for _, id := range volume.DropletIDs {
				if id == droplet.ID {
					logger.Infof("🗑️ Detaching and deleting volume: %s", volume.ID)

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
	logger.Infof("🗑️ Deleting droplet with ID: %d", dropletID)
	_, err = p.doClient.Droplets().Delete(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	logger.Infof("✅ Deleted droplet: %s", name)
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
