# Talis

Talis is a multi-cloud infrastructure provisioning and configuration project that uses:

- Pulumi (in Go) to create cloud instances on AWS or DigitalOcean
- NixOS for system configuration and package management

## Overview

- **Multi-cloud**: With a single codebase, you can choose which cloud provider to use—AWS or DigitalOcean
- **Pulumi**: Handles infrastructure creation (VM instances, security groups, etc.)
- **NixOS**: Provides declarative system configuration, ensuring reproducible environments

## Requirements

- Go (1.20 or higher)
- Pulumi CLI
- SSH key pair for instance access
- Cloud Credentials:
  - For DigitalOcean: Personal Access Token in `DIGITALOCEAN_TOKEN` environment variable
  - For AWS: Coming soon

## Project Structure

```
talis/
├── cmd/
│   └── main.go                    # Main entry point
├── internal/
│   ├── compute/                   # Cloud provider implementations
│   │   ├── compute.go            # ComputeProvider interface and common types
│   │   └── digitalocean.go       # DigitalOcean implementation
│   └── types/
│       └── infrastructure/        # Infrastructure types and logic
│           ├── models.go         # Type definitions
│           ├── validation.go     # Request validation
│           ├── pulumi.go        # Pulumi logic
│           ├── nix.go           # NixOS configuration
│           └── infrastructure.go # Main infrastructure logic
├── nix/                          # NixOS configurations
│   ├── base.nix                  # Base system configuration
│   └── cloud/                    # Cloud-specific configurations
│       └── digitalocean.nix      # DigitalOcean configuration
└── .env.example                  # Environment variables example
```

## Key Files

### internal/compute/
- **compute.go**: Defines the `ComputeProvider` interface and common types:
  ```go
  type ComputeProvider interface {
      ConfigureProvider(stack auto.Stack) error
      CreateInstance(ctx *pulumi.Context, name string, config InstanceConfig) (pulumi.Resource, error)
      GetNixOSConfig() string
      ValidateCredentials() error
      GetEnvironmentVars() map[string]string
  }
  ```
- **digitalocean.go**: `ComputeProvider` implementation for DigitalOcean

### internal/types/infrastructure/
- **models.go**: Main data structure definitions
- **validation.go**: Request validation
- **pulumi.go**: Pulumi stack management
- **nix.go**: NixOS configuration and provisioning
- **infrastructure.go**: Main infrastructure management logic

## Setup

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Configure environment variables:
```bash
# Pulumi
PULUMI_ACCESS_TOKEN=your_pulumi_token_here

# DigitalOcean
DIGITALOCEAN_TOKEN=your_digitalocean_token_here
SSH_KEY_ID=your_key_id_here
```

## Usage

### Create Instances

```json
{
    "name": "talis",
    "project_name": "talis-pulumi-ansible",
    "action": "create",
    "instances": [
        {
            "provider": "digitalocean",
            "number_of_instances": 1,
            "region": "nyc3",
            "size": "s-1vcpu-1gb",
            "image": "ubuntu-22-04-x64",
            "tags": ["talis-do-instance"],
            "ssh_key_name": "your-ssh-key-name",
            "provision": true
        }
    ]
}
```

### Delete Instances

```json
{
    "name": "talis",
    "project_name": "talis-pulumi-ansible",
    "action": "delete",
    "instances": [
        {
            "provider": "digitalocean",
            "number_of_instances": 1,
            "region": "nyc3",
            "size": "s-1vcpu-1gb"
        }
    ]
}
```

## Extensibility

### Adding New Providers

1. Create new file in `internal/compute/` (e.g., `aws.go`)
2. Implement the `ComputeProvider` interface
3. Add the provider in `NewComputeProvider` in `compute.go`
4. Create NixOS configuration in `nix/cloud/`

### Customizing NixOS

Modify files in `nix/`:
- `base.nix`: Common configuration
- `cloud/provider.nix`: Provider-specific configuration

## Upcoming Features

- AWS support
- More NixOS configuration options
- Error handling improvements
- Webhook notification system
- CI/CD integration

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

---
