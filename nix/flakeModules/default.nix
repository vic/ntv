{ inputs, ... }:
{
  imports = [
    inputs.ntv.inputs.flake-parts.flakeModules.flakeModules
    inputs.ntv.inputs.flake-parts.flakeModules.modules
    ./ntv.nix
    ./packages.nix
    ./overlays.nix
    ./nixpkgs-shell.nix
    ./devshell-shell.nix
    ./devenv-shell.nix
    ./default-shell.nix
  ];

  config.flake.modules = { };
}
