# Notes for packagers

I'm actively optimizing for developer and packager experience, so packaging of
Holo and its plugins should be straight-forward.

## Dependencies

Run-time dependencies:

* `git diff` is used by `holo-files` for file diffs. If you want to make this
  an optional dependency, please mark it as "strongly suggested" etc.

Build-time dependencies:

* `go`
* `perl` (only required if you run `make check`)

## Configuration files

`/etc/holorc` should be marked as a configuration file.

## Caveats

### The right package names

Please set the package names identical to the repo names: `holo` for this repo,
and `holo-foo-bar` for plugins (e.g. `holo-users-groups` or `holo-run-scripts`).
This is important because [holo-build](https://github.com/holocm/holo-build)
will autogenerate depedencies on these packages when appropriate.

### Holo includes holo-files

Holo includes the `holo-files` plugin. Since `holo-build` will also generate
dependencies on `holo-files`, the `holo` package must provide this package.

### Verify Holo API version

Holo uses a custom protocol to communicate with its plugins. This protocol
carries a version number, and the `holo` package should announce this version
number as a Provides relation (or whatever this is called in your package
format). Similarly, each plugin package should depend on the right API version.

Changes to the API version will be noted in the release notes. If you don't
know which API version is current, look for `HOLO_API_VERSION` in the source
code. For example:

    $ ack HOLO_API_VERSION src
    src/holo/impl/plugin.go
    92:     env = append(env, "HOLO_API_VERSION=2")

    src/holo-files/main.go
    40:     if version := os.Getenv("HOLO_API_VERSION"); version != "2" {
    41:             fmt.Fprintf(os.Stderr, "!! holo-files plugin called with unknown HOLO_API_VERSION %s\n", version)
