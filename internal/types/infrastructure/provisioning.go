package infrastructure

import (
	"fmt"
	"os"
	"sync"

	"github.com/celestiaorg/talis/internal/types"
)

// RunProvisioning applies Ansible configuration to all instances
func (i *Infrastructure) RunProvisioning(instances []types.InstanceInfo) error {
	// Check if any instance requires provisioning
	needsProvisioning := false
	for _, inst := range i.instances {
		if inst.Provision {
			needsProvisioning = true
			break
		}
	}

	if !needsProvisioning {
		fmt.Println("⏭️ Skipping provisioning as requested")
		return nil
	}

	fmt.Println("⚙️ Starting Ansible provisioning...")

	// Create inventory file for all instances
	instanceMap := make(map[string]string)
	for _, instance := range instances {
		instanceMap[instance.Name] = instance.PublicIP
	}

	// Create inventory file with the user's SSH key
	sshKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	if err := i.provisioner.CreateInventory(instanceMap, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	// Configure hosts in parallel
	errChan := make(chan error, len(instances))
	var wg sync.WaitGroup

	for _, instance := range instances {
		wg.Add(1)
		go func(inst types.InstanceInfo) {
			defer wg.Done()
			if err := i.provisionInstance(inst); err != nil {
				errChan <- fmt.Errorf("failed to provision %s: %w", inst.Name, err)
			}
		}(instance)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Run Ansible playbook
	fmt.Println("📝 Running Ansible playbook...")
	if err := i.provisioner.RunAnsiblePlaybook(i.jobID); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %w", err)
	}

	fmt.Println("✅ Ansible playbook completed successfully")
	return nil
}

// provisionInstance configures a single instance with Ansible
func (i *Infrastructure) provisionInstance(instance types.InstanceInfo) error {
	fmt.Printf("🔧 Starting provisioning for %s (%s)...\n", instance.Name, instance.PublicIP)

	// Use the user's SSH key path
	sshKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	if err := i.provisioner.ConfigureHost(instance.PublicIP, sshKeyPath); err != nil {
		return fmt.Errorf("failed to configure host: %w", err)
	}

	fmt.Printf("✅ Provisioning completed for %s\n", instance.Name)
	return nil
}
