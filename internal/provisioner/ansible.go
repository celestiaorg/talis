// Package provisioner provides infrastructure provisioning functionality
package provisioner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/events"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/provisioner/config"
	"github.com/celestiaorg/talis/internal/types"
)

const (
	// AnsibleDebug is the verbose mode for ansible
	//
	//nolint:unused // Will be used in future implementation
	AnsibleDebug = false

	// PathToPlaybook is the path to the ansible main playbook
	PathToPlaybook = "ansible/main.yml"

	// DefaultSSHUser is the default user for SSH connections
	DefaultSSHUser = "root"

	// DefaultSSHKeyPath is the default path to the SSH private key
	DefaultSSHKeyPath = "$HOME/.ssh/id_rsa"

	// SSHCommonArgs are the common SSH arguments for ansible
	SSHCommonArgs = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

	// InventoryBasePath is the base path for ansible inventory files
	InventoryBasePath = "ansible/inventory"
)

// AnsibleProvisioner implements provisioning using Ansible
type AnsibleProvisioner struct {
	config *config.AnsibleConfig
}

// NewAnsibleProvisioner creates a new Ansible provisioner
func NewAnsibleProvisioner(config *config.AnsibleConfig) (*AnsibleProvisioner, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &AnsibleProvisioner{
		config: config,
	}, nil
}

// Configure implements Provisioner.Configure
func (a *AnsibleProvisioner) Configure(_ context.Context, instances []types.InstanceInfo) error {
	logger.InfoWithFields("Configuring instances with Ansible", map[string]interface{}{
		"job_id":         a.config.JobID,
		"owner_id":       a.config.OwnerID,
		"instance_count": len(instances),
	})

	// If no instances provided, try to get them from DB
	if len(instances) == 0 {
		logger.Info("No instances provided, requesting inventory from database...")
		events.Publish(events.Event{
			Type:    events.EventInventoryRequested,
			JobID:   a.config.JobID,
			JobName: a.config.JobName,
			OwnerID: a.config.OwnerID,
		})
		return nil
	}

	// Configure hosts in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(instances))

	for _, instance := range instances {
		wg.Add(1)
		go func(inst types.InstanceInfo) {
			defer wg.Done()
			if err := a.waitForSSH(inst.PublicIP); err != nil {
				errChan <- fmt.Errorf("failed waiting for SSH on instance %s: %w", inst.Name, err)
				return
			}
		}(instance)
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

	// Create inventory file with the user's SSH key
	sshKeyPath := os.ExpandEnv(a.config.SSHKeyPath)
	instanceMap := make(map[string]string)
	for _, instance := range instances {
		instanceMap[instance.Name] = instance.PublicIP
	}

	if err := a.CreateInventory(instanceMap, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	// Run Ansible playbook
	logger.Info("📝 Running provisioning playbook...")
	inventoryPath := filepath.Join(a.config.InventoryBasePath, fmt.Sprintf("inventory_%s_ansible.ini", a.config.JobID))
	if err := a.RunPlaybook(inventoryPath); err != nil {
		return fmt.Errorf("failed to run playbook: %w", err)
	}

	logger.Info("✅ Provisioning completed successfully")
	return nil
}

// waitForSSH waits for SSH to be available on the given host
func (a *AnsibleProvisioner) waitForSSH(host string) error {
	logger.Infof("⏳ Waiting for SSH to be available on %s...", host)

	for i := 0; i < 30; i++ {
		args := []string{
			"-i", DefaultSSHKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("%s@%s", DefaultSSHUser, host),
			"echo 'SSH is ready'",
		}

		// #nosec G204 -- command arguments are constructed from validated inputs
		checkCmd := exec.Command("ssh", args...)
		if err := checkCmd.Run(); err == nil {
			logger.Infof("✅ SSH connection established to %s", host)
			return nil
		}

		if i == 29 {
			return fmt.Errorf("timeout waiting for SSH to be ready on %s after 5 minutes", host)
		}

		logger.Infof("  Retrying SSH connection to %s in 10 seconds... (%d/30)", host, i+1)
		time.Sleep(10 * time.Second)
	}

	return nil
}

// CreateInventory implements Provisioner.CreateInventory
func (a *AnsibleProvisioner) CreateInventory(instances map[string]string, keyPath string) error {
	logger.Info("🎭 Creating inventory for job ", a.config.JobID, "...")

	// If no instances provided, return error
	if len(instances) == 0 {
		return fmt.Errorf("no instances provided")
	}

	inventoryPath := filepath.Join(a.config.InventoryBasePath, fmt.Sprintf("inventory_%s_ansible.ini", a.config.JobID))

	// Ensure the inventory directory exists with secure permissions
	if err := os.MkdirAll(filepath.Dir(inventoryPath), 0750); err != nil {
		return fmt.Errorf("failed to create inventory directory: %w", err)
	}

	// Create the inventory file with secure permissions
	// #nosec G304 -- inventoryPath is constructed from validated inputs
	f, err := os.OpenFile(inventoryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create inventory file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Error("Failed to close inventory file:", err)
		}
	}()

	// Write header with SSH settings and variables first
	header := fmt.Sprintf("[all:vars]\nansible_ssh_common_args='%s'\n\n[all]\n", SSHCommonArgs)
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("failed to write inventory header: %w", err)
	}

	// Write all instances
	for name, ip := range instances {
		line := fmt.Sprintf("%s ansible_host=%s ansible_user=%s ansible_ssh_private_key_file=%s\n",
			name, ip, a.config.SSHUser, keyPath)

		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write instance to inventory: %w", err)
		}
	}

	logger.Info("✅ Created inventory file at ", inventoryPath)
	return nil
}

// RunPlaybook implements Provisioner.RunPlaybook
func (a *AnsibleProvisioner) RunPlaybook(inventoryPath string) error {
	logger.InfoWithFields("Running Ansible playbook", map[string]interface{}{
		"job_id":         a.config.JobID,
		"inventory_path": inventoryPath,
		"playbook_path":  a.config.PlaybookPath,
	})

	// Validate paths exist
	if _, err := os.Stat(inventoryPath); err != nil {
		return fmt.Errorf("inventory file not found: %w", err)
	}

	if _, err := os.Stat(a.config.PlaybookPath); err != nil {
		return fmt.Errorf("playbook file not found: %w", err)
	}

	// Prepare command arguments
	args := []string{
		"-i", inventoryPath,
		// Run serially for consistency
		"--forks", "1",
	}

	// Add verbosity if debug mode is enabled
	if AnsibleDebug {
		args = append(args, "-vvv")
	}

	// Add playbook path
	args = append(args, a.config.PlaybookPath)

	// Run ansible-playbook command
	// #nosec G204 -- command arguments are constructed from validated inputs
	cmd := exec.Command("ansible-playbook", args...)

	// Set environment variables
	env := os.Environ()
	env = append(env, "ANSIBLE_HOST_KEY_CHECKING=false")
	env = append(env, "ANSIBLE_RETRY_FILES_ENABLED=false")
	cmd.Env = env

	// Redirect output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ansible playbook (check output above for details): %w", err)
	}

	return nil
}
