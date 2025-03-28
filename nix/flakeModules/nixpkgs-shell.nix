# Uses nixpkgs.mkShell to provide a shell with the versioned tools.
{ lib, config, ... }:
{

  options.ntv.mkShell.enabled = lib.mkEnableOption {
    description = "Enable mkShell integration.";
    default = true;
  };

  options.ntv.mkShell.attrs = lib.mkOption {
    description = "Additional attributes to pass to pkgs.mkShell";
    type = lib.types.attrs;
    default = { };
  };

  config.perSystem =
    { pkgs, self', ... }:
    {
      devShells.nixpkgs = pkgs.mkShell (
        {
          name = "nixpkgs-shell";
          buildInputs = pkgs.lib.attrValues self'.packages.default.versioned;
        }
        // config.ntv.mkShell.attrs
      );
    };

}
