# v2.2 (TBD)

Changes:

- User definitions now have a new flag `SkipBaseGroups` which, if set, causes `holo-build` to ignore existing auxiliary
  groups that this user is in before the first `holo apply` of this user.
- Fix a bug where user definitions previously behaved as if `SkipBaseGroups = true`, even if the intention always was to
  behave like `SkipBaseGroups = false` (which is the default).
- Fix a bug where `user:root` could not be provisioned to because the user ID 0 was misinterpreted as "no UID".

# v2.1 (2017-11-30)

Changes:

- `holo-files` now recognizes Alpine Linux and derivatives and handles `.apk-new` files correctly on these
  distributions.

# v2.0.3 (2017-10-22)

This contains the actual fix for the bugfix that was supposed to be fixed in the previous release.

# v2.0.2 (2017-10-21)

Bugfixes:

- Fix a bug where most invocations of `holo` would create empty directories called `base` and `provisioned` in the
  current working directory.

# v2.0.1 (2017-10-05)

Bugfixes:

- Fix a bug where, when applying a symlink over a file, `holo apply` would not stay silent if the symlink has already
  been provisioned.

# v2.0 (2017-08-18)

Backwards-incompatible changes:

- `holo-test` has been removed from the public interface. Plugins that wish to use it are advised to vendor it from this
  repo into their own.

Packagers beware:

- Add `Provides` and `Replaces` package relations from this package to `holo-run-scripts`, `holo-ssh-keys` and
  `holo-users-groups` (these packages are now included in this one).
- New runtime dependencies: `shadow` (inherited from `holo-users-groups`) and `openssh` (inherited from
  `holo-ssh-keys`). See [PACKAGING.md](./PACKAGING.md) for details.

Changes:

- `holo`, `holo-files`, `holo-ssh-keys` and `holo-users-groups` have been merged into a single binary, thus massively
  reducing total installation size.
- Fix a bug in `holo-test` where tests could fail because of randomized names of temporary directories.
- Fix a bug in `holo-users-groups` where scrubbing of a user definition fails
  when the definition only adds an auxiliary group to a pre-existing user with no other auxiliary groups.
- Install the ALPM hook in the standard location.
- When Holo is installed via `go get`, show the version string "unknown" instead of an empty string.

# v1.3.1 (2017-03-19)

Bugfixes:

- Fix a bug where, on Arch Linux, the post-installation hook could get confused when Holo sorted entity names differently than sort(1) did.

# v1.3 (2017-03-18)

Special thanks to new contributor @LukeShu who did a lot of the hard work that went into this release, both in terms of
new features, boring refactoring work and documentation proof-reading.

Changes:

