# Notes for packagers

## Dependencies

Run-time dependencies for this repo:

* `openssh`
* `holo (>=1.2, <2.0)`

Build-time dependencies for this repo:

* `go`
* `perl` (only required if you run `make check`)

## Install/update/remove scripts

Older versions of holo-ssh-keys required that the `/etc/holorc` be modified
with Holo through the use of install/update/remove scripts in the package. This
is not necessary anymore.
