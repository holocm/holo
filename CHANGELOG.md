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

