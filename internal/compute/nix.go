// Package compute provides infrastructure and system configuration capabilities.
package compute

// This file implements NixOS configuration and management functionality.
// It provides tools for:
//   - Installing and configuring Nix package manager
//   - Installing and managing NixOS systems
//   - Applying NixOS configurations to remote hosts
//   - Managing SSH connections and remote command execution
//
// Key components:
//   - NixConfigurator: Main struct for managing Nix/NixOS configurations
//   - ConfigOptions: Configuration options for NixOS systems
//
// Common operations:
//   - InstallNix: Installs the Nix package manager on a remote host
//   - PrepareNixOS: Prepares or updates Nix installation
//   - InstallNixOS: Performs NixOS installation
//   - ApplyConfiguration: Applies NixOS configuration to a host
//   - UpdateSystem: Updates an existing NixOS system
//
// Usage example:
//   configurator := NewNixConfigurator()
//   err := configurator.InstallNix("192.168.1.100", "~/.ssh/id_rsa")
//   err = configurator.ApplyConfiguration("192.168.1.100", "~/.ssh/id_rsa", nixConfig)

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// NixConfigurator handles system configuration using Nix
type NixConfigurator struct {
	// Channel to use for nix installations
	channel string
}

// NewNixConfigurator creates a new Nix configuration manager
func NewNixConfigurator() *NixConfigurator {
	return &NixConfigurator{
		channel: "nixos-23.11", // Default channel
	}
}

// InstallNix installs Nix on the target machine
func (n *NixConfigurator) InstallNix(
	host string,
	sshKeyPath string,
) error {
	// Wait for SSH to be available
	fmt.Printf("⏳ Waiting for SSH to be available on %s...\n", host)
	for i := 0; i < 60; i++ {
		checkCmd := exec.Command("ssh",
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=10",
			"-o", "ServerAliveInterval=5",
			"-o", "ServerAliveCountMax=3",
			fmt.Sprintf("root@%s", host),
			"echo 'SSH is ready' && test -w /root")

		if err := checkCmd.Run(); err == nil {
			fmt.Printf("✅ SSH connection established to %s\n", host)
			break
		}

		if i == 59 {
			return fmt.Errorf("timeout waiting for SSH to be ready")
		}

		fmt.Printf("  Retrying in 10 seconds... (%d/60)\n", i+1)
		time.Sleep(10 * time.Second)
	}

	// Give the system a moment to stabilize
	fmt.Println("⏳ Waiting for system to stabilize...")
	time.Sleep(15 * time.Second)

	// Verificar si Nix ya está instalado
	checkCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		"command -v nix")

	if err := checkCmd.Run(); err == nil {
		fmt.Printf("✅ Nix already installed on %s, skipping installation\n", host)
		return nil
	}

	fmt.Printf("🔧 Installing Nix on %s...\n", host)

	// Clean up any previous Nix installation files
	cleanupCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`rm -f /etc/bash.bashrc.backup-before-nix \
		 /etc/profile.backup-before-nix \
		 /etc/zshrc.backup-before-nix \
		 /etc/bashrc.backup-before-nix \
		 /etc/profile.d/nix.sh.backup-before-nix \
		 /etc/profile.d/nix.sh \
		 && \
		 rm -rf /nix /etc/nix ~/.nix* /root/.nix* && \
		 mkdir -p /etc/profile.d && \
		 touch /etc/profile.d/nix.sh && \
		 systemctl daemon-reload`)

	cleanupCmd.Stdout = os.Stdout
	cleanupCmd.Stderr = os.Stderr

	if err := cleanupCmd.Run(); err != nil {
		fmt.Printf("⚠️ Warning: Failed to clean up previous installation: %v\n", err)
	}

	// Download and install Nix directly
	cmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`sh <(curl -L https://nixos.org/nix/install) --daemon --yes`)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Nix: %v", err)
	}

	// Configure Nix
	fmt.Println("⚙️ Configuring Nix...")
	configCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`source /etc/profile && \
		 nix-channel --add https://nixos.org/channels/nixos-unstable nixos && \
		 nix-channel --update`)

	configCmd.Stdout = os.Stdout
	configCmd.Stderr = os.Stderr

	if err := configCmd.Run(); err != nil {
		fmt.Printf("⚠️ Warning: Failed to configure Nix channels: %v\n", err)
	}

	fmt.Printf("✅ Nix installation completed on %s\n", host)
	return nil
}

