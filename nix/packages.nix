{
  perSystem =
    { pkgs, ... }:
    let

      ntv = pkgs.buildGoModule {
        pname = "ntv";
        src = ./..;
        version = pkgs.lib.trim (builtins.readFile ./../packages/app/VERSION);
        vendorHash = "sha256-Vj+LUF7k0KlPv1uhSIfGU6i1CMBz3+llL/UK1G9FaZw=";
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

    };
}
