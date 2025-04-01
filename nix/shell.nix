{ inputs, ... }:
{
  imports = [ inputs.devshell.flakeModule ];
  perSystem =
    { pkgs, self', ... }:
    let
      bats-lib = pkgs.writeTextFile {
        name = "bats-lib.bash";
        text = ''
          source ${inputs.bats-support}/load.bash
          source ${inputs.bats-assert}/load.bash
        '';
      };

      gotest = pkgs.writeShellApplication {
        name = "gotest";
        runtimeInputs = [
          pkgs.go
          pkgs.findutils
        ];
        text = ''
          go mod vendor
          go mod tidy
          find ./packages -type d -print0 | xargs -0 go test
        '';
        meta.description = "Run all go tests";
      };

      go-ntv = pkgs.writeShellApplication {
        name = "ntv";
        runtimeInputs = [
          pkgs.go
        ];
        text = ''
          (
            cd "''${PROJECT_ROOT:-"$(git rev-parse --show-toplevel)"}"/
            go run main.go "$@"
          )
        '';
        meta.description = "ntv (development version)";
      };

      e2e = pkgs.writeShellApplication {
        name = "e2e";
        runtimeInputs = [
          go-ntv
          pkgs.go
          pkgs.bats
          pkgs.nix
          pkgs.coreutils
          pkgs.jq
        ];
        text = ''
          export BATS_LIB=${bats-lib}
          export LANG=en_US.UTF-8
          PROJECT_ROOT="$(git rev-parse --show-toplevel)"
          export PROJECT_ROOT
          go mod vendor
          go mod tidy
          bats e2e "$@"
        '';
        meta.description = "Run e2e tests (requires network)";
      };
    in
    {
      devshells.default =
        { ... }:
        {
          imports = [ "${inputs.devshell}/extra/git/hooks.nix" ];

          commands = [
            { package = e2e; }
            { package = gotest; }
            { package = go-ntv; }
          ];

          git.hooks.enable = true;
          git.hooks.pre-push.text = ''
            set -e
            nix flake check
            ${pkgs.lib.getExe e2e}
          '';

          devshell.packages = [
            pkgs.go
            pkgs.gopls
          ];
          devshell.packagesFrom = [ self'.packages.default ];
        };
    };
}
