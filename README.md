# nix-versions - Search nix packages versions.

This CLI utility helps you find the versions of nix packages that were available in a particular nixpkgs revision.

It uses the following backends for searching nixpkgs revisions:

- [lazamar/nix-versions](https://lazamar.co.uk/nix-versions/)
- [nixhub](https://nixhub.io)

It can search for packages by name/description or by the executable programs they provide.
This is possible thanks to [nix-search-cli](https://github.com/peterldowns/nix-search-cli)'s ElasticSearch client for [search.nixos.org](https://search.nixos.org)

## Installation

Install with nix

```shell
> nix profile install github:vic/nix-versions
> nix-versions --help
```

Or use directly from github

```shell
> nix run github:vic/nix-versions -- --help
```

#### Examples

```shell
# List known versions of emacs
> nix-versions emacs

# Latest versions of packages providing some executable programs.
#   packages providing `emacsclient`. (eg. emacs-nox, emacs-gtk, emacs)
#   packages providing `pip`. (eg. python312Packages.pip, python313Packages.pip)
> nix-versions --exact bin/pip@latest bin/emacsclient@'>27 <29 latest'
Version  Attribute              Nixpkgs-Revision
24.0     python313Packages.pip  21808d22b1cda1898b71cf1a1beb524a97add2c4
24.0     python312Packages.pip  21808d22b1cda1898b71cf1a1beb524a97add2c4
28.1     emacs-nox              7cf5ccf1cdb2ba5f08f0ac29fc3d04b0b59a07e4
28.2     emacs-gtk              459104f841356362bfb9ce1c788c1d42846b2454
28.1     emacs                  7cf5ccf1cdb2ba5f08f0ac29fc3d04b0b59a07e4


# Any package having an executable program that contains `rust` on its name
> nix-versions --exact=false bin/rust@latest
Version     Attribute                        Nixpkgs-Revision
0.23.1      rustywind                        21808d22b1cda1898b71cf1a1beb524a97add2c4
0.16.0      rustypaste                       21808d22b1cda1898b71cf1a1beb524a97add2c4
0.1.1       rustycli                         21808d22b1cda1898b71cf1a1beb524a97add2c4
0.3.7       rusty-psn-gui                    21808d22b1cda1898b71cf1a1beb524a97add2c4
0.3.7       rusty-psn                        21808d22b1cda1898b71cf1a1beb524a97add2c4
0.5.0       rusty-man                        21808d22b1cda1898b71cf1a1beb524a97add2c4
1.0.0       rustus                           21808d22b1cda1898b71cf1a1beb524a97add2c4
1.7.3       rustup-toolchain-install-master  21808d22b1cda1898b71cf1a1beb524a97add2c4
2017-10-29  rustup                           28e0126876d688cf5fd15da1c73fbaba256574f0
2.3.0       rustscan                         21808d22b1cda1898b71cf1a1beb524a97add2c4

# Return only the most recent version
> nix-versions --limit 1 emacs

# Only versions between 25 and 27. Output JSON
# same as 'emacs@>= 25 <= 27'
> nix-versions --constraint '>= 25 <= 27' --json emacs

# Latest of 29 series.
# same as 'emacs@latest~29'
> nix-versions --constraint '~29' --limit 1 emacs

# Do not include emacs-nox and emacs-gtk
> nix-versions --exact emacs

# Show versions of pip from nixhub.io in the order that nixhub returns them
> nix-versions --nixhub --sort=false python312Packages.pip

# Use release channel `nixpkgs/nixos-24.05` (using lazamar search)
> nix-versions --channel nixos-24.05 python312Packages.pip

# NixHub.io has rate-limits but will likely have indexed more recent versions.
# https://www.jetify.com/docs/nixhub/#rate-limits
> nix-versions --nixhub bun@latest
1.2.5    bun        573c650e8a14b2faa0041645ab18aed7e60f0c9a

# https://lazamar.co.uk/nix-versions/ has no rate-limit, we scrap the webpage.
> nix-versions --lazamar bun@latest
1.1.43   bun        21808d22b1cda1898b71cf1a1beb524a97add2c4
```

#### nix-versions --help

```
nix-versions - show available nix packages versions

USAGE:
   nix-versions [options] PKG_ATTRIBUTE_NAME...

PKG_ATTRIBUTE_NAME:
   A package attribute name like `emacs` or `python312Packages.pip`.
   Use https://search.nixos.org to find the attribute name for a package.

   If you don't know the attribute name, you can search for packages
   by query (prefixed by `~`) or by program name (prefixed by `bin/`).

   For example `~ git porcelain` will search the index at search.nixos.org for
   packages that match the query.

   And using `bin/pip` will search for packages that provide that program.

   Optionally you can add a version constraint to the package name like
   `bin/emacs@^25.x` or `emacs@>= 25 <= 27` or `~ git porcelain @latest`.

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

   --exact              Only include results whose attribute is exactly PKG_ATTRIBUTE_NAME (default: false)
   --limit n            Limit to a number of results. 1 means only last and `-1` only first. (default: 0)
   --reverse            New versions first (default: false)
   --sort               Sorted by version instead of using backend ordering (default: true)

   FORMAT

   --json  Output JSON array of versions (default: false)
   --text  Output text table of versions (default: true)

   NIX VERSIONS BACKEND

   --channel value  Nixpkgs channel for lazamar backend. Enables lazamar when set. (default: "nixpkgs-unstable")
   --lazamar        Use https://lazamar.co.uk/nix-versions as backend (default: false)
   --nixhub         Use https://www.nixhub.io/ as backend (default: true)

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
