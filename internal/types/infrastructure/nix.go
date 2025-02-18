package infrastructure

import (
	"fmt"
	"sync"
)

// RunNixProvisioning applies Nix configuration to all instances
func (i *Infrastructure) RunNixProvisioning(instances []InstanceInfo) error {
	// Check if any instance requires provisioning
	needsProvisioning := false
	for _, inst := range i.instances {
		if inst.Provision {
			needsProvisioning = true
			break
		}
	}

	if !needsProvisioning {
		fmt.Println("⏭️ Skipping Nix provisioning as requested")
		return nil
	}

	fmt.Println("⚙️ Starting Nix provisioning...")
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

// getNixOSConfig returns the NixOS configuration for a specific instance
func (i *Infrastructure) getNixOSConfig(instanceName string) string {
	return fmt.Sprintf(`
{ config, pkgs, ... }:
{
	# Provider specific settings
	imports = [ 
		./base.nix
		./%s 
	];
	
	# Custom instance settings
	networking.hostName = "%s";
}
    `, i.provider.GetNixOSConfig(), instanceName)
}

// provisionInstance configures a single instance with Nix
func (i *Infrastructure) provisionInstance(instance InstanceInfo) error {
	fmt.Printf("🔧 Starting Nix provisioning for %s (%s)...\n", instance.Name, instance.IP)

	// Instalar Nix
	if err := i.nixConfig.InstallNix(instance.IP, "~/.ssh/id_rsa"); err != nil {
		return fmt.Errorf("failed to install Nix: %v", err)
	}

	// Preparar NixOS
	if err := i.nixConfig.PrepareNixOS(instance.IP, "~/.ssh/id_rsa"); err != nil {
		return fmt.Errorf("failed to prepare NixOS: %v", err)
	}

	// Aplicar configuración
	config := i.getNixOSConfig(instance.Name)
	if err := i.nixConfig.ApplyConfiguration(instance.IP, "~/.ssh/id_rsa", config); err != nil {
		return fmt.Errorf("failed to apply configuration: %v", err)
	}

	fmt.Printf("✅ Nix provisioning completed for %s\n", instance.Name)
	return nil
}
