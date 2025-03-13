# Talis 🦍

Talis is a multi-cloud infrastructure provisioning and configuration project that uses:

- Direct Cloud Provider APIs to create and manage cloud instances
- Ansible for initial system configuration and package installation

## Overview

- **Multi-cloud**: With a single codebase, you can choose which cloud provider to use (currently supporting DigitalOcean, with more providers coming soon)
- **Direct API Integration**: Uses cloud provider APIs directly for better control and reliability
- **Ansible**: Provides initial system configuration and package installation
- **Extensive Testing**: Comprehensive test coverage for all cloud provider operations

## Requirements

- Go (1.22 or higher)
- Ansible (2.9 or higher)
- SSH key pair for instance access
- Cloud Credentials:
  - For [DigitalOcean](https://www.digitalocean.com/): Personal Access Token in `DIGITALOCEAN_TOKEN` environment variable
  - For [Linode](https://www.linode.com/): Coming soon
  - For [Vultr](https://www.vultr.com/): Coming soon
  - For [DataPacket](https://www.datapacket.com/): Coming soon

## Project Structure

```
talis/
├── cmd/
│   └── main.go                    # Main entry point
├── internal/
│   ├── api/                       # API related code
│   │   └── v1/
│   │       ├── handlers/         # Request handlers
│   │       ├── middleware/       # API middleware
│   │       └── routes/           # Route definitions
│   ├── compute/                   # Cloud provider implementations
│   │   ├── provider.go           # ComputeProvider interface and common types
│   │   ├── digitalocean.go       # DigitalOcean implementation
│   │   └── ansible.go            # Ansible configuration and provisioning
│   ├── db/                        # Database layer
│   │   └── job/                  # Job database models
│   └── types/                     # Common types and models
│       └── infrastructure/        # Infrastructure types and logic
├── ansible/                       # Ansible configurations
│   ├── playbook.yml              # Main Ansible playbook
│   └── inventory_*_ansible.ini   # Generated inventory files
├── scripts/                       # Utility scripts
└── .env.example                   # Environment variables example
```

## Key Components

### internal/compute/
- **provider.go**: Defines the `ComputeProvider` interface and common types
- **digitalocean.go**: Implementation for DigitalOcean with comprehensive test coverage
- **ansible.go**: Ansible configuration and provisioning

### internal/api/v1/
- **handlers/**: HTTP request handlers
- **middleware/**: API middleware (logging, auth, etc.)
- **routes/**: API route definitions

### ansible/
- **playbook.yml**: Main Ansible playbook
- **inventory_*_ansible.ini**: Generated inventory files

## Setup

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Configure environment variables:
```bash
# DigitalOcean
DIGITALOCEAN_TOKEN=your_digitalocean_token_here
SSH_KEY_NAME=your_ssh_key_name_here
```

3. Ensure your SSH key is available:
```bash
# The default path is ~/.ssh/id_rsa
# You can specify a different path in the request
```

## Usage

### Using the CLI

```bash
# Build the CLI
make build-cli

# Create infrastructure using a JSON file
talis infra create -f create.json

# Delete infrastructure using a JSON file
talis infra delete -f delete.json

# List all jobs
talis jobs list

# Get job status
talis jobs get --id job-20240315-123456
```

### Create Instances

Example configuration (create.json):
```json
{
    "instance_name": "talis",
    "project_name": "talis-test",
    "instances": [
        {
            "provider": "digitalocean",
            "number_of_instances": 1,
            "provision": true,
            "region": "nyc3",
            "size": "s-1vcpu-1gb",
            "image": "ubuntu-22-04-x64",
            "tags": ["talis-do-instance"],
            "ssh_key_name": "your-ssh-key-name"
        },
        {
            "provider": "digitalocean",
            "name": "talis-validator",
            "number_of_instances": 1,
            "provision": true,
            "region": "nyc3",
            "size": "s-2vcpu-2gb",
            "image": "ubuntu-22-04-x64",
            "tags": ["talis-validator"],
            "ssh_key_name": "your-ssh-key-name"
        }
    ]
}
```

The `instance_name` field is used as a base name for instances. Each instance gets a suffix that is incremented starting from 0 (e.g., "talis-0"). Individual instances can have custom names by specifying the `name` field in the instance object, as shown in the example above with "talis-validator".

### Delete Instances

Example configuration (delete.json):
```json
{
    "id": 10,
    "instance_name": "talis",
    "project_name": "talis-test",
    "instances": [
        {
            "provider": "digitalocean",
            "number_of_instances": 1,
            "region": "nyc3"
        },
        {
            "provider": "digitalocean",
            "name": "talis-validator",
            "region": "nyc3"
        }
    ]
}
```

When deleting instances, you can specify which instances to delete by providing the `name` field in the instance object. If no specific names are provided, instances are deleted in FIFO order (oldest first).

## Extensibility

### Adding New Providers

1. Create new file in `internal/compute/` (e.g., `aws.go`)
2. Implement the `ComputeProvider` interface:
   ```go
   type ComputeProvider interface {
       ValidateCredentials() error
       GetEnvironmentVars() map[string]string
       ConfigureProvider(stack interface{}) error
       CreateInstance(ctx context.Context, name string, config InstanceConfig) ([]InstanceInfo, error)
       DeleteInstance(ctx context.Context, name string, region string) error
   }
   ```
3. Add the provider in `NewComputeProvider` in `provider.go`
4. Add comprehensive tests following the pattern in `digitalocean_test.go`

### Customizing Ansible

Modify files in `ansible/`:
- `playbook.yml`: Main Ansible playbook
- Add new roles in `ansible/roles/` for modular configurations

## Upcoming Features

- AWS provider implementation
- Linode provider implementation
- Vultr provider implementation
- DataPacket provider implementation
- Webhook notification system
- Enhanced job management and monitoring
- 100 Light Nodes deployment support

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific tests
go test ./internal/compute -run TestDigitalOceanProvider
```

### Code Quality

The project uses:
- golangci-lint for code quality
- go test for unit and integration testing
- yamllint for YAML file validation

Run the linters:
```bash
make lint
```

---
