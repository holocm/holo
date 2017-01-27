# Notes for packagers

I'm actively optimizing for developer and packager experience, so packaging of
Holo and its plugins should be straight-forward.

## Dependencies

Run-time dependencies for this repo:

* `git diff` is used for file diffs. This was "strongly suggested" before, but
  is now strictly required.

Build-time dependencies for this repo:

* `go`
* `perl` (only required if you run `make check`)

## Configuration files

`/etc/holorc` and `/etc/holorc.d/*` should be marked as configuration files.

## Caveats

### The right package names

Please set the package names identical to the repo names: `holo` for this repo,
and `holo-foo-bar` for plugins (e.g. `holo-users-groups` or `holo-run-scripts`).
This is important because [holo-build](https://github.com/holocm/holo-build)
will autogenerate depedencies on these packages when appropriate.

### Holo includes holo-files

Holo includes the `holo-files` plugin. Since `holo-build` will also generate
dependencies on `holo-files`, the `holo` package must have a Provides relation
(or whatever this is called in your package format) to `holo-files`.

### Verify Holo API version

Holo uses a custom protocol to communicate with its plugins. This protocol
carries a version number, and the `holo` package should announce this version
number as a Provides relation (or whatever this is called in your package
format). Similarly, each plugin package should depend on the right API version.

Changes to the API version will be noted in the release notes. If you don't
know which API version is current, look for `API_VERSION` in the source
code. For example:

    $ ack API_VERSION ./src/
    src/holo/impl/plugin.go
    101:    env = append(env, "HOLO_API_VERSION=3")

    src/holo-files/main.go
    43:             fmt.Fprintln(os.Stdout, "MIN_API_VERSION=3")
    44:             fmt.Fprintln(os.Stdout, "MAX_API_VERSION=3")
