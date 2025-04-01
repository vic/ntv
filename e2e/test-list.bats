#!/usr/bin/env bash
set -euo pipefail

load "$BATS_LIB"

function setup() {
  LIST="ntv list --channel nixos-24.05"
  ONE="$LIST --one"
  DIR="$(mktemp -d)"
  cd $DIR
  true
}

function teardown() {
  rm -rf $DIR
}

assert_lines() {
  test "$1" == "${#lines[@]}"
}

assert_header() {
  assert_line -n 0 -e "Name\s+Version\s+NixInstallable\s+VerBackend"
}
assert_item() {
  n="$1"
  name="${2:-"[a-zA-Z.]+"}"
  version="${3:-"[0-9a-zA-Z.-]+"}"
  flake="${4:-"nixpkgs/[0-9a-zA-Z]+"}"
  attr="${5:-"[a-zA-Z0-9.,^-]+"}"
  back="${6:-"lazamar:nixos-24.05"}"
  assert_line -n "$n" -e "$name\s+$version\s+$flake#$attr\s+$back"
}

@test 'help' {
  run ntv list --help
  assert_success
  assert_output -p "List Nix packages versions."
  assert_output -p "SYNOPSIS"
  assert_output -p "OPTIONS"
  assert_output -p "https://nix-versions.alwaysdata.net"
}

@test 'list all by default' {
  run $LIST hello
  assert_success
  assert_header
  assert_item 1 hello 1.0.0.2 "" haskellPackages.hello
  assert_item 4 hello 2.12.1 "" hello
  assert_lines 5
}

@test 'list -1' {
  run $LIST hello -1
  assert_success
  assert_header
  assert_item 1 hello 2.12.1 "" hello
  assert_lines 2
}

@test 'list --one' {
  run $ONE hello
  assert_success
  assert_header
  assert_item 1 hello 2.12.1 "" hello
  assert_lines 2
}

@test 'default backend is nixhub' {
  run ntv list hello -1
  assert_success
  assert_header
  assert_item 1 hello "" "" hello nixhub
  assert_lines 2
}

@test 'change default backend to nixhub' {
  run $ONE hello --nixhub
  assert_success
  assert_header
  assert_item 1 hello "" "" hello nixhub
}

@test 'change default backend to history' {
  run $ONE hello --history
  assert_success
  assert_header
  assert_item 1 hello "" "" hello history
}

@test 'change default backend to lazamar channel' {
  run $ONE hello --channel nixpkgs-unstable
  assert_success
  assert_header
  assert_item 1 hello "" "" hello lazamar:nixpkgs-unstable
}

@test 'change backend per spec' {
  run $ONE lazamar:nixos-24.05:hello lazamar:nixpkgs-unstable:hello nixhub:hello history:hello system:hello
  assert_success
  assert_header
  assert_item 1 hello "" "" hello lazamar:nixos-24.05
  assert_item 2 hello "" "" hello lazamar:nixpkgs-unstable
  assert_item 3 hello "" "" hello nixhub
  assert_item 4 hello "" "" hello history
  assert_item 5 hello "" nixpkgs hello system
  assert_lines 6
}

@test 'flake ref' {
  run $LIST github:vic/nix-versions/v1.0.0
  assert_success
  assert_header
  assert_item 1 nix-versions 1.0.0 github:vic/nix-versions/v1.0.0 default flake
  assert_lines 2
}

@test 'non existing package' {
  run $LIST does-not-exist
  assert_failure
  assert_output -p "no packages found for attribute-path \`does-not-exist\`. try using \`*does-not-exist*\`"
}

@test 'version semver constraint' {
  run $LIST 'hello@ >1 <2.12.1 '
  assert_success
  assert_header
  assert_item 1 hello 2.10
  assert_item 2 hello 2.12
  assert_lines 3
}

@test 'version regex constraint' {
  run $LIST 'hello@.*2$'
  assert_success
  assert_header
  assert_item 1 hello 1.0.0.2 "" haskellPackages.hello
  assert_item 2 hello 2.12 "" hello
  assert_lines 3
}

@test 'all' {
  run $LIST hello@2.12 --all
  assert_success
  assert_header
  assert_item 3 hello 2.12 "" hello
  assert_lines 5
}

@test 'preserve output selector' {
  run $ONE hello^out,dev@2.10
  assert_success
  assert_header
  assert_item 1 hello 2.10 "" hello\\^out,dev
  assert_lines 2
}

@test 'flake ref version constrained' {
  run $LIST github:vic/nix-versions/v1.0.0@2.0.0
  assert_success
  assert_header
  assert_lines 1 # no results since flake contains only 1.0.0
}

@test 'search packages providing exact program' {
  run $ONE bin/emacsclient
  assert_success
  assert_header
  assert_item 1 emacs-pgtk
  assert_item 2 emacs-nox
  assert_item 3 emacs-gtk
  assert_item 4 emacs
  assert_lines 5
}

@test 'search packages providing glob attribute-path' {
  run $ONE 'nix-search-c*'
  assert_success
  assert_header
  assert_item 1 nix-search-cli
  assert_lines 2
}

@test 'search packages providing glob program' {
  run $ONE 'bin/*leam'
  assert_success
  assert_header
  assert_item 1 gleam
  assert_lines 2
}

@test 'installable output' {
  run $ONE hello --installable
  assert_success
  assert_lines 1
  assert_output -e "nixpkgs/\w+#hello"
}

@test 'installable output preserves out selector' {
  run $ONE hello^out -i
  assert_success
  assert_lines 1
  assert_output -e "nixpkgs/\w+#hello\\^out"
}

@test 'read specs from file clean' {
  echo "hello@2.10" >$DIR/file
  run $ONE -r $DIR/file
  assert_success
  assert_item 1 hello 2.10
}

@test 'read specs from file dirty' {
  echo -e "# comment \n   #comment\n  \n\nhello@  2.10  # comment" >$DIR/file
  run $ONE -r $DIR/file
  assert_success
  assert_item 1 hello 2.10
}

@test 'read specs from file asdf tool-versions' {
  echo -e "# comment \n   #comment\n  \n\nhello\t 2.10  # comment" >$DIR/tool-versions
  run $ONE -r $DIR/tool-versions
  assert_success
  assert_item 1 hello 2.10
}

@test 'json output' {
  run $ONE hello@2.10 --json
  assert_success
  assert_line -p '"installable": "'
}

@test 'flake output' {
  run $ONE hello@2.10 --flake
  assert_success
  assert_line -p 'inputs."hello".url = "nixpkgs/'
}
