# Notes for packagers

I'm actively optimizing for developer and packager experience, so packaging of
Holo and its plugins should be straight-forward.

## Dependencies

Run-time dependencies for this repo:

* `git` (specifically, the `git diff` subcommand)
* `openssh` (specifically, the `ssh-keygen` tool)
* `shadow` (the package that provides the `{user,group}{add,mod,del}` tools)

Build-time dependencies for this repo:

* `go`
* `perl` (for `make check`, and compiling the manpages)

The test suite requires at least version 6.8 of openssh, but this
version requirement is is not nescessary for actual use.

## Configuration files

`/etc/holorc` and `/etc/holorc.d/*` should be marked as configuration files.

## Caveats

Holo includes several plugins for itself. Since
[holo-build](https://github.com/holocm/holo-build) will also generate
dependencies on these, the `holo` package must have a Provides relation (or
whatever this is called in your package format) to `holo-files`,
`holo-run-scripts`, `holo-ssh-keys` and `holo-users-groups`.
