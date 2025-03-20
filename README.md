# nix-versions - Search nix packages versions.

This CLI utility helps you find the versions of nix packages that were available in a particular nixpkgs revision.

It uses the following backends for searching nixpkgs revisions:

- [lazamar/nix-versions](https://lazamar.co.uk/nix-versions/)
- [nixhub](https://nixhub.io)

It can search for packages by attribute-path, or by fuzzy searching [search.nixos.org](https://search.nixos.org) by name/description or by the executable programs they provide.
This is possible thanks to [nix-search-cli](https://github.com/peterldowns/nix-search-cli)'s ElasticSearch client.

<details>
<summary>   

### Installation

</summary>

Install with nix

```shell
> nix profile install github:vic/nix-versions
> nix-versions --help
```

Or use directly from github

```shell
> nix run github:vic/nix-versions -- --help
```

</details>

<details>
<summary>

##### Usage Examples

</summary>

```shell
# Show known versions of emacs on Lazamar-index (including emacs-nox, emacs-gtk, etc)
> nix-versions --lazamar emacs


# don't include packages with other attribute names like, emacs-nox, emacs-gtk, etc.
> nix-versions --lazamar --exact emacs


# Latest versions of packages providing `pwd`.
# --exact means that packages must provide an executable named exactly `pwd`.
> nix-versions --exact bin/pwd@latest


# Latest versions of packages providing some executable programs.
#   packages providing `emacsclient`. (eg. emacs-nox, emacs-gtk, emacs)
#   packages providing `pip`. (eg. python312Packages.pip, python313Packages.pip)
> nix-versions --exact bin/pip@latest bin/emacsclient@'>27 <29 latest'
Version  Attribute              Nixpkgs-Revision
24.0     python313Packages.pip  2d068ae5c6516b2d04562de50a58c682540de9bf
24.0     python312Packages.pip  2d068ae5c6516b2d04562de50a58c682540de9bf
28.2     emacs-nox              09ec6a0881e1a36c29d67497693a67a16f4da573
28.2     emacs-gtk              09ec6a0881e1a36c29d67497693a67a16f4da573
28.2     emacs                  09ec6a0881e1a36c29d67497693a67a16f4da573


# Any package having an executable program that contains `rust` on its name
> nix-versions --exact=false bin/rust@latest
Version  Attribute                        Nixpkgs-Revision
0.1.1    rustycli                         2d068ae5c6516b2d04562de50a58c682540de9bf
0.5.0    rusty-man                        2d068ae5c6516b2d04562de50a58c682540de9bf
0.5.7    rusty-psn-gui                    5d9b5431f967007b3952c057fc92af49a4c5f3b2
0.5.7    rusty-psn                        5d9b5431f967007b3952c057fc92af49a4c5f3b2
0.16.0   rustypaste                       2d068ae5c6516b2d04562de50a58c682540de9bf
0.24.0   rustywind                        8f76cf16b17c51ae0cc8e55488069593f6dab645
1.1.3    rustus                           8f76cf16b17c51ae0cc8e55488069593f6dab645
1.7.3    rustup-toolchain-install-master  2d068ae5c6516b2d04562de50a58c682540de9bf
1.27.1   rustup                           2d068ae5c6516b2d04562de50a58c682540de9bf
2.4.1    rustscan                         b58e19b11fe72175fd7a9e014a4786a91e99da5f


# Packages matching the `netscape` query on search.nixos.org
> nix-versions '~netscape'@latest
Version  Attribute         Nixpkgs-Revision
0.1.3    netsurf.libnslog  0d534853a55b5d02a4ababa1d71921ce8f0aee4c
0.1.6    netsurf.libnspsl  2d068ae5c6516b2d04562de50a58c682540de9bf
0.2.2    netsurf.libnsfb   2d068ae5c6516b2d04562de50a58c682540de9bf
0.4      netselect         6c5c5f5100281f8f4ff23f13edd17d645178c87c
0.4.2    netsurf.libdom    0d534853a55b5d02a4ababa1d71921ce8f0aee4c
0.6.2    netscanner        0d534853a55b5d02a4ababa1d71921ce8f0aee4c
0.6.9    netsniff-ng       e05f8bda630a0836d777d84de14b3c16eb758514
0.9.2    netsurf.libcss    0d534853a55b5d02a4ababa1d71921ce8f0aee4c
1.0.0    netsurf.libnsgif  de0fe301211c267807afd11b12613f5511ff7433
3.11     netsurf.browser   2d068ae5c6516b2d04562de50a58c682540de9bf


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

</details>

<details>
<summary>

###### `nix-versions --help`

</summary>

```
nix-versions - show available nix packages versions

USAGE:
   nix-versions [options] PKG_ATTRIBUTE_NAME...

PKG_ATTRIBUTE_NAME:
   A package attribute name like `emacs` or `python312Packages.pip`.
   Use https://search.nixos.org to find the attribute name for a package.

   If you don't know the attribute name, you can search for packages
   by query (prefixed by `~`) or by program name (prefixed by `bin/`).

   For example `~ cursor editor` will search the index at search.nixos.org for
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

</details>

<details>
<summary>

###### Motivation

</summary>

- `nixpkgs` is an outstanding repository of programs, some say it's the largest most up-to-date repository. However since nixpkgs is only a repo of receipes, it will likely only contain the most recent version of a package. That's why sites like lazamar's and nixhub help searching for historic revisions of nixpkgs that used to contain a particular program version.

- I'm trying to use this CLI app to help other utilities find previous versions of nixpkgs programs.

</details>
