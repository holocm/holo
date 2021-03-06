This test checks the platform integration for Arch Linux.

* `/etc/targetfile-deleted-with-pacsave.conf` has no config file and no repo
  files. So we assume that the application package and all holograms using that
  application have been uninstalled, and that the package manager saved the
  config file with a `.pacsave` suffix while uninstalling the application
  package. This file should be cleaned up, too.
* `/etc/targetfile-with-pacnew.conf` has a config file and repo file with an
  existing target base, and there is also a `.pacnew` file that the package manager
  has placed next to the config file as part of an update of the application
  package. We should recognize this file and move it into `/var/lib/holo/files/base`.
* `/etc/repofile-deleted-with-pacnew.conf` has a config file whose repo file
  was deleted. But during the same package manager run that deleted the repo
  file, the application was updated and a `.pacnew` file was placed next to the
  target file. This test was added after I found a bug in this situation: The
  `.pacnew` file would not be picked up during scrubbing.

[Reference](https://wiki.archlinux.org/index.php/Pacnew_and_Pacsave_files)
