![logo](./docs/logo.png)

# Talis

Talis is a multi-cloud infrastructure provisioning and configuration project that uses:

- Pulumi (in Go) to create cloud instances on AWS or DigitalOcean
- Ansible for initial system configuration and package installation

## Overview

- **Multi-cloud**: With a single codebase, you can choose which cloud provider to use—AWS or DigitalOcean
- **Pulumi**: Handles infrastructure creation (VM instances, security groups, etc.)
- **Ansible**: Provides initial system configuration and package installation

## Requirements

- Go (1.20 or higher)
- Pulumi CLI
- Ansible (2.9 or higher)
- SSH key pair for instance access
- Cloud Credentials:
  - For DigitalOcean: Personal Access Token in `DIGITALOCEAN_TOKEN` environment variable
  - For AWS: Coming soon

## Project Structure

```
talis/
├── cmd/
│   ├── main.go                    # Main entry point
├── internal/
│   ├── api/                       # API related code
│   │   └── v1/
│   │       ├── handlers/         # Request handlers
│   │       ├── middleware/       # API middleware
│   │       └── routes/           # Route definitions
│   ├── application/              # Application layer
│   │   └── job/                 # Job service implementation
│   ├── compute/                   # Cloud provider implementations
│   │   ├── compute.go            # ComputeProvider interface and common types
│   │   ├── digitalocean.go       # DigitalOcean implementation
│   │   └── ansible.go            # Ansible configuration and provisioning
│   ├── db/                        # Database layer
│   │   └── job/                 # Job database models
│   ├── domain/                    # Domain layer
│   │   └── job/                 # Job domain models and interfaces
│   ├── infrastructure/           # Infrastructure layer
│   │   └── persistence/         # Data persistence implementations
│   │       └── postgres/        # PostgreSQL implementations
│   └── types/
│       └── infrastructure/        # Infrastructure types and logic
│           ├── models.go         # Type definitions
│           ├── validation.go     # Request validation
│           ├── pulumi.go        # Pulumi logic
│           └── infrastructure.go # Main infrastructure logic
├── ansible/                      # Ansible configurations
│   ├── playbook.yml             # Main Ansible playbook
│   └── inventory_*_ansible.ini  # Generated inventory files
├── scripts/                      # Utility scripts
└── .env.example                  # Environment variables example
├── Makefile                      # Build and development commands
└── Pulumi.*.yaml                 # Pulumi stack configurations
```

## Key Files

### cmd/
- **main.go**: Application entry point with server setup

### internal/api/v1/
- **handlers/**: HTTP request handlers
- **middleware/**: API middleware (logging, auth, etc.)
- **routes/**: API route definitions

### internal/application/
- **job/**: Job service implementation with business logic

### internal/compute/
- **compute.go**: Defines the `ComputeProvider` interface and common types
- **digitalocean.go**: `ComputeProvider` implementation for DigitalOcean
- **ansible.go**: Ansible configuration and provisioning

### internal/db/
- **job/**: Job database models and operations

### internal/domain/
- **job/**: Job domain models, interfaces and business rules

### internal/infrastructure/
- **persistence/postgres/**: PostgreSQL implementations of repositories

### internal/types/infrastructure/
- **models.go**: Main data structure definitions
- **validation.go**: Request validation
- **pulumi.go**: Pulumi stack management
- **infrastructure.go**: Main infrastructure management logic

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
# Pulumi
PULUMI_ACCESS_TOKEN=your_pulumi_token_here

# DigitalOcean
DIGITALOCEAN_TOKEN=your_digitalocean_token_here
SSH_KEY_ID=your_key_id_here
```

3. Ensure your SSH key is available:
```bash
# The default path is /root/.ssh/id_rsa
# You can specify a different path in the request
```

## Usage

### Using the CLI

Talis provides a command-line interface for managing infrastructure and jobs.

```bash
# Build the CLI
make build-cli

# Create infrastructure using a JSON file
talis infra create -f create.json

# Delete infrastructure using a JSON file
talis infra delete -f delete.json

# List all jobs
talis jobs list

# List jobs with filters
talis jobs list --limit 10 --status running

# Get job status
talis jobs get --id job-20240315-123456
```

### API Usage

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
            "provision": true,
            "region": "nyc3",
            "size": "s-1vcpu-1gb",
            "image": "ubuntu-22-04-x64",
            "tags": ["talis-do-instance"],
            "ssh_key_name": "your-ssh-key-name"
        }
    ]
}
```

### CLI Commands

#### Infrastructure Management
```bash
# Create infrastructure
talis infra create -f config.json

# Delete infrastructure
talis infra delete -f config.json
```

#### Job Management
```bash
# List all jobs
talis jobs list

# List with filters
talis jobs list --limit 10 --status running

# Get specific job
talis jobs get --id <job-id>
```

### Ansible Provisioning

When `provision: true` is set in the instance configuration, Talis will:

1. Wait for the instance to be accessible via SSH
2. Create an Ansible inventory file
3. Run the Ansible playbook that:
   - Updates system packages
   - Installs required software (nginx, docker, etc.)
   - Configures basic services
   - Sets up firewall rules

The Ansible playbook can be customized by modifying `ansible/playbook.yml`.

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

### Customizing Ansible

Modify files in `ansible/`:
- `playbook.yml`: Main Ansible playbook
- Add new roles in `ansible/roles/` for modular configurations

## Upcoming Features

- More Ansible playbook options
- AWS support
- Webhook notification system
- 100 Light Nodes deployment

---
