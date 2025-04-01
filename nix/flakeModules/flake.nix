{
  outputs = _inputs: {
    flakeModules.default = ./default.nix;
    flakeModule = ./default.nix;
  };
}
