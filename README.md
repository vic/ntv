# nix-versions - Search nix packages versions with lazamar/nix-versions or nixhub

This is a minimal CLI app written in Go that interfaces with [https://lazamar.co.uk/nix-versions/](https://lazamar.co.uk/nix-versions/) and [nixhub](https://nixhub.io).

## Usage

```
# Use directly from flake
nix run github:vic/nix-versions -- --help

# or install it on your profile.
nix profile install github:vic/nix-versions
nix-versions --help
```

#### Examples

```
# List known versions of emacs
nix-versions emacs 

# Return only the most recent version
nix-versions --limit 1 emacs

# Only versions between 24 and 27. Output JSON
nix-versions --constraint '>= 25 <= 27' --json emacs

# Latest of 29 series.
nix-versions --constraint '~29' --limit 1 emacs

# Show versions of pip from nixhub.io in the order that nixhub returns them
nix-versions --nixhub --sort=false python312Packages.pip

# Use release channel `nixpkgs/nixos-24.05` (using lazamar search)
nix-versions --channel nixos-24.05 python312Packages.pip
```

#### nix-versions --help

```
NAME:
   nix-versions - show available nix packages versions

USAGE:
   nix-versions [global options] PKG_ATTRIBUTE_NAME

AUTHOR:
   Victor Hugo Borja <vborja@apache.org>

GLOBAL OPTIONS:
   --help, -h  show help

   FILTERING

   --constraint '~1.0'  Version constraint. eg: '~1.0'. See github.com/Masterminds/semver
   --limit 1            Limit to a number of results. 1 means only last and `-1` only first. (default: 0)
   --reverse            New versions first (default: false)
   --sort               Sorted by version instead of using backend ordering (default: true)

   FORMAT

   --json  Output JSON array of versions (default: false)
   --text  Output text table of versions (default: true)

   NIX VERSIONS BACKEND

   --channel value  Nixpkgs channel for lazamar backend. Enables lazamar when set. (default: "nixpkgs-unstable")
   --lazamar        Use https://lazamar.co.uk/nix-versions as backend (default: true)
   --nixhub         Use https://www.nixhub.io/ as backend (default: false)
```
