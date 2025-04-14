# SSH Private Key Customization Proposal

This document outlines the proposed changes to allow customization of the SSH private key used by Ansible for connecting to newly provisioned instances.

## Problem

Currently, the Ansible inventory generation process hardcodes the path to the SSH private key as `~/.ssh/id_rsa`. This causes failures when users utilize different key types (e.g., ED25519) or store their keys in non-default locations, as Ansible cannot find the correct private key to establish an SSH connection.

The `SSHKeyName` parameter provided during instance creation relates to the *public key* registered in the user's DigitalOcean account, which is added to the droplet's `authorized_keys` file. This is distinct from the *private key* file needed by Ansible on the control machine.

## Requirements

1.  Users must provide an `SSHKeyName` corresponding to a key in their DigitalOcean account (existing requirement).
2.  By default, Ansible should assume the private key is an RSA key located at `$HOME/.ssh/id_rsa`.
3.  Users should be able to specify an `SSHKeyType` (e.g., "ed25519") to indicate the key type. If specified, Ansible should assume the key is at the default location for that type (e.g., `$HOME/.ssh/id_ed25519`).
4.  Users should be able to provide a custom `SSHKeyPath` to specify the exact location of the private key file.
5.  The priority for determining the private key path for Ansible should be:
    1.  Custom `SSHKeyPath` (if provided).
    2.  Default path based on `SSHKeyType` (if provided and `SSHKeyPath` is not).
    3.  Default RSA path (`$HOME/.ssh/id_rsa`) (if neither `SSHKeyPath` nor `SSHKeyType` is provided).

## Proposed Solution

### 1. Modify Data Structures (`internal/types/instance.go`)

Add two new optional fields to the `InstanceRequest` struct (or the relevant struct defining a single instance within `InstancesRequest`):

```go
// InstanceRequest defines the parameters for creating an instance.
type InstanceRequest struct {
    // ... existing fields ...
    Region        string `json:"region"`
    Size          string `json:"size"`
    Image         string `json:"image"`
    SSHKeyName    string `json:"ssh_key_name"` // Required for DO public key association
    Count         int    `json:"count"`
    Tags          []string `json:"tags"`

    // --- New Fields ---
    // SSHKeyType specifies the type of the private SSH key to use for Ansible connection (e.g., "rsa", "ed25519"). Defaults to "rsa".
    SSHKeyType    string `json:"ssh_key_type,omitempty"`
    // SSHKeyPath specifies the custom path to the private SSH key file for Ansible. Overrides default paths based on SSHKeyType.
    SSHKeyPath    string `json:"ssh_key_path,omitempty"`
    // ... potentially other fields ...
}
```

### 2. Update Validation (Optional - `internal/types/validation.go`)

Add validation logic to ensure `SSHKeyType`, if provided, is a recognized value (e.g., "rsa", "ed25519", "ecdsa").

### 3. Update Ansible Inventory Generation (`internal/compute/ansible.go`)

Modify the function responsible for generating the Ansible inventory file (`.ini`) to dynamically determine the `ansible_ssh_private_key_file` value based on the new fields:

```go
// --- Inside the inventory generation logic, assuming 'req' is the InstanceRequest ---

// Determine the private key path based on priority
privateKeyPath := "$HOME/.ssh/id_rsa" // Default (Requirement 2)

// Priority 1: Custom Path (Requirement 4 & 5)
if req.SSHKeyPath != "" {
    privateKeyPath = req.SSHKeyPath
} else if req.SSHKeyType != "" {
    // Priority 2: Key Type (Requirement 3 & 5)
    switch strings.ToLower(req.SSHKeyType) {
    case "ed25519":
        privateKeyPath = "$HOME/.ssh/id_ed25519"
    case "ecdsa":
         privateKeyPath = "$HOME/.ssh/id_ecdsa"
    // Add other types as needed
    // case "rsa": // Already covered by default
    // default: // Stick with the initial default if type is unknown/unspecified
    }
}

// Ansible automatically expands $HOME, so no specific Go code is needed here.

// Example format (adjust based on actual inventory structure):
inventoryContent += fmt.Sprintf("ansible_ssh_private_key_file=%s\n", privateKeyPath) // Use the determined path

// --- Continue generating the rest of the inventory ---
```

### 4. Update Service & Handler Layers

*   **`pkg/api/v1/handlers/instance.go`**: Ensure the `CreateInstance` handler correctly binds the new optional JSON fields (`ssh_key_type`, `ssh_key_path`) to the `types.InstanceRequest` struct during `c.BodyParser()`. No specific code changes should be needed if using standard struct tags.
*   **`internal/services/instance.go`**: Ensure the `CreateInstance` service method passes the relevant `InstanceRequest` object (or at least the `SSHKeyType` and `SSHKeyPath` fields) down to the compute layer function that generates the Ansible inventory.

## Benefits

*   Provides flexibility for users with different SSH key types and locations.
*   Maintains backward compatibility (defaults to RSA).
*   Keeps DigitalOcean public key configuration separate from Ansible private key configuration.
*   Changes are localized primarily to the type definition and the Ansible inventory generation logic. 