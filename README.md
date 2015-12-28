# holo-run-scripts - run custom scripts during `holo apply`

[![Build Status](https://travis-ci.org/holocm/holo-run-scripts.svg?branch=master)](https://travis-ci.org/holocm/holo-run-scripts)

This Holo plugin can be used to run custom scripts during `holo apply`, usually
at the end when all other entities have been provisioned. A common usecase would
be to reload services after having provisioned their configuration files.

## Installation

It is recommended to install `holo-run-scripts` as a package. The
[website](http://holocm.org) lists distributions that have a package.

To build, install [Holo](https://github.com/holocm/holo) first, then run

```
make check
sudo make install
```

## Documentation

User documentation is available in [man page form](doc/holo-run-scripts.8.pod).

For further information, visit [holocm.org](http://holocm.org).
