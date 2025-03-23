{ ... }:
{

  imports = [
    ./ntv.nix
  ];

  perSystem = (
    { inputs', ... }:
    {
      packages.nvs = inputs'.ntv.packages.default;
    }
  );

}
