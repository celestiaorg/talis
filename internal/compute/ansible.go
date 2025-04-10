package compute

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/logger"
)

const (
	// ansibleDebug is the verbose mode for ansible
	//
	//nolint:unused // Will be used in future implementation
	ansibleDebug = false

	// pathToPlaybook is the path to the ansible main playbook
	pathToPlaybook = "ansible/main.yml"

	// defaultSSHUser is the default user for SSH connections
	defaultSSHUser = "root"

	// defaultSSHKeyPath is the default path to the SSH private key
	defaultSSHKeyPath = "$HOME/.ssh/id_rsa"

	// sshCommonArgs are the common SSH arguments for ansible
	sshCommonArgs = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

	// inventoryBasePath is the base path for ansible inventory files
	inventoryBasePath = "ansible/inventory"
)

// InstanceRepository defines the interface for instance operations
type InstanceRepository interface {
	GetByJobID(ctx context.Context, ownerID, jobID uint) ([]models.Instance, error)
}

// AnsibleConfigurator implements the Provisioner interface
type AnsibleConfigurator struct {
	// jobID is the unique identifier for the current job
	jobID string
	// instances keeps track of all instances to be configured
	instances map[string]string
	// sshKeyPath stores the SSH key path for all instances
	sshKeyPath string
	// mutex protects the instances map
	mutex sync.Mutex
	// instanceRepo provides access to instance data
	instanceRepo InstanceRepository
	// ownerID is the ID of the owner of the instances
	ownerID uint
}

// NewAnsibleConfigurator creates a new Ansible configurator
func NewAnsibleConfigurator(jobID string, instanceRepo InstanceRepository) *AnsibleConfigurator {
	return &AnsibleConfigurator{
		jobID:        jobID,
		instances:    make(map[string]string),
		instanceRepo: instanceRepo,
	}
}

// SetOwnerID sets the owner ID for the configurator
func (a *AnsibleConfigurator) SetOwnerID(ownerID uint) {
	a.ownerID = ownerID
}

// createInventoryFromDB creates the inventory file using database instances
func (a *AnsibleConfigurator) createInventoryFromDB(ctx context.Context) error {
	// Convert jobID to uint for repository
	var jobIDUint uint
	_, err := fmt.Sscanf(a.jobID, "job-%d", &jobIDUint)
	if err != nil {
		return fmt.Errorf("failed to parse job ID: %w", err)
	}

	// Fetch instances from database
	instances, err := a.instanceRepo.GetByJobID(ctx, a.ownerID, jobIDUint)
	if err != nil {
		return fmt.Errorf("failed to fetch instances from database: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("no instances found for job %s", a.jobID)
	}

	// Create inventory path with base name
	inventoryPath := fmt.Sprintf("%s/inventory_%s_ansible.ini", inventoryBasePath, a.jobID)

	// Create ansible inventory directory with secure permissions
	if err := os.MkdirAll(inventoryBasePath, 0750); err != nil {
		return fmt.Errorf("failed to create ansible inventory directory: %w", err)
	}

	// Create inventory file with secure permissions
	// #nosec G304 -- inventory path is constructed from validated job ID
	f, err := os.OpenFile(inventoryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create inventory file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close inventory file: %w", closeErr)
		}
	}()

	// Write header with SSH settings and variables first
	header := fmt.Sprintf("[all:vars]\nansible_ssh_common_args='%s'\n\n[all]\n", sshCommonArgs)
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("failed to write inventory header: %w", err)
	}

	// Write all instances
	for _, instance := range instances {
		// Skip deleted instances
		if instance.DeletedAt.Valid {
			logger.Debugf("Skipping deleted instance: %s", instance.Name)
			continue
		}

		// Always use the provided key path and root user for SSH
		expandedKeyPath := os.ExpandEnv(defaultSSHKeyPath)
		line := fmt.Sprintf("%s ansible_host=%s ansible_user=%s ansible_ssh_private_key_file=%s",
			instance.Name, instance.PublicIP, defaultSSHUser, expandedKeyPath)
		line += "\n"

		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write instance to inventory: %w", err)
		}

		// Store in internal map for compatibility
		a.mutex.Lock()
		a.instances[instance.Name] = instance.PublicIP
		a.mutex.Unlock()
	}

	logger.Info("✅ Created inventory file at ", inventoryPath)
	return nil
}

