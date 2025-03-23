{ inputs, ntv, ... }:
let

  nv-packages-set =
    pkgs:
    let
      inherit (pkgs) lib;
      toAttr =
        nv:
        let
          name-path = lib.splitString "." nv.name;
          attr-path = lib.splitString "." nv.attr_path;
          pkgs' =
            inputs.${nv.name}.packages.${pkgs.system} or inputs.${nv.name}.legacyPackages.${pkgs.system};
          pkg = lib.getAttrFromPath attr-path pkgs';
          attr = lib.setAttrByPath name-path pkg;
        in
        attr;
      attrs = lib.map toAttr ntv;
      packages = lib.foldl lib.recursiveUpdate { } attrs;
    in
    packages;

  nv-packages-list =
    pkgs:
    let
      inherit (pkgs) lib;
      set = nv-packages-set pkgs;
      list = lib.map (nv: lib.getAttrFromPath (lib.splitString "." nv.name) set) ntv;
    in
    list;

  overlay = _final: prev: nv-packages-set prev;

  packagesEnv =
    pkgs:
    pkgs.buildEnv {
      name = "ntv";
      paths = nv-packages-list pkgs;
    };

  devShell =
    pkgs:
    pkgs.mkShell {
      buildInputs = [ (packagesEnv pkgs) ];
    };

in
{
  flake.lib.ntv = ntv;
  flake.overlays.default = overlay;

  perSystem = (
    { pkgs, ... }:
    {
      packages = nv-packages-set pkgs;
      devShells.default = devShell pkgs;
    }
  );

}
