{ config, lib, ... }:

{
  options.ntv.defaultShell = lib.mkOption {
    description = "The default shell to use.";
    default = "devshell";
    type = lib.types.str;
  };

  config.perSystem =
    { self', ... }:
    {
      devShells.default = self'.devShells.${config.ntv.defaultShell};
    };

}
