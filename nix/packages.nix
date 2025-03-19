{
  perSystem =
    { pkgs, ... }:
    let
      nix-versions = pkgs.buildGoModule {
        pname = "nix-versions";
        src = ./..;
        version = pkgs.lib.trim (builtins.readFile ./../packages/app/VERSION);
        vendorHash = "sha256-oxqRIk6WgZhioT8ysBehTNALlckhEiWVAKu/CLfI7dU=";
        meta = with pkgs.lib; {
          description = "CLI for searching nix packages versions using lazamar or nixhub, written in Go";
          homepage = "https://github.com/vic/nix-versions";
        };
        nativeBuildInputs = [ pkgs.makeWrapper ];
        postInstall = ''
          wrapProgram $out/bin/nix-versions \
            --prefix PATH : ${with pkgs; lib.makeBinPath [ nix ]}
        '';
      };

    in
    {

      packages = {
        default = nix-versions;
        inherit nix-versions;
      };

    };
}
