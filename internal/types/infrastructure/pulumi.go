package infrastructure

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// setupInfrastructure initializes the cloud provider and Pulumi stack
func (i *Infrastructure) setupInfrastructure() error {
	// Verify Pulumi token
	pulumiToken := os.Getenv("PULUMI_ACCESS_TOKEN")
	if pulumiToken == "" {
		return fmt.Errorf("PULUMI_ACCESS_TOKEN environment variable is not set")
	}

	// Validate provider credentials
	if err := i.provider.ValidateCredentials(); err != nil {
		return err
	}

	fmt.Printf("Starting Pulumi with stack: %s\n", i.projectName)
	totalInstances := 0
	for _, inst := range i.instances {
		totalInstances += inst.NumberOfInstances
	}
	fmt.Printf("Creating %d instance(s)...\n", totalInstances)
	fmt.Printf("Using SSH key: %s\n", i.instances[0].SSHKeyName)

	stack, err := i.createPulumiStack(pulumiToken)
	if err != nil {
		return fmt.Errorf("failed to create stack: %v", err)
	}

	if err := i.provider.ConfigureProvider(*stack); err != nil {
		return fmt.Errorf("failed to configure provider: %v", err)
	}

	i.stack = stack
	return nil
}

// createPulumiStack creates a new Pulumi stack
func (i *Infrastructure) createPulumiStack(pulumiToken string) (*auto.Stack, error) {
	fmt.Println("🔧 Creating Pulumi stack...")

	envVars := map[string]string{
		"PULUMI_ACCESS_TOKEN": pulumiToken,
	}
	// Add provider-specific environment variables
	for k, v := range i.provider.GetEnvironmentVars() {
		envVars[k] = v
	}

	workspaceOpts := []auto.LocalWorkspaceOption{
		auto.WorkDir("."),
		auto.PulumiHome(".pulumi"),
		auto.EnvVars(envVars),
	}

	fmt.Println("📁 Creating Pulumi workspace...")
	stack, err := auto.UpsertStackInlineSource(context.Background(),
		i.projectName,
		"talis",
		func(ctx *pulumi.Context) error {
			fmt.Println("⚡ Running Pulumi program...")
			return i.createInstances(ctx)
		},
		workspaceOpts...,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create/select stack: %v", err)
	}

	fmt.Println("✅ Pulumi stack created successfully")
	return &stack, nil
}

// createInstances creates multiple instances in parallel
func (i *Infrastructure) createInstances(ctx *pulumi.Context) error {
	totalInstances := 0
	for _, inst := range i.instances {
		totalInstances += inst.NumberOfInstances
	}
	fmt.Printf("🚀 Creating %d instances...\n", totalInstances)
	errChan := make(chan error, totalInstances)
	var wg sync.WaitGroup

	instanceIndex := 0
	for _, inst := range i.instances {
		for idx := 0; idx < inst.NumberOfInstances; idx++ {
			wg.Add(1)
			go func(index int, instance Instance) {
				defer wg.Done()
				name := fmt.Sprintf("%s-%d", i.name, index)
				_, err := i.provider.CreateInstance(ctx, name, compute.InstanceConfig{
					Region:   instance.Region,
					Size:     instance.Size,
					Image:    instance.Image,
					SSHKeyID: instance.SSHKeyName,
					Tags:     instance.Tags,
				})
				if err != nil {
					errChan <- fmt.Errorf("failed to create instance %s: %v", name, err)
				}
			}(instanceIndex, inst)
			instanceIndex++
		}
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
