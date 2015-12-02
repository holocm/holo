This test checks the various situations that `holo diff` can encounter when
diffing the current state of a target file to the last provisioned version,
stored at `/var/lib/holo/files/provisioned`.

* targets that have been added by the user: This case is covered by
  `test/01-basic` already.
* targets that have been modified by the user: `/etc/file-modified.conf` and
  `/etc/symlink-modified.conf`
* targets that have been deleted by the user: `/etc/file-deleted.conf` and
  `/etc/symlink-deleted.conf`
* targets whose types have changed by the user: `/etc/file-to-symlink.conf` and
  `/etc/symlink-to-file.conf`
* targets that have *not* been changed by the user: `/etc/file-unmodified.conf`
  and `/etc/symlink-unmodified.conf`
