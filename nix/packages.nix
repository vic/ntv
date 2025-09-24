{ ... }:
{
  perSystem =
    { pkgs, ... }:
    let

      ntv = pkgs.buildGoModule {
        pname = "ntv";
        src = ./..;
        version = pkgs.lib.trim (builtins.readFile ./../packages/app/VERSION);
        vendorHash = "sha256-HvOwS4Tpv3nL4zRFf4L/SR9T3vwy5TJlHKfC/5Yq3AE=";
        meta = with pkgs.lib; {
          description = "Nix Tool Versions";
          homepage = "https://github.com/vic/ntv";
          mainProgram = "ntv";
        };
      };
    in
    {

      packages = {
        default = ntv;
        inherit ntv;
      };

      checks = { inherit ntv; };

    };
}
