{
  inputs,
  config,
  lib,
  ...
}:

{
  options.ntv.devenv.enabled = lib.mkOption {
    description = "Enable devenv integration.";
    default = inputs ? devenv;
    type = lib.types.bool;
  };

  imports = [
    (inputs.devenv or inputs.ntv.inputs.devenv).flakeModule

    (lib.mkIf config.ntv.devenv.enabled {
      perSystem =
        { pkgs, self', ... }:
        {
          devenv.shells.devenv =
            { ... }:
            {
              imports = [
                (config.flake.modules.devenv or { })
              ];

              packages = pkgs.lib.attrValues self'.packages.default.versioned;
            };
        };
    })
  ];

}
