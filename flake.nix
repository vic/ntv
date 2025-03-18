{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } (
      { ... }:
      {
        systems = import inputs.systems;

        perSystem =
          { pkgs, ... }:
          let
            nix-versions = pkgs.buildGoModule {
              pname = "nix-versions";
              version = "1.0.0";
              src = ./src;
              vendorHash = "sha256-asGQka4gkEHMLz/lncQwS4liugOIqVCh1H6dB3+snoQ=";
              meta = with pkgs.lib; {
                description = "CLI for searching nix packages versions using lazamar or nixhub, written in Go";
                homepage = "https://github.com/vic/nix-versions";
                # license = licenses.apache-2;
                # maintainers = with maintainers; [ vic ];
              };
            };

          in
          {
            packages = {
              default = nix-versions;
              inherit nix-versions;
            };
          };
      }
    );
}
