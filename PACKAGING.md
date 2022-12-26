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

## Configuration files

`/etc/holorc` and `/etc/holorc.d/*` should be marked as configuration files.

## Caveats

Holo includes several plugins for itself. Since
[holo-build](https://github.com/holocm/holo-build) will also generate
dependencies on these, the `holo` package must have a Provides relation (or
whatever this is called in your package format) to `holo-files`,
`holo-generators`, `holo-run-scripts`, `holo-ssh-keys` and `holo-users-groups`.

Since the results of `golangci-lint` are known to be unstable over time, esp.
between different Go versions, packagers are advised to set the environment
variable `SKIP_STATIC_CHECK=true` when running `make check`. This setting will
skip the potentially unstable checks while still running the vast majority of
useful tests.
