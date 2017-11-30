This test checks the platform integration for Alpine Linux.

* `/etc/targetfile-with-apknew.conf` has a config file and repo file with an
  existing target base, and there is also a `.apk-new` file that the package manager
  has placed next to the config file as part of an update of the application
  package. We should recognize this file and move it into `/var/lib/holo/files/base`.
* `/etc/repofile-deleted-with-apknew.conf` has a config file whose repo file
  was deleted. But during the same package manager run that deleted the repo
  file, the application was updated and a `.apk-new` file was placed next to the
  target file. This test was added after I found a bug in this situation: The
  `.apk-new` file would not be picked up during scrubbing.
