{
  perSystem =
    { pkgs, ... }:
    let

      ntv = pkgs.buildGoModule {
        pname = "ntv";
        src = ./..;
        version = pkgs.lib.trim (builtins.readFile ./../packages/app/VERSION);
        vendorHash = "sha256-Am1HLlNbj86lod0yL8ge5tGtaffMMOquiNWPY00hZ3E=";
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

      checks.ntv = ntv;

    };
}
