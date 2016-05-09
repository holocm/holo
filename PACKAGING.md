# Notes for packagers

## Dependencies

Run-time dependencies for this repo:

* `HOLO_API_VERSION=3`, satifsied by `holo>=1.1` (see [holo/PACKAGING.md](https://github.com/holocm/holo/blob/master/PACKAGING.md) for details)

Build-time dependencies for this repo:

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
    mkdir /usr/share/holo/run-scripts
    holo apply file:/etc/holorc
    rmdir /usr/share/holo/run-scripts
}
```
