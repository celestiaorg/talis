{ config, pkgs, ... }:

{
  imports = [ 
    # Import cloud specific configurations if needed
    ./cloud/digitalocean.nix 
  ];

  # System packages
  environment.systemPackages = with pkgs; [
    # System tools
    vim
    git
    curl
    wget
    htop
    tmux
    tree
    
    docker.io
    docker-compose

    # Network tools
    net-tools
    inetutils
    mtr
    tcpdump
  ];

  # Nix configuration
  nix = {
    settings = {
      auto-optimise-store = true;
      experimental-features = [ "nix-command" "flakes" ];
    };
    gc = {
      automatic = true;
      dates = "weekly";
      options = "--delete-older-than 30d";
    };
  };

  # Common shell configurations
  programs.bash.enableCompletion = true;
  programs.zsh.enable = true;
} 