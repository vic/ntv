{ inputs, ... }:

{

  flake.overlays.default = final: prev: inputs.self.packages.${prev.system}.default.versioned;
  
}
