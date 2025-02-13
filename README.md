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
│   ├── main.go                    # Main entry point
│   └── migrate/                   # Database migration tools
│       └── main.go               # Migration entry point
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
│   │   └── nix.go                # NixOS configuration and provisioning
│   ├── db/                        # Database layer
│   │   ├── migrations/          # Database migrations
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
├── nix/                          # NixOS configurations
│   ├── base.nix                  # Base system configuration
│   └── cloud/                    # Cloud-specific configurations
│       └── digitalocean.nix      # DigitalOcean configuration
├── migrations/                    # SQL migration files
│   └── *.sql                     # Migration SQL files
├── scripts/                      # Utility scripts
└── .env.example                  # Environment variables example
├── Makefile                      # Build and development commands
└── Pulumi.*.yaml                 # Pulumi stack configurations
```

## Key Files

### cmd/
- **main.go**: Application entry point with server setup
- **migrate/main.go**: Database migration tool

### internal/api/v1/
- **handlers/**: HTTP request handlers
- **middleware/**: API middleware (logging, auth, etc.)
- **routes/**: API route definitions

### internal/application/
- **job/**: Job service implementation with business logic

### internal/compute/
- **compute.go**: Defines the `ComputeProvider` interface and common types
- **digitalocean.go**: `ComputeProvider` implementation for DigitalOcean
- **nix.go**: NixOS installation and configuration management

### internal/db/
- **migrations/**: Database migration management
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

### migrations/
- SQL files for database schema and updates

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
