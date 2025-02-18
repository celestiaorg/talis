{ config, pkgs, ... }:

{
  imports = [ ];

  # System packages
  environment.systemPackages = with pkgs; [
    # System tools
    vim
    git
    curl
    wget
  ];

  # Nix configuration
  nix = {
    settings = {
      auto-optimise-store = true;
      cores = 1;
      max-jobs = 1;
      sandbox = false;
      substituters = [ "https://cache.nixos.org" ];
      trusted-substituters = [ "https://cache.nixos.org" ];
      trusted-public-keys = [ "cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY=" ];
    };
    gc = {
      automatic = true;
      dates = "weekly";
      options = "--delete-older-than 30d";
    };
  };

  # Common shell configurations
  programs.bash.completion.enable = true;

  boot.loader.grub.enable = true;

  services.openssh.enable = true;
  services.openssh.settings.PermitRootLogin = "yes";

  users.users.root.openssh.authorizedKeys.keys = [
    "$(cat ~/.ssh/authorized_keys)"
  ];

  system.stateVersion = "23.11";
} 