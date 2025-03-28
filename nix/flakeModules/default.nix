{ inputs, config, ... }:
{
  imports = [
    inputs.ntv.inputs.flake-parts.flakeModules.flakeModules
    inputs.ntv.inputs.flake-parts.flakeModules.modules
    ./ntv-flake.nix
    ./ntv-tools.nix
    ./packages.nix
    ./overlays.nix
    ./nixpkgs-shell.nix
    ./devshell-shell.nix
    ./devenv-shell.nix
    ./default-shell.nix
  ];

  config.flake.lib.ntv = config.ntv;
  config.flake.modules = { };
}
