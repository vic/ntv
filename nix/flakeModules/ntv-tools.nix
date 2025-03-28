{ lib, ... }:
{
  options.ntv.tools = lib.mkOption {
    description = "Tools pinned to specific versions by ntv.";
    type = lib.types.attrsOf (
      lib.types.submodule {
        options = {
          spec = lib.mkOption {
            type = lib.types.str;
            description = "The original spec given to ntv.";
          };
          name = lib.mkOption {
            type = lib.types.str;
            description = "The resolved name of the package.";
          };
          version = lib.mkOption {
            type = lib.types.str;
            description = "The resolved version of the package.";
          };
          installable = lib.mkOption {
            type = lib.types.str;
            description = "The nix installable in the form: flake#attrPath.";
          };
        };
      }
    );
  };

}
