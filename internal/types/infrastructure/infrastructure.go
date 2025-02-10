package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/tty47/talis/internal/compute"
)

const (
	// ansibleTimeout is the timeout for the Ansible playbook
	ansibleTimeout = 30
)

// InstanceRequest represents the API request structure
type InstanceRequest struct {
	Name        string     `json:"name"`
	ProjectName string     `json:"project_name"`
	Action      string     `json:"action"`
	Instances   []Instance `json:"instances"`
}

// Instance represents a single instance configuration
type Instance struct {
	Provider          string   `json:"provider"`
	NumberOfInstances int      `json:"number_of_instances"`
	Region            string   `json:"region"`
	Size              string   `json:"size"`
	Image             string   `json:"image"`
	UserData          string   `json:"user_data"`
	Tags              []string `json:"tags"`
	SSHKeyName        string   `json:"ssh_key_name"`
}

// Infrastructure represents the infrastructure management
type Infrastructure struct {
	request  InstanceRequest
	instance Instance
	stack    *auto.Stack
	provider compute.ComputeProvider
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
		request:  req,
		instance: instance,
		provider: provider,
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

// setupInfrastructure initializes the cloud provider and Pulumi stack
func (i *Infrastructure) setupInfrastructure() error {
	fmt.Printf("Starting Pulumi with stack: %s\n", i.request.ProjectName)
	fmt.Printf("Creating %d instance(s)...\n", i.instance.NumberOfInstances)
	fmt.Printf("Using SSH key: %s\n", i.instance.SSHKeyName)

	stack, err := i.createPulumiStack()
	if err != nil {
		return fmt.Errorf("failed to create stack: %v", err)
	}

	// Configure the provider before using it
	if err := i.provider.ConfigureProvider(*stack); err != nil {
		return fmt.Errorf("failed to configure provider: %v", err)
	}

	i.stack = stack
	return nil
}

// createPulumiStack creates a new Pulumi stack
func (i *Infrastructure) createPulumiStack() (*auto.Stack, error) {
	stack, err := auto.UpsertStackInlineSource(context.Background(), i.request.ProjectName, "talis", func(ctx *pulumi.Context) error {
		return i.createInstances(ctx)
	})
	if err != nil {
		return nil, err
	}
	return &stack, nil
}

// createInstances creates multiple instances in parallel
func (i *Infrastructure) createInstances(ctx *pulumi.Context) error {
	errChan := make(chan error, i.instance.NumberOfInstances)
	var wg sync.WaitGroup

	for idx := 0; idx < i.instance.NumberOfInstances; idx++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if err := i.createSingleInstance(ctx, index); err != nil {
				errChan <- err
			}
		}(idx)
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

// createSingleInstance creates a single instance
func (i *Infrastructure) createSingleInstance(ctx *pulumi.Context, index int) error {
	instanceName := fmt.Sprintf("%s-%s-%d", i.request.Name, time.Now().Format("20060102-150405"), index+1)
	fmt.Printf("Creating instance %d/%d: %s\n", index+1, i.instance.NumberOfInstances, instanceName)

	config := compute.InstanceConfig{
		Region:   i.instance.Region,
		Size:     i.instance.Size,
		Image:    i.instance.Image,
		UserData: i.instance.UserData,
		SSHKeyID: i.instance.SSHKeyName,
		Tags:     i.instance.Tags,
	}

	instanceInfo, err := i.provider.CreateInstance(
		ctx,
		instanceName,
		config,
	)
	if err != nil {
		return fmt.Errorf("failed to create instance %s: %v", instanceName, err)
	}

	ctx.Export(fmt.Sprintf("%s_public_ip", instanceName), instanceInfo.PublicIP)
	return nil
}

// handleCreate manages the creation of infrastructure and runs Ansible playbooks
func (i *Infrastructure) handleCreate() (interface{}, error) {
	fmt.Println("🚀 Creating infrastructure...")
	up, err := i.stack.Up(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %v", err)
	}

	fmt.Println("✅ Infrastructure created successfully!")

	// Collect instances and run Ansible
	instances := make(map[string]string)
	for outputName, output := range up.Outputs {
		if ip, ok := output.Value.(string); ok {
			instanceName := outputName[:len(outputName)-len("_public_ip")]
			instances[instanceName] = ip
			fmt.Printf("📍 Instance %s - IP: %s\n", instanceName, ip)
			fmt.Printf("🔑 SSH Command: ssh root@%s\n", ip)
		}
	}

	// Run Ansible for the created instances
	if err := i.runAnsibleProvisioning(instances); err != nil {
		return instances, fmt.Errorf("infrastructure created but ansible provisioning failed: %v", err)
	}

	return instances, nil
}

// runAnsibleProvisioning runs the Ansible playbook for all instances
func (i *Infrastructure) runAnsibleProvisioning(instances map[string]string) error {
	fmt.Printf("⏳ Waiting for instances to be ready (%ds)...\n", ansibleTimeout)
	time.Sleep(ansibleTimeout * time.Second)

	// Get base name from first instance
	var baseName string
	for name := range instances {
		baseName = strings.Split(name, "-")[0]
		break
	}

	// Create inventory with all instances
	sshKeyPath := getEnvOrDefault("SSH_KEY_PATH", "~/.ssh/id_rsa")
	if err := compute.CreateInventory(instances, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create inventory: %v", err)
	}

	// Run Ansible playbook once for all instances
	if err := compute.RunAnsiblePlaybook(baseName); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %v", err)
	}

	fmt.Println("✅ Ansible provisioning completed successfully!")
	return nil
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// handleDelete manages the deletion of infrastructure
func (i *Infrastructure) handleDelete() (interface{}, error) {
	fmt.Println("🗑️  Destroying infrastructure...")
	_, err := i.stack.Destroy(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to destroy infrastructure: %v", err)
	}
	fmt.Println("✅ Infrastructure destroyed successfully!")
	return map[string]string{"status": "deleted"}, nil
}

// Validate validates the infrastructure request
func (r *InstanceRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if r.Action == "" {
		return fmt.Errorf("action is required")
	}
	if len(r.Instances) == 0 {
		return fmt.Errorf("at least one instance configuration is required")
	}

	for i, instance := range r.Instances {
		if err := instance.Validate(); err != nil {
			return fmt.Errorf("invalid instance configuration at index %d: %v", i, err)
		}
	}

	return nil
}

// Validate validates the instance configuration
func (i *Instance) Validate() error {
	if i.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if i.NumberOfInstances < 1 {
		return fmt.Errorf("number_of_instances must be greater than 0")
	}
	if i.Region == "" {
		return fmt.Errorf("region is required")
	}
	if i.Size == "" {
		return fmt.Errorf("size is required")
	}
	if i.Image == "" {
		return fmt.Errorf("image is required")
	}
	if i.SSHKeyName == "" {
		return fmt.Errorf("ssh_key_name is required")
	}
	return nil
}
