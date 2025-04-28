package compute

import (
	"context"
	"fmt"
	"strconv"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
	"gorm.io/gorm"
)

// VirtFusionProvider implements the Provider interface for VirtFusion
type VirtFusionProvider struct {
	client     *Client
	config     *config.VirtFusionConfig
	sshManager *SSHKeyManager
	db         *gorm.DB
}

// NewVirtFusionProvider creates a new VirtFusion provider instance
func NewVirtFusionProvider() (types.Provider, error) {
	cfg, err := config.NewVirtFusionConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create VirtFusion config: %w", err)
	}

	return &VirtFusionProvider{
		config: cfg,
	}, nil
}

// // Provider defines the interface for cloud providers
// type Provider interface {
// 	// ValidateCredentials validates the provider credentials
// 	ValidateCredentials() error

// 	// GetEnvironmentVars returns the environment variables needed for the provider
// 	GetEnvironmentVars() map[string]string

// 	// ConfigureProvider configures the provider with the given stack
// 	ConfigureProvider(stack interface{}) error

// 	// CreateInstance creates a new instance
// 	CreateInstance(ctx context.Context, req *types.InstanceRequest) error

// 	// DeleteInstance deletes an instance
// 	DeleteInstance(ctx context.Context, providerInstanceID int) error
// }

// // Provisioner is the interface for system configuration
// type Provisioner interface {
// 	// ConfigureHost configures a single host
// 	ConfigureHost(ctx context.Context, host string, sshKeyPath string) error

// 	// ConfigureHosts configures multiple hosts in parallel, ensuring SSH readiness
// 	ConfigureHosts(ctx context.Context, hosts []string, sshKeyPath string) error

// 	// CreateInventory creates an Ansible inventory file from instance info
// 	CreateInventory(instance *types.InstanceRequest, sshKeyPath string) (string, error)

// 	// RunAnsiblePlaybook runs the Ansible playbook
// 	RunAnsiblePlaybook(inventoryName string) error
// }

