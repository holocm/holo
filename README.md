# Holo - minimalistic config management

[![Build Status](https://travis-ci.org/holocm/holo.svg?branch=master)](https://travis-ci.org/holocm/holo)
[![Coverage Status](https://coveralls.io/repos/github/holocm/holo/badge.svg?branch=master)](https://coveralls.io/github/holocm/holo?branch=master)

Holo is a radically simple configuration management tool that relies as much as
possible on package management for the whole system setup and maintenance
process. This is achieved by using metapackages to define personal package
selections for all systems or for certain types of systems.

Holo has a plugin interface to extend its capabilities. It comes with the following
core plugins:

* `holo-files` provisions configuration files.
* `holo-run-scripts` invokes custom scripts during the provisioning phase.
* `holo-ssh-keys` provisions `.ssh/authorized_keys`.
* `holo-users-groups` creates and modifies UNIX user accounts and groups, as
  stored in `/etc/passwd` and `/etc/group`.

<small>If you've written a new plugin, send me a link via the issue tracker and
I'll link to it here.</small>

## Installation

It is recommended to install Holo as a package. The
[website](http://holocm.org) lists distributions that have a package.

Holo requires [Go](https://golang.org) and [Perl](https://perl.org) as
build-time dependencies; and `git-diff` and [shadow](https://pkg-shadow.alioth.debian.org/)
as runtime dependencies. Once you're all set, the build is done with

```
make
make check
sudo make install
```

## Documentation

User documentation is available in man page form:

* [holo(8)](doc/holo.8.pod)
* [holo-files(8)](doc/holo-files.8.pod)
* [holo-run-scripts(8)](doc/holo-run-scripts.8.pod)
* [holo-ssh-keys(8)](doc/holo-ssh-keys.8.pod)
* [holo-users-groups(8)](doc/holo-users-groups.8.pod)
* [holorc(5)](doc/holorc.5.pod)
* [holo-plugin-interface(7)](doc/holo-plugin-interface.7.pod)
* [holo-test(7)](doc/holo-test.7.pod) (not a public interface)

For further information, visit [holocm.org](http://holocm.org).