// CreateInventory creates the inventory file with all instances
func (a *AnsibleConfigurator) CreateInventory(instances map[string]string, _ string) error {
	logger.Info("🎭 Creating inventory for job ", a.jobID, "...")

	// If we have a repository and owner ID, try to use the database first
	if a.instanceRepo != nil && a.ownerID > 0 {
		if err := a.createInventoryFromDB(context.Background()); err == nil {
			return nil
		}
		// If database fails, fall back to using provided instances
		logger.Warn("⚠️ Failed to create inventory from database, falling back to provided instances")
	}

	// If no instances provided and database failed, return error
	if len(instances) == 0 {
		return fmt.Errorf("no instances provided and database query failed")
	}

	// Create inventory path with base name
	inventoryPath := fmt.Sprintf("%s/inventory_%s_ansible.ini", inventoryBasePath, a.jobID)

	// Create ansible inventory directory with secure permissions
	if err := os.MkdirAll(inventoryBasePath, 0750); err != nil {
		return fmt.Errorf("failed to create ansible inventory directory: %w", err)
	}

	// Create inventory file with secure permissions
	// #nosec G304 -- inventory path is constructed from validated job ID
	f, err := os.OpenFile(inventoryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create inventory file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close inventory file: %w", closeErr)
		}
	}()

	// Write header with SSH settings and variables first
	header := fmt.Sprintf("[all:vars]\nansible_ssh_common_args='%s'\n\n[all]\n", sshCommonArgs)
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("failed to write inventory header: %w", err)
	}

	// Write all instances
	for name, ip := range instances {
		// Always use the provided key path and root user for SSH
		expandedKeyPath := os.ExpandEnv(defaultSSHKeyPath)
		line := fmt.Sprintf("%s ansible_host=%s ansible_user=%s ansible_ssh_private_key_file=%s",
			name, ip, defaultSSHUser, expandedKeyPath)
		line += "\n"

		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write instance to inventory: %w", err)
		}
	}

	// Store instances in internal map
	a.mutex.Lock()
	for name, ip := range instances {
		a.instances[name] = ip
	}
	a.mutex.Unlock()

	logger.Info("✅ Created inventory file at ", inventoryPath)
	return nil
}

// RunAnsiblePlaybook runs the Ansible playbook for all instances in parallel
func (a *AnsibleConfigurator) RunAnsiblePlaybook(_ string) error {
	logger.Info("🎭 Running Ansible playbook...")

	// Create inventory path with name
	inventoryPath := fmt.Sprintf("%s/inventory_%s_ansible.ini", inventoryBasePath, a.jobID)

	// Prepare command arguments
	args := []string{
		"-i", inventoryPath,
		// Run serially
		"--forks", "1",
	}

	// Add verbosity if debug mode is enabled
	if ansibleDebug {
		args = append(args, "-vvv")
	}

	// Add playbook path
	args = append(args, pathToPlaybook)

	// Run ansible-playbook command
	// #nosec G204 -- command arguments are constructed from validated inputs
	cmd := exec.Command("ansible-playbook", args...)

	// Disable host key checking and known hosts file
	env := os.Environ()
	env = append(env, "ANSIBLE_HOST_KEY_CHECKING=false")
	env = append(env, "ANSIBLE_RETRY_FILES_ENABLED=false")
	cmd.Env = env

	// Redirect output to stdout
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ansible playbook (check output above for details): %w", err)
	}

	return nil
}

// ConfigureHost implements the Provisioner interface
func (a *AnsibleConfigurator) ConfigureHost(host string, sshKeyPath string) error {
	// Store instance and SSH key path
	a.mutex.Lock()
	instanceName := fmt.Sprintf("%s-%d", a.jobID, len(a.instances))
	a.instances[instanceName] = host
	a.sshKeyPath = sshKeyPath
	a.mutex.Unlock()

	logger.Infof("🔧 Configuring host %s (instance: %s)...", host, instanceName)

	// Wait for SSH to be available
	logger.Infof("⏳ Waiting for SSH to be available on %s...", host)
	for i := 0; i < 30; i++ {
		args := []string{
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("%s@%s", defaultSSHUser, host),
			"echo 'SSH is ready'",
		}

		// #nosec G204 -- command arguments are constructed from validated inputs
		checkCmd := exec.Command("ssh", args...)

		if err := checkCmd.Run(); err == nil {
			logger.Infof("✅ SSH connection established to %s", host)
			break
		}

		if i == 29 {
			return fmt.Errorf("timeout waiting for SSH to be ready on %s after 5 minutes", host)
		}

		logger.Infof("  Retrying SSH connection to %s in 10 seconds... (%d/30)", host, i+1)
		time.Sleep(10 * time.Second)
	}

	return nil
}

// ConfigureHosts configures multiple hosts in parallel
func (a *AnsibleConfigurator) ConfigureHosts(hosts []string, sshKeyPath string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			if err := a.ConfigureHost(h, sshKeyPath); err != nil {
				errChan <- fmt.Errorf("failed to configure host %s: %w", h, err)
			}
		}(host)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to configure some hosts: %v", errors)
	}

	// Create inventory with empty instances to trigger database lookup
	if err := a.CreateInventory(nil, ""); err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	// Run Ansible playbook
	if err := a.RunAnsiblePlaybook(a.jobID); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %w", err)
	}

	return nil
}
