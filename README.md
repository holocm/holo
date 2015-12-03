# holo-users-groups - provision user accounts and groups

[![Build Status](https://travis-ci.org/holocm/holo-users-groups.svg?branch=master)](https://travis-ci.org/holocm/holo-users-groups)

This Holo plugin can be used to create and modify UNIX user accounts and
groups, as stored in `/etc/passwd` and `/etc/group`.

## Installation

It is recommended to install `holo-users-groups` as a package. The
[website](http://holocm.org) lists distributions that have a package.

Holo requires [Go](https://golang.org) and [Perl](https://perl.org) as
build-time dependencies. Also, [shadow](https://pkg-shadow.alioth.debian.org/)
is needed as a runtime dependency. (Check `which useradd` to see if you have
this installed already.) Once you're all set, the build is done with

```
git submodule update --init --recursive     # if you cloned the Git repo
make
make check
sudo make install
```

## Documentation

User documentation is currently missing, but will be added soon. (There is some
documentation in the manpage of Holo itself, though.)

For further information, visit [holocm.org](http://holocm.org).
