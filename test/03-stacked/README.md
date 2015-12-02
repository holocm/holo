This testcase checks stacked application, that is: the correct application of
multiple repo files to a single config file. The following test files all have
plain stock config files:

* `/etc/plain-and-plain.conf` has two plain repo files. The second repo file
  will be installed, and the first one will be discarded.
* `/etc/plain-and-script.conf` has a plain repo file and a repo script that
  modifies the previous repo file.
* `/etc/script-and-script.conf` has the stock config file passed through two
  repo scripts.
* `/etc/link-and-script.conf` has a repo symlink, which is then resolved into a
  content buffer and passed through a repo script. This especially checks if,
  during symlink resolution, relative symlinks in the repo are correctly
  resolved against the target location.

Furthermore, `/etc/link-through-scripts.conf` is the same basic setup as
`/etc/script-and-script.conf`, but the stock config file is a symlink.

As a bonus round, `/etc/check-ordering.conf` has two repository entries, a
plain file and a holoscript both within the same repo subdirectory, to check
that the repo entries are applied in the correct order (first the plain file,
then the holoscript).
