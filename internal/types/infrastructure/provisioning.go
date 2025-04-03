package infrastructure

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/celestiaorg/talis/internal/compute"
)

// RunProvisioning applies Ansible configuration to all instances
func (i *Infrastructure) RunProvisioning(instances []InstanceInfo) error {
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
		instanceMap[instance.Name] = instance.IP
	}

	// Create inventory file with the user's SSH key
	sshKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	if err := i.provisioner.CreateInventory(instanceMap, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	// Set payloads for instances if provided
	for idx, instanceReq := range i.instances {
		if instanceReq.Payload != "" && idx < len(instances) {
			// Find the corresponding created instance
			var instanceName string
			if instanceReq.Name != "" {
				instanceName = instanceReq.Name
			} else {
				// Use the generated name
				for _, instance := range instances {
					if strings.Contains(instance.Name, fmt.Sprintf("%s-%d", i.name, idx)) {
						instanceName = instance.Name
						break
					}
				}
			}

			if instanceName != "" {
				if provisioner, ok := i.provisioner.(*compute.AnsibleConfigurator); ok {
					provisioner.SetPayload(instanceName, instanceReq.Payload)
				}
			}
		}
	}

	// Configure hosts in parallel
	errChan := make(chan error, len(instances))
	var wg sync.WaitGroup

	for _, instance := range instances {
		wg.Add(1)
		go func(inst InstanceInfo) {
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
func (i *Infrastructure) provisionInstance(instance InstanceInfo) error {
	fmt.Printf("🔧 Starting provisioning for %s (%s)...\n", instance.Name, instance.IP)

	// Use the user's SSH key path
	sshKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	if err := i.provisioner.ConfigureHost(instance.IP, sshKeyPath); err != nil {
		return fmt.Errorf("failed to configure host: %w", err)
	}

	fmt.Printf("✅ Provisioning completed for %s\n", instance.Name)
	return nil
}
