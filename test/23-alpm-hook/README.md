This test checks the platform integration hooks Arch Linux.  It is
similar to the 20-archlinux test (which tests holo-files itself),
except that it should leave `etc/repofile-deleted-with-pacnew.conf`,
as we list the other files, but not it on stdin to the hook script.

[Reference](https://www.archlinux.org/pacman/alpm-hooks.5.html)
