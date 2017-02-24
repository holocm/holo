# Holo - minimalistic config management

[![Build Status](https://travis-ci.org/holocm/holo.svg?branch=master)](https://travis-ci.org/holocm/holo)
[![Coverage Status](https://coveralls.io/repos/github/holocm/holo/badge.svg?branch=master)](https://coveralls.io/github/holocm/holo?branch=master)

Holo is a radically simple configuration management tool that relies as much as
possible on package management for the whole system setup and maintenance
process. This is achieved by using metapackages to define personal package
selections for all systems or for certain types of systems.

Holo has a plugin interface to extend its capabilities. It comes with the core
plugin `holo-files` to provision configuration files. Here are some other
plugins that you may find useful:

* [holo-users-groups](https://github.com/holocm/holo-users-groups) creates user
  accounts and groups.
* [holo-run-scripts](https://github.com/holocm/holo-run-scripts) invokes custom
  scripts during the provisioning phase.
* [holo-ssh-keys](https://github.com/holocm/holo-ssh-keys) provisions
  `.ssh/authorized_keys`.

<small>If you've written a new plugin, add it to this list by editing this file
and sending a pull request.</small>

## Installation

It is recommended to install Holo as a package. The
[website](http://holocm.org) lists distributions that have a package.

Holo requires [Go](https://golang.org) and [Perl](https://perl.org) as
build-time dependencies. There are no runtime dependencies other than a libc.
Once you're all set, the build is done with

```
make
make check
sudo make install
```

## Documentation

User documentation is available in man page form:

* [holo(8)](doc/holo.8.pod)
* [holo-files(8)](doc/holo-files.8.pod)
* [holorc(5)](doc/holorc.5.pod)
* [holo-test(7)](doc/holo-test.7.pod)
* [holo-plugin-interface(7)](doc/holo-plugin-interface.7.pod)

For further information, visit [holocm.org](http://holocm.org).
