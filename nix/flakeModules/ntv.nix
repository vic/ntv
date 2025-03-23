{ lib, config, ... }:
{
  options.ntv.flake =
    with lib;
    with types;
    mkOption {
      type = unspecified;
    };

  config.flake.lib.ntv.flake = config.ntv.flake;
}
