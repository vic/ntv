{ inputs, ... }:
{
  imports = [ inputs.devshell.flakeModule ];
  perSystem =
    { pkgs, self', ... }:
    {
      devshells.default =
        { ... }:
        {
          imports = [ "${inputs.devshell}/extra/git/hooks.nix" ];

          git.hooks.enable = true;
          git.hooks.pre-push.text = "nix flake check";

          devshell.packages = [ pkgs.gopls ];
          devshell.packagesFrom = [ self'.packages.default ];
        };
    };
}
