# Talis 🦍

Talis is a multi-cloud infrastructure provisioning and configuration project that uses:

- Hypervisor (VitFusion) API to create and manage cloud instances
- Ansible for initial system configuration and package installation

## Overview

- **Multi-cloud**: With a single codebase, you can choose which cloud provider to use (currently supporting DigitalOcean, with more providers coming soon) (This is done through the hypervisor API)
- **Ansible**: Provides initial system configuration and package installation
- **Extensive Testing**: Comprehensive test coverage for all cloud provider operations

## Requirements

- Go (1.24 or higher)
- Ansible (2.9 or higher)
- SSH key pair for instance access
- Cloud Credentials: (To Be Updated for Hypervisor API)
  - For [DigitalOcean](https://www.digitalocean.com/): Personal Access Token in `DIGITALOCEAN_TOKEN` environment variable
  - For [Linode](https://www.linode.com/): Coming soon
  - For [Vultr](https://www.vultr.com/): Coming soon
  - For [DataPacket](https://www.datapacket.com/): Coming soon

## Project Structure

See [architecture doc](./docs/architecture.md)

## Setup

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Add your DigitalOcean Personal Access Token to the `.env` file:
```bash
echo "DIGITALOCEAN_TOKEN=your_digitalocean_token_here" >> .env
```
   Alternatively, you can set it as an environment variable directly.

3. Configure the SSH key for Talis server operations (one of these options):
   - Provide the SSH key content directly (preferred for deployment):
     ```bash
     # Add to .env file
     echo "TALIS_SSH_KEY=-----BEGIN OPENSSH PRIVATE KEY-----..." >> .env
     # Or set as environment variable
     export TALIS_SSH_KEY="-----BEGIN OPENSSH PRIVATE KEY-----..."
     ```

## Usage

### Using the CLI

If this is your first time using talis you will need to initialize a user and a project.

```bash
# Create a user
make run-cli ARGS="users create --username my-user"

# Create a project
make run-cli ARGS="projects create --name my-project --owner-id <user-id>"
```

Now you can create a configuration file and use it to create infrastructure.

```bash
# Copy and modify the example create configuration
cp create.json_example create.json

# Create infrastructure using your configuration
make run-cli ARGS="infra create --file create.json"
# A delete file (e.g., delete_create.json if your input file was create.json) will be automatically generated after successful creation

# Delete infrastructure using the auto-generated file
make run-cli ARGS="infra delete --file delete.json"
```

### Example Configuration Files
See the [create.json_example](./create.json_example) and [delete.json_example](./delete.json_example) files for more information.

### Using the Go API Client

For programmatic access to the Talis API using Go, refer to the [Go API Client Usage](./client_usage.md) documentation.

## Extensibility

### Adding New Providers

TODO: add documentation for hypervisor API

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

### Running Tests

```bash
# Run all tests
make test
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

