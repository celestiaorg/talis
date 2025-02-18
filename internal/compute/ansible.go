package compute

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	// ansibleDebug is the verbose mode for ansible
	ansibleDebug = false
)

// AnsibleConfigurator implements the Provisioner interface
type AnsibleConfigurator struct {
	// inventoryID is the unique identifier for the current run
	inventoryID string
	// instances keeps track of all instances to be configured
	instances map[string]string
	// sshKeyPath stores the SSH key path for all instances
	sshKeyPath string
}

// NewAnsibleConfigurator creates a new Ansible configurator
func NewAnsibleConfigurator() *AnsibleConfigurator {
	timestamp := time.Now().Format("20060102-150405")
	inventoryID := fmt.Sprintf("talis-%s", timestamp)
	return &AnsibleConfigurator{
		inventoryID: inventoryID,
		instances:   make(map[string]string),
	}
}

// CreateInventory creates the inventory file with all instances
func (a *AnsibleConfigurator) CreateInventory(instances map[string]string, keyPath string) error {
	fmt.Printf("🎭 Creating inventory %s...\n", a.inventoryID)

	// Create inventory path with base name
	inventoryPath := fmt.Sprintf("ansible/inventory_%s_ansible.ini", a.inventoryID)

	// Ensure ansible directory exists
	if err := os.MkdirAll("ansible", 0755); err != nil {
		return err
	}

	// Create inventory file with header
	f, err := os.Create(inventoryPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header
	if _, err := f.WriteString("[all]\n"); err != nil {
		return err
	}

	// Write all instances
	for name, ip := range instances {
		line := fmt.Sprintf("%s ansible_host=%s ansible_user=root ansible_ssh_private_key_file=%s\n",
			name, ip, keyPath)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// RunAnsiblePlaybook runs the Ansible playbook for all instances in parallel
func (a *AnsibleConfigurator) RunAnsiblePlaybook(inventoryName string) error {
	fmt.Println("🎭 Running Ansible playbook...")

	// Create inventory path with name
	inventoryPath := fmt.Sprintf("ansible/inventory_%s_ansible.ini", a.inventoryID)

	// Prepare command arguments
	args := []string{
		"-i", inventoryPath,
		// Run serially
		"--forks", "1",
		// // Add extra verbosity
		// "-vvv",
	}

	// Add playbook path
	args = append(args, "ansible/playbook.yml")

	// Run ansible-playbook command
	cmd := exec.Command("ansible-playbook", args...)

	// Disable host key checking
	env := os.Environ()
	env = append(env, "ANSIBLE_HOST_KEY_CHECKING=false")
	env = append(env, "ANSIBLE_RETRY_FILES_ENABLED=false")
	cmd.Env = env

	// Redirect output to stdout
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ansible playbook (check output above for details): %v", err)
	}

	return nil
}

// ConfigureHost implements the Provisioner interface
func (a *AnsibleConfigurator) ConfigureHost(host string, sshKeyPath string) error {
	fmt.Printf("🔧 Configuring host %s with Ansible...\n", host)

	// Wait for SSH to be available
	fmt.Printf("⏳ Waiting for SSH to be available on %s...\n", host)
	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("ssh",
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("root@%s", host),
			"echo 'SSH is ready'")

		if err := checkCmd.Run(); err == nil {
			fmt.Printf("✅ SSH connection established to %s\n", host)
			break
		}

		if i == 29 {
			return fmt.Errorf("timeout waiting for SSH to be ready")
		}

		fmt.Printf("  Retrying in 10 seconds... (%d/30)\n", i+1)
		time.Sleep(10 * time.Second)
	}

	// Store instance and SSH key path
	instanceName := fmt.Sprintf("talis-%d", len(a.instances))
	a.instances[instanceName] = host
	a.sshKeyPath = sshKeyPath

	// Create/update inventory with all instances
	if err := a.CreateInventory(a.instances, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create inventory: %v", err)
	}

	// Run Ansible playbook
	if err := a.RunAnsiblePlaybook(a.inventoryID); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %v", err)
	}

	return nil
}
