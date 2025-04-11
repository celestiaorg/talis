# Talis - Infrastructure Management System

Talis is a multi-cloud infrastructure provisioning and configuration project that uses:

- Direct Cloud Provider APIs to create and manage cloud instances
- Ansible for initial system configuration and package installation

---

## Overview

- **Multi-cloud**: With a single codebase, you can choose which cloud provider to use (currently supporting DigitalOcean, with more providers coming soon)
- **Direct API Integration**: Uses cloud provider APIs directly for better control and reliability
- **Ansible**: Provides initial system configuration and package installation
- **Extensive Testing**: Comprehensive test coverage for all cloud provider operations

---

## Requirements

- Go (1.24 or higher)
- Ansible (2.9 or higher)
- SSH key pair for instance access
- Cloud Credentials:
  - For [DigitalOcean](https://www.digitalocean.com/): Personal Access Token in `DIGITALOCEAN_TOKEN` environment variable
  - For [Linode](https://www.linode.com/): Coming soon
  - For [Vultr](https://www.vultr.com/): Coming soon
  - For [DataPacket](https://www.datapacket.com/): Coming soon

---

## Project Structure

```
talis/
├── cmd/
│   └── main.go                    # Main entry point
├── internal/
│   ├── api/                       # API related code
│   │   └── v1/
│   │       ├── handlers/         # Request handlers (instances, jobs)
│   │       ├── middleware/       # API middleware
│   │       ├── routes/          # Route definitions
│   │       └── services/        # Business logic services
│   ├── compute/                   # Cloud provider implementations
│   │   ├── provider.go           # ComputeProvider interface and common types
│   │   └── digitalocean.go       # DigitalOcean implementation
│   ├── provisioner/              # Provisioning system
│   │   ├── ansible.go           # Ansible implementation
│   │   ├── interface.go         # Provisioner interfaces
│   │   ├── factory.go          # Provisioner factory
│   │   └── config/             # Configuration types
│   │       └── ansible.go      # Ansible-specific config
│   ├── events/                   # Event system
│   │   ├── events.go           # Event definitions and bus
│   ├── db/                       # Database layer
│   │   ├── db.go               # Database connection and configuration
│   │   ├── models/             # Database models (instances, jobs)
│   │   └── repos/              # Database repositories
│   └── types/                    # Common types and models
│       └── infrastructure/       # Infrastructure types and logic
├── ansible/                       # Ansible configurations
│   ├── main.yml                  # Main Ansible configuration
│   ├── stages/                   # Task stages for different configurations
│   │   └── setup.yml            # Initial setup and configuration tasks
│   ├── vars/                     # Variable definitions
│   │   └── main.yml             # Main variables file
│   └── inventory/                # Generated inventory files
│       └── inventory_*_ansible.ini
├── scripts/                       # Utility scripts
└── .env.example                   # Environment variables example
```

---

## Key Components

### internal/api/v1/
- **handlers/**: HTTP request handlers for instances and jobs
- **middleware/**: API middleware (logging, auth, etc.)
- **routes/**: API route definitions
- **services/**: Business logic services

### internal/db/
- **models/**: Database models for instances and jobs
- **repos/**: Database repositories with CRUD operations
- **db.go**: Database connection and configuration

### internal/compute/
- **provider.go**: Defines the `ComputeProvider` interface and common types
- **digitalocean.go**: Implementation for DigitalOcean with comprehensive test coverage

### ansible/
- **main.yml**: Main Ansible configuration file
- **stages/**: Contains different stages of configuration
  - **setup.yml**: Initial setup and configuration tasks
- **vars/**: Variable definitions for Ansible
  - **main.yml**: Main variables configuration
- **inventory/inventory_*_ansible.ini**: Generated inventory files for each deployment

---

## Architecture

Talis uses an event-driven architecture to manage infrastructure creation and provisioning. This design provides several benefits:

### Event-Driven Flow

```
+------------+     +------------+     +------------+
|            |     |            |     |            |
| services   +---->+  events    +---->+provisioner |
|            |     |            |     |            |
+------------+     +------------+     +-----+------+
                                           |
                   +------------+          |
                   |            |          |
                   |  compute   +----------+
                   |            |
                   +------------+
```

1. **Infrastructure Creation**
   - User requests instance creation via API
   - Service creates instances using cloud provider
   - Service emits `instances_created` event

2. **Provisioning**
   - Provisioning service listens for events
   - On `instances_created`, retrieves instances from DB
   - Configures and provisions instances using Ansible
   - Filters out terminated instances automatically

### Key Features

1. **Decoupled Components**
   - Services communicate through events
   - No direct dependencies between services
   - Easy to add new features without modifying existing code

2. **Robust Error Handling**
   - Infrastructure creation and provisioning are separate
   - Failures in one component don't affect others
   - Built-in retry mechanisms

3. **Efficient Instance Management**
   - Automatic filtering of terminated instances
   - Clear separation of instance states
   - Improved resource utilization

---

## Setup

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Configure environment variables:
```bash
# DigitalOcean
DIGITALOCEAN_TOKEN=your_digitalocean_token_here
```

3. Ensure your SSH key is available:
```bash
# The default path is ~/.ssh/id_rsa
# You can specify a different path in the request
```

---

## Usage

### Creating Instances

```bash
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d '{
    "job_name": "my-job",
    "instance_name": "test-instance",
    "instances": [{
      "provider": "do",
      "region": "nyc1",
      "size": "s-1vcpu-1gb",
      "provision": true
    }]
  }'
```

The system will:
1. Create the instance in DigitalOcean
2. Emit an event for provisioning
3. Configure the instance using Ansible
4. Update instance status in database

### Monitoring Status

```bash
curl http://localhost:8080/api/v1/instances/{id}
```

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
- `main.yml`: Main Ansible configuration
- `stages/setup.yml`: Initial system setup and configuration
- `vars/main.yml`: Variable definitions
- Add new stages in `ansible/stages/` for additional configurations

## Upcoming Features

- AWS provider implementation
- Linode provider implementation
- Vultr provider implementation
- DataPacket provider implementation
- Webhook notification system
- Enhanced job management and monitoring
- 100 Light Nodes deployment support

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Ansible 2.9+

### Setup

1. Clone the repository
```bash
git clone https://github.com/celestiaorg/talis.git
```

2. Copy environment file
```bash
cp .env.example .env
```

3. Run the service
```bash
go run cmd/main.go
```

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

