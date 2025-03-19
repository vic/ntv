# nix-versions - Search nix packages versions with lazamar/nix-versions or nixhub

This is a minimal CLI app written in Go that interfaces with [https://lazamar.co.uk/nix-versions/](https://lazamar.co.uk/nix-versions/) and [nixhub](https://nixhub.io).

## Installation

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

# Only versions between 25 and 27. Output JSON
nix-versions --constraint '>= 25 <= 27' --json emacs

# Latest of 29 series.
nix-versions --constraint '~29' --limit 1 emacs

# Also include emacs-nox and others
nix-versions --exact=false emacs

# Show versions of pip from nixhub.io in the order that nixhub returns them
nix-versions --nixhub --sort=false python312Packages.pip

# Use release channel `nixpkgs/nixos-24.05` (using lazamar search)
nix-versions --channel nixos-24.05 python312Packages.pip

# NixHub.io has rate-limits but will likely have indexed more recent versions.
# https://www.jetify.com/docs/nixhub/#rate-limits
nix-versions --nixhub bun@latest
1.2.5    bun        573c650e8a14b2faa0041645ab18aed7e60f0c9a

# https://lazamar.co.uk/nix-versions/ has no rate-limit, we scrap the webpage.
nix-versions --lazamar bun@latest
1.1.43   bun        21808d22b1cda1898b71cf1a1beb524a97add2c4
```

#### nix-versions --help

```
nix-versions - show available nix packages versions

USAGE:
   nix-versions [options] PKG_ATTRIBUTE_NAME...

OPTIONS:
   --help, -h  show help and exit
   --version   show version and exit

   FILTERING

   --constraint         Only include results that match a versions constraint. eg: '~1.0'.
                        See https://github.com/Masterminds/semver?tab=readme-ov-file#basic-comparisons

                        Constraint can also be part of PKG_ATTRIBUTE_NAME if it contains an `@` symbol.
                          'emacs@^25.x'        - Show all Emacs in the `25.x` series.
                          'emacs@>= 25 <= 27'  - Show all Emacs in the `25.x`-`27.x` series.
                          'emacs@latest'       - Only show the most recent emacs.
                          'emacs@latest<25'    - Only show the latest emacs before the `25` series
                          'emacs@latest~29'    - Only show the most recent emacs of the `29` series

   --exact              Only include results whose attribute is exactly PKG_ATTRIBUTE_NAME (default: true)
   --limit n            Limit to a number of results. 1 means only last and `-1` only first. (default: 0)
   --reverse            New versions first (default: false)
   --sort               Sorted by version instead of using backend ordering (default: true)

   FORMAT

   --json  Output JSON array of versions (default: false)
   --text  Output text table of versions (default: true)

   NIX VERSIONS BACKEND

   --channel value  Nixpkgs channel for lazamar backend. Enables lazamar when set. (default: "nixpkgs-unstable")
   --lazamar        Use https://lazamar.co.uk/nix-versions as backend (default: true)
   --nixhub         Use https://www.nixhub.io/ as backend (default: false)

Made with <3 by vic [https://x.com/oeiuwq].
See https://github.com/vic/nix-versions for examples and reporting issues.
```

- Use the `nix search` builtin command.

```shell
nix search emacs
```

- Official search https://search.nixos.org/packages

Actually `nix-search-cli` is an CLI interface to this.

Once you know the attribute path for the package you need, you can use `nix-versions` to search which nixpkgs revision corresponded to each particular package version.

## Motivation

- `nixpkgs` is an outstanding repository of programs, some say it's the largest most up-to-date repository. However since nixpkgs is only a repo of receipes, it will likely only contain the most recent version of a package. That's why sites like lazamar's and nixhub help searching for historic revisions of nixpkgs that used to contain a particular program version.

- I'm trying to use this CLI app to help other utilities find previous versions of nixpkgs programs.
