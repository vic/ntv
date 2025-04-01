{
  inputs = {
    nixpkgs.url = "nixpkgs";
    systems.url = "github:nix-systems/default";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell.url = "github:numtide/devshell";
    devshell.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;
      flake.flakeModules.default = ./default.nix;
      flake.flakeModules.devenv-shell = ./devenv-shell.nix;
      flake.flakeModules.nixpkgs-shell = ./nixpkgs-shell.nix;
    };
}
