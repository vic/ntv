{
  outputs = inputs: import ./nix inputs;
  inputs = {
    nixpkgs.url = "nixpkgs";
    systems.url = "github:nix-systems/default";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    treefmt-nix.inputs.nixpkgs.follows = "nixpkgs";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell.url = "github:numtide/devshell";
    devshell.inputs.nixpkgs.follows = "nixpkgs";
    bats-support.url = "github:ztombol/bats-support";
    bats-support.flake = false;
    bats-assert.url = "github:bats-core/bats-assert";
    bats-assert.flake = false;
  };
}
