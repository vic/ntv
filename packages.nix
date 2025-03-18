{
  perSystem =
    { pkgs, ... }:
    let
      nix-versions = pkgs.buildGoModule {
        pname = "nix-versions";
        version = builtins.readFile ./VERSION;
        src = ./src;
        vendorHash = "sha256-asGQka4gkEHMLz/lncQwS4liugOIqVCh1H6dB3+snoQ=";
        meta = with pkgs.lib; {
          description = "CLI for searching nix packages versions using lazamar or nixhub, written in Go";
          homepage = "https://github.com/vic/nix-versions";
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
