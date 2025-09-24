#!/usr/bin/env bash
set -euo pipefail

load "$BATS_LIB"

function setup() {
  INIT="ntv init --channel nixos-25.05 --override-ntv path:$PROJECT_ROOT/nix/flakeModules"
  DIR="$(mktemp -d)"
  cd $DIR
  true
}

function teardown() {
  rm -rf $DIR
}

@test 'create a flake' {
  run $INIT hello@latest
  assert_success
  assert_output -p "This file was generated"
  assert_output -p "inputs.\"ntv\".url = \"path:$PROJECT_ROOT/nix/flakeModules\";"
  assert_output -p "inputs.\"hello\".url = \"nixpkgs/"

  # lock inputs
  echo "$output" >flake.nix
  run nix flake lock
  assert_success

  # program is exposed on flake
  run nix run .#hello
  assert_success
  assert_output -p "Hello, world!"

  # test development shell
  run nix develop -c hello
  assert_success
  assert_output -p "Hello, world!"
}
