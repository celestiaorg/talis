package compute

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

const (
	// ansibleDebug is the verbose mode for ansible
	//
	//nolint:unused // Will be used in future implementation
	ansibleDebug = false

	// pathToPlaybook is the path to the ansible main playbook
	pathToPlaybook = "ansible/main.yml"
)

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
}

// NewAnsibleConfigurator creates a new Ansible configurator
func NewAnsibleConfigurator(jobID string) *AnsibleConfigurator {
	return &AnsibleConfigurator{
		jobID:     jobID,
		instances: make(map[string]string),
	}
}

// CreateInventory creates the inventory file directly from InstanceInfo
func (a *AnsibleConfigurator) CreateInventory(instances []types.InstanceInfo, sshKeyPath string) error {
	fmt.Printf("⚙️ Creating Ansible inventory for job %s...\n", a.jobID)

	if len(instances) == 0 {
		logger.Warnf("No instances provided to CreateInventory for job %s, skipping inventory creation.", a.jobID)
		return nil // Nothing to do
	}

	// Create inventory path with base name
	inventoryPath := fmt.Sprintf("ansible/inventory_%s_ansible.ini", a.jobID)

	// Create ansible directory with secure permissions
	if err := os.MkdirAll("ansible", 0750); err != nil {
		return fmt.Errorf("failed to create ansible directory: %w", err)
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
	header := "[all:vars]\nansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'\n\n[all]\n"
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("failed to write inventory header: %w", err)
	}

	// Write all instances using data from InstanceInfo
	for _, instance := range instances {
		// Start base line with name, host, user, and key
		line := fmt.Sprintf("%s ansible_host=%s ansible_user=root ansible_ssh_private_key_file=%s",
			instance.Name, instance.PublicIP, sshKeyPath)

		// Add payload variables directly from InstanceInfo
		payloadPresent := instance.PayloadPath != ""
		line += fmt.Sprintf(" payload_present=%t", payloadPresent)
		if payloadPresent {
			destFilename := filepath.Base(instance.PayloadPath)
			destPath := filepath.Join("/root", destFilename) // Ensure consistent path joining
			// Quote string values for safety in inventory
			line += fmt.Sprintf(" payload_src_path=\"%s\"", instance.PayloadPath)
			line += fmt.Sprintf(" payload_dest_path=\"%s\"", destPath)
			line += fmt.Sprintf(" payload_execute=%t", instance.ExecutePayload)
		}

		line += "\n"

		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write instance '%s' to inventory: %w", instance.Name, err)
		}
	}

	fmt.Printf("✅ Created inventory file at %s with %d instances\n", inventoryPath, len(instances))
	return nil
}

// RunAnsiblePlaybook runs the Ansible playbook for all instances in parallel
func (a *AnsibleConfigurator) RunAnsiblePlaybook(inventoryName string) error {
	fmt.Println("🎭 Running Ansible playbook...")

	// Create inventory path with name
	inventoryPath := fmt.Sprintf("ansible/inventory_%s_ansible.ini", inventoryName)

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

	fmt.Println("✅ Ansible playbook completed successfully")
	return nil
}

// ConfigureHost implements the Provisioner interface
//
// NOTE: this isn't a create name since all it is really doing is ensuring SSH readiness
func (a *AnsibleConfigurator) ConfigureHost(host string, sshKeyPath string) error {
	// Store instance and SSH key path
	a.mutex.Lock()
	instanceName := fmt.Sprintf("%s-%d", a.jobID, len(a.instances))
	a.instances[instanceName] = host
	a.sshKeyPath = sshKeyPath
	a.mutex.Unlock()

	fmt.Printf("🔧 Configuring host %s (instance: %s)...\n", host, instanceName)

	// Wait for SSH to be available
	fmt.Printf("⏳ Waiting for SSH to be available on %s...\n", host)
	for i := 0; i < 30; i++ {
		args := []string{
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("root@%s", host),
			"echo 'SSH is ready'",
		}

		// #nosec G204 -- command arguments are constructed from validated inputs
		checkCmd := exec.Command("ssh", args...)

		if err := checkCmd.Run(); err == nil {
			fmt.Printf("✅ SSH connection established to %s\n", host)
			break
		}

		if i == 29 {
			return fmt.Errorf("timeout waiting for SSH to be ready on %s after 5 minutes", host)
		}

		fmt.Printf("  Retrying SSH connection to %s in 10 seconds... (%d/30)\n", host, i+1)
		time.Sleep(10 * time.Second)
	}

	return nil
}

// ConfigureHosts ensures SSH readiness for multiple hosts in parallel.
// It no longer creates the inventory or runs the playbook.
//
// NOTE: this isn't a create name since all it is really doing is ensuring SSH readiness
func (a *AnsibleConfigurator) ConfigureHosts(hosts []string, sshKeyPath string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(hosts))

	fmt.Printf("🔧 Ensuring SSH readiness for %d hosts...\n", len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			if err := a.ConfigureHost(h, sshKeyPath); err != nil {
				errChan <- fmt.Errorf("failed to ensure SSH readiness for host %s: %w", h, err)
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
		return fmt.Errorf("failed to ensure SSH readiness for some hosts: %v", errors)
	}

	fmt.Printf("✅ SSH readiness confirmed for all hosts.\n")
	return nil
}