// NewComputeProvider creates a new compute provider based on the provider name
func NewComputeProvider(provider models.ProviderID) (types.Provider, error) {
	switch provider {
	case models.ProviderVirtFusion:
		return NewVirtFusionProvider()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ValidateCredentials validates the VirtFusion credentials
func (p *VirtFusionProvider) ValidateCredentials() error {
	ctx := context.Background()
	_, _, err := p.client.Hypervisors().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for the provider
func (p *VirtFusionProvider) GetEnvironmentVars() map[string]string {
	return p.config.GetEnvironmentVars()
}

// ConfigureProvider configures the provider with the given stack
func (p *VirtFusionProvider) ConfigureProvider(stack interface{}) error {
	// Configure SSH key manager if needed
	if p.sshManager == nil {
		userRepo := repos.NewUserRepository(p.db)
		p.sshManager = NewSSHKeyManager(userRepo)
	}
	return nil
}

// CreateInstance creates a new instance in VirtFusion
func (p *VirtFusionProvider) CreateInstance(ctx context.Context, name string, config types.InstanceConfig) ([]types.InstanceInfo, error) {
	logger.Debugf("Creating new instance: name=%s, size=%s, image=%s, region=%s", name, config.Size, config.Image, config.Region)

	// Get package mapping
	pkg, exists := GetPackageMapping(config.Size)
	if !exists {
		logger.Errorf("Unsupported package size: %s (available: %+v)", config.Size, StandardPackages)
		return nil, fmt.Errorf("unsupported package size: %s", config.Size)
	}
	logger.Debugf("Package mapping found: memory=%d, disk=%d, cpu=%d", pkg.Memory, pkg.Disk, pkg.CPU)

	// Get template mapping
	template, exists := GetTemplateMapping(config.Image)
	if !exists {
		logger.Errorf("Unsupported image: %s (available: %+v)", config.Image, TemplateMapping)
		return nil, fmt.Errorf("unsupported image: %s", config.Image)
	}
	logger.Debugf("Template mapping found: %s", template)

	// Get available hypervisors
	hypervisors, _, err := p.client.Hypervisors().List(ctx)
	if err != nil {
		logger.Errorf("Failed to list hypervisors: %v", err)
		return nil, fmt.Errorf("failed to list hypervisors: %w", err)
	}

	// Select hypervisor based on configuration
	var selectedHypervisor *types.Hypervisor
	desiredHypervisorID := config.HypervisorID
	desiredHypervisorGroup := config.HypervisorGroup

	logger.Debugf("Selecting hypervisor with ID=%d, Group=%s", desiredHypervisorID, desiredHypervisorGroup)

	// First try to find by specific ID if provided
	if desiredHypervisorID > 0 {
		for _, h := range hypervisors {
			if h.ID == desiredHypervisorID {
				if !h.Enabled || h.Maintenance {
					return nil, fmt.Errorf("hypervisor %d is not available (enabled=%v, maintenance=%v)", h.ID, h.Enabled, h.Maintenance)
				}
				selectedHypervisor = h
				logger.Debugf("Selected hypervisor by ID: %d", h.ID)
				break
			}
		}
		if selectedHypervisor == nil {
			return nil, fmt.Errorf("hypervisor with ID %d not found", desiredHypervisorID)
		}
	} else if desiredHypervisorGroup != "" {
		// If no specific ID but group is specified, find first available in group
		for _, h := range hypervisors {
			if !h.Enabled || h.Maintenance {
				continue
			}
			if h.Group != nil && h.Group.Name == desiredHypervisorGroup {
				selectedHypervisor = h
				logger.Debugf("Selected hypervisor by group: %s (ID: %d)", h.Group.Name, h.ID)
				break
			}
		}
		if selectedHypervisor == nil {
			return nil, fmt.Errorf("no available hypervisor found in group %s", desiredHypervisorGroup)
		}
	} else {
		// If neither ID nor group specified, find first available hypervisor
		for _, h := range hypervisors {
			if h.Enabled && !h.Maintenance {
				selectedHypervisor = h
				logger.Debugf("Selected first available hypervisor: ID=%d", h.ID)
				break
			}
		}
		if selectedHypervisor == nil {
			return nil, fmt.Errorf("no available hypervisors found")
		}
	}

	if selectedHypervisor.Group == nil {
		return nil, fmt.Errorf("selected hypervisor %d has no group information", selectedHypervisor.ID)
	}

	logger.Debugf("Selected hypervisor: id=%d, name=%s, group=%s", selectedHypervisor.ID, selectedHypervisor.Name, selectedHypervisor.Group.Name)

	// Get network profile ID based on request or hypervisor configuration
	var networkProfileID int
	var firewallRulesets []int
	if config.Network != nil {
		if config.Network.ProfileID > 0 {
			// Use network profile from request
			networkProfileID = config.Network.ProfileID
			logger.Debugf("Using network profile ID from request: %d", networkProfileID)
		}
		if len(config.Network.FirewallRulesets) > 0 {
			firewallRulesets = config.Network.FirewallRulesets
		}
	}

	// If no network profile specified in request, try to get from hypervisor
	if networkProfileID == 0 {
		// First try primary network
		for _, network := range selectedHypervisor.Networks {
			if network.Primary {
				networkProfileID = network.ID
				logger.Debugf("Using primary network profile ID from hypervisor: %d", networkProfileID)
				break
			}
		}

		// If no primary network, try default network
		if networkProfileID == 0 {
			for _, network := range selectedHypervisor.Networks {
				if network.Default {
					networkProfileID = network.ID
					logger.Debugf("Using default network profile ID from hypervisor: %d", networkProfileID)
					break
				}
			}
		}
	}

	if networkProfileID == 0 {
		return nil, fmt.Errorf("no valid network profile found for hypervisor %d", selectedHypervisor.ID)
	}

	// Use default firewall rulesets if none specified
	if len(firewallRulesets) == 0 {
		firewallRulesets = []int{2, 3} // Default Celestia App and Node rulesets
	}

	// Calculate total disk size including volumes
	totalDiskSize := pkg.Disk
	for _, vol := range config.Volumes {
		totalDiskSize += vol.SizeGB
	}

	// Validate volumes if any
	if err := ValidateVolumes(config.Volumes); err != nil {
		logger.Errorf("Invalid volume configuration: %v (volumes: %+v)", err, config.Volumes)
		return nil, fmt.Errorf("invalid volume configuration: %w", err)
	}
	logger.Debugf("Volumes validated: %+v", config.Volumes)

	// Convert SSH keys from strings to IDs
	var sshKeyIDs []string
	for _, key := range config.SSHKeys {
		logger.Debugf("Processing SSH key: %s", key)
		// TODO: Implement SSH key ID lookup
		sshKeyIDs = append(sshKeyIDs, key)
	}

	// Create server request
	createReq := &types.ServerCreateRequest{
		Name:                 name,
		Memory:               pkg.Memory,
		Disk:                 totalDiskSize,
		CPU:                  pkg.CPU,
		Image:                template,
		HypervisorID:         selectedHypervisor.ID,
		HypervisorGroup:      selectedHypervisor.Group.Name,
		SSHKeys:              sshKeyIDs,
		Description:          p.generateVolumeDescription(config.Volumes),
		UserData:             p.generateVolumeUserData(config.Volumes),
		PackageID:            1, // Set a default package ID
		FirewallRulesets:     firewallRulesets,
		IPv4:                 1,                // Request IPv4 address
		Traffic:              1000,             // Set default traffic limit in GB
		NetworkSpeedInbound:  1000,             // Set default network speed in Mbps
		NetworkSpeedOutbound: 1000,             // Set default network speed in Mbps
		UserID:               1,                // Set the owner ID directly
		NetworkProfile:       networkProfileID, // Use the selected network profile
		StorageProfile:       1,                // Set a default storage profile
	}

	// Update network speeds if specified in request
	if config.Network != nil && config.Network.Bandwidth > 0 {
		createReq.NetworkSpeedInbound = config.Network.Bandwidth
		createReq.NetworkSpeedOutbound = config.Network.Bandwidth
		logger.Debugf("Using network bandwidth from request: %d Mbps", config.Network.Bandwidth)
	}

	// Create the server
	logger.Debugf("Server creation request prepared: %+v", createReq)
	server, _, err := p.client.Servers().Create(ctx, createReq)
	if err != nil {
		logger.Errorf("Failed to create server: %v", err)
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	// Return instance info
	return []types.InstanceInfo{
		{
			ID:     fmt.Sprintf("%d", server.ID),
			Name:   server.Name,
			Status: server.Status,
			IP:     server.IPAddress,
		},
	}, nil
}

// GetInstanceStatus returns the status of an instance
func (p *VirtFusionProvider) GetInstanceStatus(ctx context.Context, id int) (models.InstanceStatus, error) {
	logger.Debugf("Getting instance status: id=%d", id)

	server, _, err := p.client.Servers().Get(ctx, id)
	if err != nil {
		logger.Errorf("Failed to get server status: %v", err)
		return models.InstanceStatusUnknown, fmt.Errorf("failed to get server status: %w", err)
	}

	// Map server status to instance status
	var status models.InstanceStatus
	switch server.Status {
	case "ready":
		status = models.InstanceStatusReady
	case "pending":
		status = models.InstanceStatusPending
	case "provisioning":
		status = models.InstanceStatusProvisioning
	case "terminated":
		status = models.InstanceStatusTerminated
	default:
		status = models.InstanceStatusUnknown
	}

	logger.Debugf("Instance status retrieved: %s", status)
	return status, nil
}

// ListInstances returns a list of all instances
func (p *VirtFusionProvider) ListInstances(ctx context.Context) ([]types.InstanceInfo, error) {
	logger.Debug("Listing all instances")

	servers, _, err := p.client.Servers().List(ctx)
	if err != nil {
		logger.Errorf("Failed to list servers: %v", err)
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	var instances []types.InstanceInfo
	for _, server := range servers {
		instance := types.InstanceInfo{
			ID:       fmt.Sprintf("%d", server.ID),
			Name:     server.Name,
			Status:   server.Status,
			IP:       server.IPAddress,
			PublicIP: server.IPAddress,
		}
		instances = append(instances, instance)
	}

	logger.Debugf("Listed %d instances", len(instances))
	return instances, nil
}

// DeleteInstance deletes an instance
func (p *VirtFusionProvider) DeleteInstance(ctx context.Context, projectID, instanceID string) error {
	logger.Debugf("Deleting instance: projectID=%s, instanceID=%s", projectID, instanceID)

	id, err := strconv.Atoi(instanceID)
	if err != nil {
		logger.Errorf("Invalid instance ID format: %v", err)
		return fmt.Errorf("invalid instance ID format: %w", err)
	}

	_, err = p.client.Servers().Delete(ctx, id)
	if err != nil {
		logger.Errorf("Failed to delete server: %v", err)
		return fmt.Errorf("failed to delete server: %w", err)
	}

	logger.Debugf("Instance deleted successfully: projectID=%s, instanceID=%s", projectID, instanceID)
	return nil
}

// RecoverInstance attempts to recover a failed instance
func (p *VirtFusionProvider) RecoverInstance(ctx context.Context, id int) error {
	logger.Debugf("Attempting to recover instance: id=%d", id)

	// Get the current server state
	server, _, err := p.client.Servers().Get(ctx, id)
	if err != nil {
		logger.Errorf("Failed to get server info: %v", err)
		return fmt.Errorf("failed to get server info: %w", err)
	}

	// Create new server with same configuration
	createReq := &types.ServerCreateRequest{
		Name:        server.Name,
		Memory:      server.Memory,
		CPU:         server.CPU,
		Disk:        server.Disk,
		Description: server.Description,
		PackageID:   1, // Use default package
	}

	newServer, _, err := p.client.Servers().Create(ctx, createReq)
	if err != nil {
		logger.Errorf("Failed to create recovery server: %v", err)
		return fmt.Errorf("failed to create recovery server: %w", err)
	}

	logger.Debugf("Recovery instance created successfully: old_id=%d, new_id=%d", id, newServer.ID)
	return nil
}

// generateVolumeDescription generates a description for the server based on volume configuration
func (p *VirtFusionProvider) generateVolumeDescription(volumes []types.VolumeConfig) string {
	if len(volumes) == 0 {
		return ""
	}
	return fmt.Sprintf("Server configured with %d additional volumes", len(volumes))
}

// generateVolumeUserData generates user data for volume configuration
func (p *VirtFusionProvider) generateVolumeUserData(volumes []types.VolumeConfig) string {
	if len(volumes) == 0 {
		return ""
	}
	// TODO: Implement volume configuration in user data
	return ""
}

// validateVolumes validates the volume configuration
func (p *VirtFusionProvider) validateVolumes(volumes []types.VolumeConfig) error {
	for i, vol := range volumes {
		if vol.SizeGB <= 0 {
			return fmt.Errorf("invalid size for volume %d: %d", i, vol.SizeGB)
		}
		if vol.Name == "" {
			return fmt.Errorf("volume name is required for volume %d", i)
		}
		if vol.MountPoint == "" {
			return fmt.Errorf("mount point is required for volume %d", i)
		}
	}
	return nil
}

// calculateTotalDiskSize calculates the total disk size including volumes
func (p *VirtFusionProvider) calculateTotalDiskSize(baseDisk int, volumes []types.VolumeConfig) int {
	total := baseDisk
	for _, vol := range volumes {
		total += vol.SizeGB
	}
	return total
}
