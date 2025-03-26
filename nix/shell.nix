{ inputs, ... }:
{
  imports = [ inputs.devshell.flakeModule ];
  perSystem =
    { pkgs, self', ... }:
    {
      devshells.default =
        { ... }:
        {
          devshell.packages = [ pkgs.gopls ];
          devshell.packagesFrom = [ self'.packages.default ];
        };
    };
}
