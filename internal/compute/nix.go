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
	// Esperar a que SSH esté disponible
	fmt.Printf("⏳ Waiting for SSH to be available on %s...\n", host)
	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("ssh",
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("root@%s", host),
			"echo 'SSH is ready'")

		if err := checkCmd.Run(); err == nil {
			break
		}

		if i == 29 {
			return fmt.Errorf("timeout waiting for SSH to be ready")
		}

		fmt.Printf("  Retrying in 10 seconds... (%d/30)\n", i+1)
		time.Sleep(10 * time.Second)
	}

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

	// Descargar e instalar Nix directamente
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

	// Configurar Nix
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

// ApplyConfiguration applies the Nix configuration to a host
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
		"nixos-rebuild test && nixos-rebuild switch") // test primero para validar

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

	// Instalar nixos-install
	fmt.Println("📦 Installing NixOS tools...")
	installCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`source /etc/profile && \
		 nix-env -iA \
		   nixos.nixos-install-tools \
		   nixos.nixos-rebuild \
		   nixos.nix`)

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install NixOS tools: %v", err)
	}

	// Crear directorios necesarios
	createDirsCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`mkdir -p /etc/nixos && \
		 touch /etc/nixos/configuration.nix && \
		 chmod 644 /etc/nixos/configuration.nix`)

	createDirsCmd.Stdout = os.Stdout
	createDirsCmd.Stderr = os.Stderr

	if err := createDirsCmd.Run(); err != nil {
		return fmt.Errorf("failed to create NixOS directories: %v", err)
	}

	// Crear archivo base.nix
	baseNixCmd := exec.Command("ssh",
		"-i", sshKeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", host),
		`cat > /etc/nixos/base.nix << 'EOL'
{ config, pkgs, ... }:
{
  imports = [ ];
  
  boot.loader.grub.enable = true;
  boot.loader.grub.version = 2;
  
  networking.useDHCP = true;
  
  services.openssh.enable = true;
  services.openssh.permitRootLogin = "yes";
  
  users.users.root.openssh.authorizedKeys.keys = [
    "$(cat ~/.ssh/authorized_keys)"
  ];
  
  system.stateVersion = "23.11";
}
EOL`)

	baseNixCmd.Stdout = os.Stdout
	baseNixCmd.Stderr = os.Stderr

	if err := baseNixCmd.Run(); err != nil {
		return fmt.Errorf("failed to create base.nix: %v", err)
	}

	fmt.Printf("✅ NixOS preparation completed on %s\n", host)
	return nil
}

// InstallNixOS performs the NixOS installation
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
