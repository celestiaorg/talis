package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// Infrastructure represents the infrastructure management
type Infrastructure struct {
	request     InstanceRequest         // Request containing the infrastructure configuration
	instances   []Instance              // Instance configuration
	stack       *auto.Stack             // Pulumi stack for managing the infrastructure
	provider    compute.ComputeProvider // Compute provider implementation
	provisioner compute.Provisioner
}

// NewInfrastructure creates a new infrastructure instance
func NewInfrastructure(req InstanceRequest, instance Instance) (*Infrastructure, error) {
	if instance.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	provider, err := compute.NewComputeProvider(instance.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %v", err)
	}

	return &Infrastructure{
		request:     req,
		instances:   req.Instances,
		provider:    provider,
		provisioner: compute.NewProvisioner(),
	}, nil
}

// Execute executes the requested action (create/delete)
func (i *Infrastructure) Execute() (interface{}, error) {
	if err := i.setupInfrastructure(); err != nil {
		return nil, fmt.Errorf("failed to setup infrastructure: %v", err)
	}

	switch i.request.Action {
	case "create":
		return i.handleCreate()
	case "delete":
		return i.handleDelete()
	default:
		return nil, fmt.Errorf("invalid action: %s", i.request.Action)
	}
}

// updateJobStatus updates the job status (this es un placeholder - implementar con una DB real)
func (i *Infrastructure) updateJobStatus(jobID, status string, result interface{}) {
	// TODO: Store in database and trigger webhook if configured
	fmt.Printf("Job %s status updated to %s: %v\n", jobID, status, result)
}

// handleCreate handles the creation of infrastructure
func (i *Infrastructure) handleCreate() (interface{}, error) {
	fmt.Println("🚀 Creating infrastructure...")

	// Create and configure the stack
	_, err := i.stack.Up(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %v", err)
	}

	// Get the created instances information
	instances, err := i.GetOutputs()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance information: %v", err)
	}

	// Give instances a moment to fully initialize
	fmt.Println("⏳ Waiting for instances to be fully ready...")
	time.Sleep(30 * time.Second)

	// Run Nix provisioning if enabled
	if err := i.RunProvisioning(instances); err != nil {
		return nil, fmt.Errorf("failed to provision instances: %v", err)
	}

	return instances, nil
}

// handleDelete handles the deletion of infrastructure
func (i *Infrastructure) handleDelete() (interface{}, error) {
	fmt.Println("🗑️ Deleting infrastructure...")

	_, err := i.stack.Destroy(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to delete infrastructure: %v", err)
	}

	return map[string]string{"status": "deleted"}, nil
}

// Update updates an existing infrastructure
func (i *Infrastructure) Update() (interface{}, error) {
	// Get Pulumi token
	pulumiToken := os.Getenv("PULUMI_ACCESS_TOKEN")
	if pulumiToken == "" {
		return nil, fmt.Errorf("PULUMI_ACCESS_TOKEN environment variable is not set")
	}

	// Create a new workspace with project settings
	workspaceOpts := []auto.LocalWorkspaceOption{
		auto.WorkDir("."),
		auto.PulumiHome(".pulumi"),
		auto.EnvVars(map[string]string{
			"PULUMI_ACCESS_TOKEN": pulumiToken,
		}),
	}

	ws, err := auto.NewLocalWorkspace(context.Background(), workspaceOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %v", err)
	}

	// List all stacks to debug
	stacks, err := ws.ListStacks(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list stacks: %v", err)
	}

	fmt.Println("Available stacks:")
	for _, s := range stacks {
		fmt.Printf("- %s\n", s.Name)
	}

	// Use the project name as the stack name since that's how it was created
	stackName := i.request.ProjectName
	fmt.Printf("Attempting to select stack: %s\n", stackName)

	// Get existing stack
	stack, err := auto.SelectStack(context.Background(), stackName, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to select stack %s: %v", stackName, err)
	}

	// Get instance IP from stack outputs
	outputs, err := stack.Outputs(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get stack outputs: %v", err)
	}

	instances := make([]InstanceInfo, 0)
	for name, output := range outputs {
		if strings.HasSuffix(name, "_ip") {
			instanceName := strings.TrimSuffix(name, "_ip")
			if ip, ok := output.Value.(string); ok {
				instances = append(instances, InstanceInfo{
					Name: instanceName,
					IP:   ip,
				})
				fmt.Printf("Found instance: %s with IP: %s\n", instanceName, ip)
			}
		}
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instance IPs found in stack outputs")
	}

	// Apply new Nix configuration
	if err := i.RunProvisioning(instances); err != nil {
		return nil, fmt.Errorf("failed to update configuration: %v", err)
	}

	return instances, nil
}

// GetOutputs gets the outputs from the Pulumi stack
func (i *Infrastructure) GetOutputs() ([]InstanceInfo, error) {
	if i.stack == nil {
		return nil, fmt.Errorf("stack not initialized")
	}

	outputs, err := i.stack.Outputs(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get stack outputs: %v", err)
	}

	var instances []InstanceInfo
	for name, output := range outputs {
		if strings.HasSuffix(name, "_ip") {
			instanceName := strings.TrimSuffix(name, "_ip")
			if ip, ok := output.Value.(string); ok {
				instances = append(instances, InstanceInfo{
					Name: instanceName,
					IP:   ip,
				})
				fmt.Printf("📍 Instance %s - IP: %s\n", name, ip)
			}
		}
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instance IPs found in stack outputs")
	}

	return instances, nil
}
