{
  inputs,
  config,
  lib,
  ...
}:
{
  options.ntv.devshell.enabled = lib.mkOption {
    description = "Enable devshell integration.";
    default = true;
    type = lib.types.bool;
  };

  imports = [
    (inputs.devshell or inputs.ntv.inputs.devshell).flakeModule

    (lib.mkIf config.ntv.devshell.enabled {
      perSystem =
        { pkgs, self', ... }:
        {
          devshells.devshell =
            { ... }:
            {
              imports = [
                (config.flake.modules.devshell or { })
              ];

              commands = pkgs.lib.mapAttrsToList (_name: pkg: {
                package = pkg;
              }) self'.packages.default.versioned;
            };
        };
    })
  ];

}
