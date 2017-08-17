This test checks the platform integration for RPM-based distributions.

* `/etc/targetfile-with-rpmnew.conf` has a config file and repo file with an
  existing target base, and there is also a `.rpmnew` file that the package manager
  has placed next to the config file as part of an update of the application
  package. We should recognize this file and move it into `/var/lib/holo/files/base`.
* `/etc/targetfile-with-rpmsave.conf` is the same basic situation, but instead
  of saving the new default config in `$TARGET_PATH.rpmnew`, RPM decided to
  overwrite the configuration file directly, and save a backup of the previous
  configuration at `$TARGET_PATH.rpmsave`. (It does that sometimes, apparently.)
* `/etc/repofile-deleted-with-rpmnew.conf` has a config file whose repo file
  was deleted. But during the same package manager run that deleted the repo
  file, the application was updated and a `.rpmnew` file was placed next to the
  target file. This test was added after I found a bug in this situation: The
  `.rpmnew` file would not be picked up during scrubbing.
* `/etc/repofile-deleted-with-rpmsave.conf` is the same, but with an `.rpmsave`
  file instead of an `.rpmnew` file.

[Reference 1](https://ask.fedoraproject.org/en/question/25722/what-are-rpmnew-files/)
[Reference 2](http://www.rpm.org/max-rpm/ch-rpm-upgrade.html)
