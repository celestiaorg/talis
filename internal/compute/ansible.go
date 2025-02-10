package compute

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// ansibleDebug is the verbose mode for ansible
	ansibleDebug = false
)

// CreateInventory creates the inventory file with all instances
func CreateInventory(instances map[string]string, keyPath string) error {
	// Get base name from first instance (they all share the same base)
	var baseName string
	for name := range instances {
		baseName = strings.Split(name, "-")[0]
		break
	}

	fmt.Printf("🎭 Creating inventory for %s...\n", baseName)

	// Create inventory path with base name
	inventoryPath := fmt.Sprintf("ansible/inventory_%s.ini", baseName)

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
func RunAnsiblePlaybook(inventoryName string) error {
	fmt.Println("🎭 Running Ansible playbook for all instances in parallel...")

	// Create inventory path with name
	inventoryPath := fmt.Sprintf("ansible/inventory_%s.ini", inventoryName)

	// Prepare command arguments
	args := []string{
		"-i", inventoryPath,
		// Set number of parallel processes (forks)
		"-f", "10",
	}

	// Add verbose flag if debug is enabled
	if ansibleDebug {
		args = append(args, "-vvv")
	}

	// Add playbook path
	args = append(args, "ansible/playbook.yml")

	// Run ansible-playbook command
	cmd := exec.Command("ansible-playbook", args...)

	// Disable host key checking
	env := os.Environ()
	env = append(env, "ANSIBLE_HOST_KEY_CHECKING=false")
	cmd.Env = env

	// Redirect output to stdout
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ansible playbook: %v", err)
	}

	return nil
}
