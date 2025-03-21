nv-inputs@{ flake-parts, systems, ... }:
flake-parts.lib.mkFlake { inputs = nv-inputs; } {
  systems = import systems;
  imports = [
    ./packages.nix
    ./treefmt.nix
  ];
  flake.lib.mkFlake =
    { inputs, nix-versions }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;
      imports = [
        { _module.args = { inherit nix-versions; }; }
        ./flakeModules/nix-versions.nix
      ];
    };
}
