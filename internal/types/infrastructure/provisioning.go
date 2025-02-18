package infrastructure

import (
	"fmt"
	"sync"
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
	errChan := make(chan error, len(instances))
	var wg sync.WaitGroup

	for _, instance := range instances {
		wg.Add(1)
		go func(inst InstanceInfo) {
			defer wg.Done()
			if err := i.provisionInstance(inst); err != nil {
				errChan <- fmt.Errorf("failed to provision %s: %v", inst.Name, err)
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

	return nil
}

// provisionInstance configures a single instance with Ansible
func (i *Infrastructure) provisionInstance(instance InstanceInfo) error {
	fmt.Printf("🔧 Starting provisioning for %s (%s)...\n", instance.Name, instance.IP)

	if err := i.provisioner.ConfigureHost(instance.IP, "~/.ssh/id_rsa"); err != nil {
		return fmt.Errorf("failed to configure host: %v", err)
	}

	fmt.Printf("✅ Provisioning completed for %s\n", instance.Name)
	return nil
}
