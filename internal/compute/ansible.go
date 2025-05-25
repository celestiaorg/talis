package compute

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

const (
	// ansibleDir is the base directory for all ansible files
	ansibleDir = "ansible"

	// ansibleDebug is the verbose mode for ansible
	//
	//nolint:unused // Will be used in future implementation
	ansibleDebug = false

	// pathToPlaybook is the path to the ansible main playbook
	pathToPlaybook = ansibleDir + "/main.yml"

	// ansibleKeyPath is the fixed path where the SSH key will be stored
	ansibleKeyPath = ansibleDir + "/talis_ssh_key"
)

// keyFileMutex protects access to the SSH key file to prevent race conditions
var keyFileMutex sync.RWMutex

// AnsibleConfigurator implements the Provisioner interface
type AnsibleConfigurator struct {
	// jobID is the unique identifier for the current job
	jobID string
	// instances keeps track of all instances to be configured
	instances map[string]string
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

// GetSSHKeyFromEnv returns the SSH key from the environment variable
func (a *AnsibleConfigurator) GetSSHKeyFromEnv() string {
	return os.Getenv(constants.EnvTalisSSHKey)
}

// EnsureSSHKeyFile ensures the SSH key is written to a file and returns the path
// If the file already exists with the correct content, it won't be overwritten
// This function is goroutine-safe and optimized for concurrent reads.
func (a *AnsibleConfigurator) EnsureSSHKeyFile() (string, error) {
	// Get key content before acquiring any lock
	keyContent := a.GetSSHKeyFromEnv()
	if keyContent == "" {
		return "", fmt.Errorf("no SSH key found in environment variable %s", constants.EnvTalisSSHKey)
	}

	// Create ansible directory with secure permissions if it doesn't exist
	// This is safe to do without locking
	if err := os.MkdirAll(ansibleDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create ansible directory: %w", err)
	}

	// First acquire a read lock to check if the file exists with correct content
	keyFileMutex.RLock()
	existingContent, err := os.ReadFile(ansibleKeyPath)
	if err == nil && string(existingContent) == keyContent {
		// File exists with correct content, we're done
		keyFileMutex.RUnlock()
		logger.Debugf("SSH key file already exists with correct content at %s", ansibleKeyPath)
		return ansibleKeyPath, nil
	}
	// Release the read lock so we can acquire a write lock
	keyFileMutex.RUnlock()

	// Acquire a write lock for writing the file
	keyFileMutex.Lock()
	defer keyFileMutex.Unlock()

	// Check again in case another goroutine wrote the file while we were waiting
	existingContent, err = os.ReadFile(ansibleKeyPath)
	if err == nil && string(existingContent) == keyContent {
		logger.Debugf("SSH key file already exists with correct content at %s", ansibleKeyPath)
		return ansibleKeyPath, nil
	}

	// Write the key with secure permissions
	if err := os.WriteFile(ansibleKeyPath, []byte(keyContent), 0600); err != nil {
		return "", fmt.Errorf("failed to write SSH key to file: %w", err)
	}

	logger.Debugf("Created SSH key file at %s", ansibleKeyPath)
	return ansibleKeyPath, nil
}

// CreateInventory creates the inventory file directly from InstanceRequest and returns the inventory path file
func (a *AnsibleConfigurator) CreateInventory(instance *types.InstanceRequest) (string, error) {
	fmt.Printf("⚙️ Creating Ansible inventory for job %s...\n", a.jobID)

	if instance == nil {
		logger.Warnf("No instance provided to CreateInventory for job %s, skipping inventory creation.", a.jobID)
		return "", nil // Nothing to do
	}

	// Ensure SSH key file exists
	keyPath, err := a.EnsureSSHKeyFile()
	if err != nil {
		return "", err
	}

	// Create inventory path with base name
	inventoryPath := filepath.Join(ansibleDir, fmt.Sprintf("inventory_%s_ansible.ini", a.jobID))

	// Create inventory file with secure permissions
	// #nosec G304 -- inventory path is constructed from validated job ID
	f, err := os.OpenFile(inventoryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create inventory file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close inventory file: %w", closeErr)
		}
	}()

	// Write header with SSH settings and variables first
	header := "[all:vars]\nansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'\n\n[all]\n"
	if _, err := f.WriteString(header); err != nil {
		return "", fmt.Errorf("failed to write inventory header: %w", err)
	}

	// Write all instances using data from InstanceRequest
	// Start base line with name, host, user, and key
	line := fmt.Sprintf("%s ansible_host=%s ansible_user=root ansible_ssh_private_key_file=%s",
		instance.PublicIP, instance.PublicIP, keyPath)

	// Add payload variables directly from InstanceRequest
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
		// Use PublicIP in error message as well
		return "", fmt.Errorf("failed to write instance with IP '%s' to inventory: %w", instance.PublicIP, err)
	}

	fmt.Printf("✅ Created inventory file at %s\n", inventoryPath)
	return inventoryPath, nil
}

// RunAnsiblePlaybook runs the Ansible playbook for all instances in parallel
func (a *AnsibleConfigurator) RunAnsiblePlaybook(inventoryPath string, tags []string) error {
	fmt.Println("🎭 Running Ansible playbook...")

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

	if len(tags) > 0 {
		tagsStr := strings.Join(tags, ",")
		args = append(args, "--tags", tagsStr)
	}

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
func (a *AnsibleConfigurator) ConfigureHost(ctx context.Context, host string) error {
	// Store instance
	a.mutex.Lock()
	instanceName := fmt.Sprintf("%s-%d", a.jobID, len(a.instances))
	a.instances[instanceName] = host
	a.mutex.Unlock()

	// Ensure SSH key file exists
	keyPath, err := a.EnsureSSHKeyFile()
	if err != nil {
		return err
	}

	fmt.Printf("🔧 Configuring host %s (instance: %s)...\n", host, instanceName)

	// Wait for SSH to be available
	fmt.Printf("⏳ Waiting for SSH to be available on %s...\n", host)
	for i := 0; i < 30; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for SSH to be available on %s", host)
		default:
			// continue
		}

		args := []string{
			"-i", keyPath,
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
//
// NOTE: this isn't a create name since all it is really doing is ensuring SSH readiness
func (a *AnsibleConfigurator) ConfigureHosts(ctx context.Context, hosts []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(hosts))

	fmt.Printf("🔧 Ensuring SSH readiness for %d hosts...\n", len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			if err := a.ConfigureHost(ctx, h); err != nil {
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
