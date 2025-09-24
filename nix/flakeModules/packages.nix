# Produces a package-set for on eachSystem, named `ntv` containing the pinned-versions tools.
{ inputs, config, ... }:
{
  perSystem =
    { pkgs, inputs', ... }:
    let
      # Each config.ntv.tools entry is a tool with a specific version.
      # it has a corresponding input on the flake.
      #
      # We use the tool attribute to access the tool's versioned package.
      # And place it under the same name in the package set.
      getTool =
        _name: tool:
        let
          inputHasPackages = inputs.${tool.name} ? packages;
          input = inputs'.${tool.name};
          inputPkgs = if inputHasPackages then input.packages else input.legacyPackages;
          parts = pkgs.lib.splitString "#" tool.installable;
          attrPath = pkgs.lib.last parts;
          pkgPath = pkgs.lib.splitString "." attrPath;
          pkg = pkgs.lib.getAttrFromPath pkgPath inputPkgs;
        in
        pkg;

      ntv = pkgs.lib.mapAttrs getTool config.ntv.tools;

      # An installable environment containing all the tools.
      env = pkgs.buildEnv {
        name = "ntv";
        paths = pkgs.lib.attrValues ntv;
      };

      packages = ntv // {
        default = env // {
          versioned = ntv;
        };
      };
    in
    {
      inherit packages;
    };
}
