# Notes for packagers

## Dependencies

Run-time dependencies for this repo:

* `shadow` (the package that provides the `{user,group}{add,mod,del}` tools)
* `HOLO_API_VERSION=2` (see [holo/PACKAGING.md](https://github.com/holocm/holo/blob/master/PACKAGING.md) for details)

Build-time dependencies for this repo:

* `go`
* `perl` (only required if you run `make check`)

## Install/update/remove scripts

To register the plugin with Holo, include the following install/update/remove scripts, as described in [holo-plugin-interface(7)](https://github.com/holocm/holo/blob/master/doc/holo-plugin-interface.7.pod):

```bash
post_install() {
    holo apply file:/etc/holorc
}
post_update() {
    holo apply file:/etc/holorc
}
post_remove() {
    mkdir /usr/share/holo/users-groups
    holo apply file:/etc/holorc
    rmdir /usr/share/holo/users-groups
}
```
