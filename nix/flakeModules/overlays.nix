{ inputs, ... }:

{

  flake.overlays.default = _final: prev: inputs.self.packages.${prev.system}.default.versioned;

}
