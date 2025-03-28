# This flake-module defines a schema for the
# ntv flake generator.
#
# This schema is mirrored by types at flake.go source.
# Having this schema exported in at the output `lib.ntv`
# so that ntv can later load it and modify it as needed
# in the go runtime. And possibly generate an updated flake.
#
# config.ntv is also used to generate a package set
# from the versioned tools in the flake. see ./package-set.nix.
# building on this package set, ntv can also export
# packages and development shells of different types.
{ lib, config, ... }:
{
  config.flake.lib.ntv = config.ntv;

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

  options.ntv.flake = lib.mkOption {
    description = "Flake generator data for ntv.";
    type = lib.types.submodule {
      options = {
        inputs = lib.mkOption {
          description = "The inputs of flake.";
          type = lib.types.listOf (
            lib.types.submodule {
              options = {
                name = lib.mkOption {
                  type = lib.types.str;
                  description = "The name for a dependency";
                };
                url = lib.mkOption {
                  type = lib.types.str;
                  description = "A flake compatible url";
                };
                flake = lib.mkOption {
                  type = lib.types.bool;
                  description = "Marks the input as a flake.";
                };
                follows = lib.mkOption {
                  description = "List of follow overrides for the input.";
                  type = lib.types.nullOr (
                    lib.types.listOf (
                      lib.types.submodule {
                        options = {
                          input = lib.mkOption {
                            type = lib.types.str;
                            description = "Name of this input's input to override";
                          };
                          follow = lib.mkOption {
                            type = lib.types.str;
                            description = "Name of the flake's input to follow";
                          };
                        };
                      }
                    )
                  );
                };
              };
            }
          );
        };
        mkFlake = lib.mkOption {
          type = lib.types.str;
          description = "A NixExpr. How to create the flake. A function like flake-parts.lib.mkFlake";
        };
        systems = lib.mkOption {
          type = lib.types.str;
          description = "A NixExpr. Producing an list of systems for the flake.";
        };
        imports = lib.mkOption {
          type = lib.types.listOf lib.types.str;
          description = "List of NixExpr. Additional flakeModules to import.";
        };
      };
    };
  };

}