// Apply Configuration applies the Nix configuration to a host
func (n *NixConfigurator) ApplyConfiguration(
	host,
	sshKeyPath,
	config string,
) error {
	resolvedPath, err := ResolvePath(sshKeyPath)
	if err != nil {
		return fmt.Errorf("failed to resolve SSH key path: %v", err)
	}

	fmt.Printf("🔧 Applying Nix configuration to %s...\n", host)

	// Create temporary configuration file
	tmpFile, err := os.CreateTemp("", "nix-config-*.nix")
	if err != nil {
		return fmt.Errorf("failed to create temp config: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(config); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	// Copy configuration to remote host
	copyCmd := exec.Command("scp",
		"-i", resolvedPath,
		"-o", "StrictHostKeyChecking=no",
		tmpFile.Name(),
		fmt.Sprintf("root@%s:/etc/nixos/configuration.nix", host))

	copyCmd.Stdout = os.Stdout
	copyCmd.Stderr = os.Stderr

	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy configuration: %v", err)
	}

	// Validate and apply configuration on the remote host
	applyCmd := exec.Command("ssh",
		"-i", resolvedPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`source /etc/profile && \
		 export PATH=$PATH:/nix/var/nix/profiles/default/bin && \
		 export NIX_PATH=nixpkgs=/nix/var/nix/profiles/per-user/root/channels/nixos:nixos-config=/etc/nixos/configuration.nix && \
		 nix-collect-garbage -d && \
		 rm -rf /nix/var/nix/profiles && \
		 mkdir -p /nix/var/nix/profiles && \
		 nix-env -iA nixos.nixos-rebuild && \
		 export PATH=$PATH:/root/.nix-profile/bin && \
		 free -h && \
		 TMPDIR=/tmp nixos-rebuild switch \
		   --option sandbox false \
		   --option cores 1 \
		   --max-jobs 1 \
		   --option binary-caches https://cache.nixos.org \
		   --option trusted-binary-caches https://cache.nixos.org \
		   --option binary-cache-public-keys cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY= \
		   --option use-substitutes true \
		   --option build-use-substitutes true \
		   --option enforce-substitutes true \
		   --option build-cores 1 \
		   --option system-features "" \
		   --no-build-output \
		   --show-trace`)

	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr

	return applyCmd.Run()
}

// PrepareNixOS prepares or updates the Nix package management system
func (n *NixConfigurator) PrepareNixOS(
	host string,
	sshKeyPath string,
) error {
	fmt.Printf("🔧 Preparing NixOS on %s...\n", host)

	// Create necessary directories first
	createDirsCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`mkdir -p /etc/nixos/cloud && \
		 touch /etc/nixos/configuration.nix && \
		 chmod 644 /etc/nixos/configuration.nix`)

	createDirsCmd.Stdout = os.Stdout
	createDirsCmd.Stderr = os.Stderr

	if err := createDirsCmd.Run(); err != nil {
		return fmt.Errorf("failed to create NixOS directories: %v", err)
	}

	// Copy NixOS configuration files
	fmt.Println("📁 Copying NixOS configuration files...")
	copyFilesCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`mkdir -p /etc/nixos/cloud`)

	if err := copyFilesCmd.Run(); err != nil {
		return fmt.Errorf("failed to create cloud directory: %v", err)
	}

	// Copy configuration files to their respective locations
	copyFilesCmd = exec.Command("scp",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-r",
		"nix/base.nix",
		fmt.Sprintf("root@%s:/etc/nixos/base.nix", host))

	if err := copyFilesCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy base.nix: %v", err)
	}

	copyFilesCmd = exec.Command("scp",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-r",
		"nix/cloud/digitalocean.nix",
		fmt.Sprintf("root@%s:/etc/nixos/cloud/digitalocean.nix", host))

	copyFilesCmd.Stdout = os.Stdout
	copyFilesCmd.Stderr = os.Stderr

	if err := copyFilesCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy NixOS configuration files: %v", err)
	}

	// Install NixOS tools first
	fmt.Println("📦 Installing NixOS tools...")
	nixosToolsCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`source /etc/profile && \
		 nix-channel --add https://nixos.org/channels/nixos-unstable nixos && \
		 nix-channel --update && \
		 nix-env -iA nixos.nixos-install nixos.nixos-rebuild`)

	nixosToolsCmd.Stdout = os.Stdout
	nixosToolsCmd.Stderr = os.Stderr

	if err := nixosToolsCmd.Run(); err != nil {
		return fmt.Errorf("failed to install NixOS tools: %v", err)
	}

	fmt.Printf("✅ NixOS preparation completed on %s\n", host)
	return nil
}