- `holo-files` now allows for fast-forwarding: When the computed content of a target file changes, but that change has
  already been done by the user, `holo-files` will now skip writing the target file and just update
  `/var/lib/holo/files/provisioned` instead of complaining that the target file does not match the previously
  provisioned content. (#24)
- When invoking Holo, plugin IDs can be used as selectors. For example, `holo apply ssh-keys` will apply all entities
  from the `holo-ssh-keys` plugin.

Bugfixes:

- Bring the scrubbing logic in line with the applying logic:
  - When a resource file is deleted while the target base is updated, restore the updated target base instead of the old
    one. (#16)
  - When a resource file is deleted and the saved version (`.pacsave`, `.rpmsave`, `.dpkg-old`) has been changed by the
    user, do not delete it. (#29)
  - Scrubbing has become more resilient against filesystem errors. When some file cannot be cleaned up, it will report
    that and keep going as much as possible. This is useful because Holo will forget about the entity once it is
    scrubbed, so the user should be informed about which actions remain to properly clean up the target file.
  - On Arch Linux, `.pacsave.N` files are now handled properly, similar to the existing handling for `.pacsave` files.
- Make sure that the cache directory (usually at `/tmp/holo.$$/`) is cleaned up even when an operation fails. (#20)
- Fix `make clean` to run correctly when the source is extracted from a tarball rather than cloned from git.
- Various fixes to `make check` to avoid false negatives.

Miscellaneous:

- Various internal refactorings.
- The documentation was proof-read and clarified in various locations.
- The test suite now checks code coverage.
- There are some files in `debian/` which should make it pretty easy to make a Debian package for Holo if anyone is
  interested in submitting it to Debian, Ubuntu etc.
- Releases are now signed by GPG key `0xD6019A3E17CA2D96`.

# v1.2.1 (2016-05-25)

Bugfixes:

- Fix installed holorc snippet.

# v1.2 (2016-05-25)

New features:

- Support modularized configuration in `/etc/holorc.d/*`, mainly to simplify the installation process for plugins. (#15)
- Add `--porcelain` option to `holo scan`.
- On Arch Linux with pacman >= 5.0, install a post-installation hook to handle .pacnew files automatically.

Miscellaneous:

- Strip binaries during build. (#14)

# v1.1.1 (2016-04-11)

Changes:

- Don't acquire the `/run/holo.pid` lockfile for readonly operations. This esp. unbreaks non-privileged usage of Holo e.g. for shell autocompletions because `/run` is writable only by root. (#13)
- Missing plugins are not a fatal error anymore. This fixes a logic deadlock during plugin uninstallation.

# v1.1 (2016-04-09)

Backward-incompatible changes:

- Packagers beware: The plugin interface version increases from **2** to **3**. (#12)
- Plugin developers: Update your plugins to
  - understand the new `info` command,
  - use the new semantics of the `diff` command (which now shall report files for diffing, instead of computing the diff itself),
  - and replace custom error messages with the new `requires --force to (update|restore)` messages where appropriate.

New features:

- Most output is now colorized appropriately, especially diffs from `holo diff`.
- `holo apply` without `--force` will now show a diff when the entity has manual changes that only `holo apply --force` will overwrite.
- Add a lockfile (`/run/holo.pid`) to prevent multiple parallel runs. (#9)
- `holo-test` now generates `colored-*-output` artifacts to allow plugin developers to inspect the colorized output of Holo. These artifacts are not validated against `expected-*-output`.

# v1.0.1 (2015-12-28)

Changes:

- Fix an edge case in `holo-files` which could cause source files to be applied in a different order than that reported by `holo scan`.
- Fix several glitches and inconsistences in the manpages. (Shout-out to @S1FeHa for proof-reading.)

# v1.0 (2015-12-18)

New features since Beta 2:

- Entities can now be identified by their source files. (#4)

Further changes since Beta 2:

- The name format for file entities has changed, from e.g. `/etc/foo.conf` to `file:/etc/foo.conf`.
- The manpages have been updated to describe the new plugin system.
- Fix a bug which caused unchanged target files to be reported during `holo apply --force`.

Plugin interface changes:

- The `HOLO_ROOT_DIR` variable is now always set, by default to `/`.
- The new `SOURCE:` directive can be used to link entities to their source files.
- The plugin interface version has increased to `HOLO_API_VERSION=2`.

# v1.0-beta.2 (2015-12-04)

Bugfixes:

- install `holorc` in the right path
- install `holo-test` to the right `$PATH`

# v1.0-beta.1 (2015-12-03)

Holo has been refactored into a plugin-based structure. The capabilities for [provisioning user accounts and groups][ug]
and [running custom provisioning scripts][rs] must now be installed separately.

**Backwards-incompatible changes:** A lot of filesystem paths change to follow the new plugin-based structure.

```
/usr/share/holo/{repo => files}/
/usr/share/holo/{provision => run-scripts}/
/usr/share/holo/{ => users-groups}/*.toml
/var/lib/holo/{ => files}/base/
/var/lib/holo/{ => files}/provisioned/
```

When updating, update all your configuration packages at the same time (to move stuff below `/usr/share/holo` into the
new locations), and take a backup of `/var/lib/holo` as the target bases will _definitely_ be messed up during the
update. Recipe:

```
cd /var/lib/holo
tar cf backup.tar base provisioned
[update Holo, install required plugins, update configuration packages for new paths]
cd /var/lib/holo/files
tar xf ../backup.tar
holo apply --force
cd /var/lib/holo
rm -r backup.tar base provisioned
```

Further changes:

- Optimize application algorithm: When the effect of the holoscript is overridden by a later repository entry that is a
  plain file, the holoscript is skipped entirely.

Known issues with this release:

- `make install` will put the holorc into the wrong place (`/etc/holo/holorc` instead of `/etc/holorc`).

This is the first release with the new split repository layout. Previous releases can be found [in the attic][ar].

[ug]: https://github.com/holocm/holo-users-groups
[rs]: https://github.com/holocm/holo-run-scripts
[ar]: https://github.com/holocm/holo-attic/releases

------------------------------------------------------------------------------------------------------------------------

# Changelog for holo-run-scripts before merge in Holo 2.0

The history of the `holo-run-scripts` repo can be found below the second parent of the merge commit 515e921429ff7ef35309da2eccbaea5df78d3222.
The following tags were in that repo at the time of merging:

    5e233dd51ffa11b17dbea2ae9d81b54feefcf77f v1.0
    29b4c9ce86a98bf2b266ad989a2c69ed1ed94ed5 v1.0-beta.1
    f5842d1e78d12bc261488d516e05bd8fcb2c0364 v1.1
    f28b1b52e2f09a8c444db1329e010679f9aa0793 v1.2
    c67e9b2b72a5b137e762b688ec3cd595270113da v1.3
    1f6c0d054c1913233fbe32884fafc619c395d73b v1.3.1

## v1.3.1 (2017-03-22)

Changes:

- Fix tests when run against Holo 1.3.

## v1.3 (2016-05-25)

Changes:

- Create `/usr/share/holo/run-scripts` during `make install`.
- Install holorc snippet instead of holoscript. (This change requires `holo >= 1.2`.)

## v1.2 (2016-04-10)

Changes:

- This release is compatible with version 3 of the Holo plugin interface, as used by Holo 1.1 and beyond.

## v1.1 (2015-12-18)

Changes:

- This release is compatible with version 3 of the Holo plugin interface, as used by Holo 1.0 and beyond.

## v1.0 (2015-12-04)

Changes:

- Find `holo-test` in `$PATH`.

## v1.0-beta.1 (2015-12-04)

Changes:

- This functionality is now offered as a Holo plugin for separate installation and packaging.

Known issues:

- If `make check` fails with "command not found: holo-test", edit the Makefile and replace `holo-test` with `/usr/lib/holo/holo-test`.

This is the first release with the new split repository layout. Previous releases can be found [in the attic](https://github.com/holocm/holo-attic/releases).

------------------------------------------------------------------------------------------------------------------------

# Changelog for holo-ssh-keys before merge in Holo 2.0

The history of the `holo-ssh-keys` repo can be found below the second parent of the merge commit e1e3d2e3d3826ddb2971f1e78b20b4dd467f3e28.
The following tags were in that repo at the time of merging:

    a62c7288c464cbe359d9d0a17bc4c9b8556e3461 v1.0
    b4a2dc668e0a5caa3d80dc5b628acb814926330f v1.1
    453ac3fb6699e877584ba6e3027a6c88d206765f v1.2
    e6e25242795166cbcb18eedbbc7bb58250122027 v1.2.1

## v1.2.1 (2017-03-22)

Changes:

- Fix tests under Holo 1.3.

## v1.2 (2016-05-25)

Changes:

- Create `/usr/share/holo/ssh-keys` during `make install`.
- Install holorc snippet instead of holoscript. (This change requires `holo >= 1.2`.)
- Strip binaries during build. (holocm/holo#14)

## v1.1 (2016-04-10)

Changes:

- This release is compatible with version 3 of the Holo plugin interface, as used by Holo 1.1 and beyond.

## v1.0 (2015-12-30)

Initial release.

------------------------------------------------------------------------------------------------------------------------

# Changelog for holo-users-groups before merge in Holo 2.0

The history of the `holo-users-groups` repo can be found below the second parent of the merge commit 2d6e87e41d62abced6f5c08428e7bea523cfb5a4.
The following tags were in that repo at the time of merging:

    aced8d55b4dae5ef84eb3a4fde87240176498638 v1.0-beta.1
    c843bb2a66160d5c6371ec11af968985c6742d33 v1.0
    881c382776c48e543885662da74ea70a4f878793 v1.1
    be71e6b0415a3071f415429b597297c8bb3ab153 v1.2
    95f3792d73c7a82c923a43ebb81e537bde973618 v1.3
    66f46701aafbaa65397c2e4d68f8773f70272018 v2.0
    722e3be714ce5371e808e21d4af97c7dd8cdf9bd v2.0.1
    88f9154e0034ffff01fc686bcfd6ea970a676f2a v2.1
    f19426859936d15e43e6da5a3b28f11acd5f95e8 v2.1.1

## v2.1.1 (2017-03-22)

Changes:

- Fix tests under Holo 1.3.

## v2.1 (2016-05-25)

Bugfixes:

- When an entity is deleted by another program, restore it correctly during force apply. (#3)

Miscellaneous:

- Strip binaries during build. (holocm/holo#14)
- Install holorc snippet instead of holoscript. (This change requires `holo >= 1.2`.)

## v2.0.1 (2016-05-09)

Bugfixes:

- Fix a bug that causes `usermod` to be called without arguments when user entities are underspecified.

## v2.0 (2016-05-09)

Backwards-incompatible changes:

- Previous versions only tracked which users/groups were touched by holo-users-groups. v2.0 will instead record more
  information (_base images_ and _provisioned images_) to make more competent decisions during `holo apply` and thus
  require less user interaction. The introduction of base images requires manual intervention in some cases. Please
  refer to the migration guide below for details.

Misc. changes:

- Create `/usr/share/holo/users-groups` during `make install`.

### Migration guide

When upgrading to v2.0, manual intervention is required in some cases. The following documentation outlines what you need to do.

#### The old state format

Versions 1.2 and 1.3 tracked which users and groups have been provisioned by holo-users-groups, by writing a single file like this:

```
$ cat /var/lib/holo/users-groups/state.toml
ProvisionedGroups = ["mygroup"]
ProvisionedUsers = ["myuser", "myotheruser"]
```

#### The new format

Version 2.0 abandons this format and uses **base images** and **provisioned images** instead. Provisioned images are created automatically after every successful `holo apply` and don't require manual intervention during the upgrade. The bigger change is the base image, which records the state of the user or group just before the first provisioning. This change allows holo-users-groups to make more competent decisions when cleaning up an entity that is not needed anymore.

After migrating to version 2.0, the first invocation of holo-users-groups will remove the state.toml and create empty base images for all the mentioned entities:

```
$ sudo holo apply
...
$ cat /var/lib/holo/users-groups/base/group:mygroup.toml
[[group]]
name = "mygroup"
$ cat /var/lib/holo/users-groups/base/user:myuser.toml
[[user]]
name = "myuser"
$ cat /var/lib/holo/users-groups/base/user:myotheruser.toml
[[user]]
name = "myotheruser"
```

#### What you need to do

These base images are *empty* because they only contain the `name` attribute. Empty base images are appropriate when the entity (the user or group) was **created** by holo-users-groups. If the entity did exist already, and holo-users-groups just modified some of its attributes, you need to amend the base image to describe the state of the entity before the first provisioning.

As an example, Arch Linux's default `/etc/passwd` contains the following `http` user:

```
$ getent passwd http
http:x:33:33:http:/srv/http:/bin/false
```

Now consider that we use an entity definition that sets a different home directory:

```
$ cat /usr/share/holo/users-groups/http-home.toml
[[user]]
name = "http"
home = "/srv/webspace"
```

In version 1.2/1.3 of holo-users-groups, this would change the home directory and record `ProvisionedUsers = ["http"]` in the `state.toml`. Upon migration to version 2.0, this would be converted to an empty base image:

```
$ cat /var/lib/holo/users-groups/base/user:http.toml
[[user]]
name = "http"
```

This implies that the `http` user was created by holo-users-groups, which is wrong! To fix this, you need to add all the entity's attributes to the base image as they were before the first provisioning (especially including the old home directory value!). The syntax is identical to entity definitions:

```
$ cat /var/lib/holo/users-groups/base/user:http.toml
[[user]]
name = "http"
uid = 33
group = "http"
home = "/srv/http"
shell = "/bin/false"
```

**Tip:** You can already create the non-empty base images *before* migrating to version 2.0. The upgrade process will recognize and use these. But be cautious: Non-empty base images must always include the actual UID or GID of the user or group.

## v1.3 (2016-04-10)

Changes:

- This release is compatible with version 3 of the Holo plugin interface, as used by Holo 1.1 and beyond.

## v1.2 (2015-12-20)

New features:

- Track users/groups that have been provisioned, and offer to delete them when the corresponding entity definitions are deleted. (#1)

## v1.1 (2015-12-18)

Changes:

- This release is compatible with version 2 of the Holo plugin interface, as used by Holo 1.0 and beyond.

## v1.0 (2015-12-04)

Bugfixes:

- Fix calculation of paths to `/etc/group` and `/etc/passwd`.
  Changes:
- Find `holo-test` in `$PATH`.

## v1.0-beta.1 (2015-12-03)

Changes:

- This functionality is now offered as a Holo plugin for separate installation and packaging.

Known issues:

- If `make check` fails with "command not found: holo-test", edit the Makefile and replace `holo-test` with `/usr/lib/holo/holo-test`.
- `holo apply` fails outside test scenarios because it tries to read `etc/passwd` and/or `etc/group` (without leading slash).

This is the first release with the new split repository layout. Previous releases can be found [in the attic](https://github.com/holocm/holo-attic/releases).
