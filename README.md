# Talis

Talis is a multi-cloud infrastructure provisioning and configuration project that uses:

- Pulumi (in Go) to create cloud instances on AWS or DigitalOcean
- Ansible to provision and configure software (e.g., install packages, manage services) on those instances

## Overview

- **Multi-cloud**: With a single codebase, you can choose which cloud provider to target—AWS or DigitalOcean—by setting an environment variable
- **Pulumi**: Handles infrastructure creation (VM instances, security groups, etc.)
- **Ansible**: Once the instance is up, Ansible installs and configures applications such as Nginx

## Requirements

- Go (1.20 or higher recommended)
- Pulumi CLI (logged in to Pulumi Cloud or a self-hosted backend)
- Ansible (installed locally, if you plan to run playbooks from your host machine)
- Cloud Credentials:
  - For AWS: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, plus optional region config via `pulumi config set aws:region <REGION>`
  - For DigitalOcean: A Personal Access Token, which you can set with `pulumi config set digitalocean:token <YOUR_TOKEN>` or as an environment variable `DIGITALOCEAN_TOKEN`

## Project Structure

```
talis/
├── Pulumi.yaml             # Pulumi project config
├── go.mod                  # Go module config (Pulumi dependencies)
├── main.go                 # Main entry point for Pulumi (in Go)
├── compute/
│   ├── ansible.go          # AnsibleProvider: creates an inventory and runs the playbook
│   ├── compute.go          # Defines the ComputeProvider interface
│   └── digitalocean.go     # DigitalOceanProvider: creates a Droplet
└── ansible/
    ├── inventory.ini       # List of hosts (updated with the VM's IP)
    └── playbook.yml        # Ansible playbook (install & configure software)
```

## Key Files

### Pulumi.yaml
- Describes the Pulumi project (runtime: Go).

### go.mod
- Specifies Go module dependencies, including pulumi-aws, pulumi-digitalocean, and pulumi/sdk.

### main.go
- Reads an environment variable `PROVIDER` to decide between AWS or DigitalOcean.
- Uses the `compute.NewComputeProvider(providerName)` function to instantiate the chosen provider.
- Defines a `userData` script to install Python or other prerequisites on the instance.
- Exports the VM's public IP as `public_ip`.

### compute/
- **compute.go**: Contains the `ComputeProvider` interface with `CreateInstance(ctx, name, userData) (InstanceInfo, error)`.
- **aws.go**: Implements `ComputeProvider` for AWS.
- **digitalocean.go**: Implements `ComputeProvider` for DigitalOcean.

### ansible/
- **inventory.ini**: Basic Ansible inventory. You'll update this with the IP from Pulumi's output.
- **playbook.yml**: Example tasks (e.g., installing Nginx, ensuring services are running).

---

## Setup and Deployment

Clone this repository:
```bash
git clone https://github.com/yourorg/talis.git
cd talis
```

Install Pulumi & Dependencies:

-  Install Pulumi
-  Install Go
-  Run:
```bash
go mod tidy
```


```bash
export DIGITALOCEAN_TOKEN=<YOUR_DO_TOKEN>;
export ACTION=create;
export SSH_KEY_ID=<YOUR_SSH_KEY_ID>; # The name of the SSH key in DigitalOcean
go run main.go
```

To create multiple instances:

```bash
export DIGITALOCEAN_TOKEN=<YOUR_DO_TOKEN>;
export ACTION=create;
export SSH_KEY_ID=<YOUR_SSH_KEY_ID>; # The name of the SSH key in DigitalOcean
export INSTANCE_COUNT=25; # Specify the number of instances to create
go run main.go
```

Choose Provider:

By default, you can set:
```bash
export PROVIDER=digitalocean
```
or
```bash
export PROVIDER=aws
```

Deploy:
```bash
go run main.go
```

Pulumi will:

- Create a VM (Ubuntu-based) on your chosen cloud.
- Run the userData script to install Python (and optionally pip/Ansible).
- Output the VM’s public IP as public_ip.¡


body example to create an instance:
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
            "user_data": "apt-get update -y && apt-get install -y python3-pip && pip3 install ansible",
            "tags": ["talis-do-instance"],
            "ssh_key_name": "<your-ssh-key-name>"
        }
    ]
}
```

---

## Ansible Usage

Update the Ansible Inventory:

After running pulumi up, note the IP Pulumi prints out (e.g. 1.2.3.4).

Edit ansible/inventory.ini:

```ini
[myserver]
myserver ansible_host=1.2.3.4 ansible_user=root
```

For AWS Ubuntu, you’d typically use ansible_user=ubuntu.

Run the Playbook:

```bash
ansible-playbook ansible/playbook.yml -i ansible/inventory.ini \
    --private-key=~/.ssh/id_rsa
```

This will install and configure the specified packages/services, such as Nginx.

---

## Cleanup

When you’re done and want to avoid ongoing cloud costs, you can run:

```bash
export ACTION=delete;
go run main.go
```

This removes all created resources from your cloud account.

---

## Extending talis

- Add More Providers: Create additional files for new providers (e.g., GCP, Azure) implementing the same ComputeProvider interface.
- Enhance Ansible: Increase complexity in playbook.yml with roles, variables, templates, etc.
- Security & Networking: Update firewall or Security Group rules in AWS or DigitalOcean.
- Automation: Integrate with CI/CD (GitHub Actions, GitLab CI, etc.) for automated deploys.
- Webhook: To keep the client updated about the status of the infrastructure, we can use a webhook.

---