// Install NixOS performs the NixOS installation
func (n *NixConfigurator) InstallNixOS(
	host string,
	sshKeyPath string,
) error {
	resolvedPath, err := ResolvePath(sshKeyPath)
	if err != nil {
		return fmt.Errorf("failed to resolve SSH key path: %v", err)
	}

	fmt.Printf("📦 Installing NixOS on %s...\n", host)

	cmd := exec.Command("ssh",
		"-i", resolvedPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		"nixos-install --no-root-passwd")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// UpdateSystem updates the NixOS system
func (n *NixConfigurator) UpdateSystem(
	host string,
	sshKeyPath string,
) error {
	resolvedPath, err := ResolvePath(sshKeyPath)
	if err != nil {
		return fmt.Errorf("failed to resolve SSH key path: %v", err)
	}

	fmt.Printf("🔄 Updating NixOS on %s...\n", host)

	cmd := exec.Command("ssh",
		"-i", resolvedPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		"nixos-rebuild switch")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ConfigOptions represents configuration options for NixOS
type ConfigOptions struct {
	SSHKeys     []string
	ExtraConfig string
	Hostname    string
}

// LoadBaseConfig loads and customizes the base NixOS configuration
func (n *NixConfigurator) LoadBaseConfig(
	opts ConfigOptions,
) (string, error) {
	baseConfig, err := os.ReadFile("nix/base.nix")
	if err != nil {
		return "", fmt.Errorf("failed to read base config: %v", err)
	}

	// Create a temporary configuration with customizations
	config := fmt.Sprintf(`
{ config, pkgs, ... }:

let
	baseConfig = %s;
in
{
	imports = [ baseConfig ];

	# Custom configurations
	networking.hostName = "%s";

	# SSH Keys
	users.users.root.openssh.authorizedKeys.keys = [
		%s
	];

	# Extra configurations
	%s
}`, string(baseConfig), opts.Hostname, strings.Join(opts.SSHKeys, "\n        "), opts.ExtraConfig)

	return config, nil
}

// ResolvePath expands the home directory if the path starts with ~
func ResolvePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}

	if path == "~" {
		return usr.HomeDir, nil
	}

	return filepath.Join(usr.HomeDir, path[2:]), nil
}

// ApplyConfig aplica una configuración específica de Nix a una instancia
func (n *NixConfigurator) ApplyConfig(host string, config string) error {
	// Crear archivo de configuración temporal
	tmpFile := fmt.Sprintf("/tmp/nix-config-%s.nix", strings.ReplaceAll(host, ".", "-"))
	cmd := exec.Command("ssh",
		"-i", "~/.ssh/id_rsa",
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		fmt.Sprintf("echo '%s' > %s", config, tmpFile))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	// Aplicar la configuración
	cmd = exec.Command("ssh",
		"-i", "~/.ssh/id_rsa",
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		fmt.Sprintf("nixos-rebuild switch -I nixos-config=%s", tmpFile))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply config: %v", err)
	}

	return nil
}
