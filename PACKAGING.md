# Notes for packagers

## Dependencies

Run-time dependencies for this repo:

* `holo (>=1.2, <2.0)`

Build-time dependencies for this repo:

* `perl` (only required if you run `make check`)

## Install/update/remove scripts

Older versions of holo-users-groups required that the `/etc/holorc` be modified
with Holo through the use of install/update/remove scripts in the package. This
is not necessary anymore.
