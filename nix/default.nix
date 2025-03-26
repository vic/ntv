inputs:
inputs.flake-parts.lib.mkFlake { inherit inputs; } {
  systems = import inputs.systems;
  imports = [
    ./shell.nix
    ./packages.nix
    ./treefmt.nix
  ];
  flake.flakeModules.default = ./flakeModules;
}
