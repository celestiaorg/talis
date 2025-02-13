{ config, pkgs, ... }:

{
  imports = [ ../base.nix ];

  # DigitalOcean specific configurations
  boot.loader.grub.device = "/dev/vda";
  
  # Enable qemu-guest-agent
  services.qemuGuest.enable = true;

  # Configure cloud-init
  services.cloud-init = {
    enable = true;
    network.enable = true;
  };

  # Optimize for cloud environment
  services.openssh.enable = true;
  services.openssh.settings.PasswordAuthentication = false;

  # DigitalOcean specific configurations
  environment.systemPackages = with pkgs; [
    # DO monitoring agent
    do-agent
    
    # Common cloud tools
    cloud-init
    cloud-utils
  ];

  # Network optimizations for DO
  networking = {
    # Enable TCP BBR for better network performance
    tcp.bbr.enable = true;
    
    # Firewall configuration
    firewall = {
      enable = true;
      allowedTCPPorts = [ 22 80 443 ];
    };
  };

  # System optimizations for cloud environment
  boot.kernel.sysctl = {
    # Network optimizations
    "net.ipv4.tcp_fastopen" = 3;
    "net.ipv4.tcp_slow_start_after_idle" = 0;
    
    # Virtual Memory settings
    "vm.swappiness" = 10;
    "vm.dirty_ratio" = 10;
  };

  # Monitoring and logging
  services = {
    # Enable DO monitoring
    do-agent.enable = true;

    # Configure log rotation
    logrotate = {
      enable = true;
      settings = {
        "/var/log/messages" = {
          rotate = 7;
          weekly = true;
          compress = true;
        };
      };
    };
  };
} 