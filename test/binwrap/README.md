We want to test holoscripts that are symlinks, so the tests have stuff like this:

```
$ grep uniq test/files/02-holoscripts/source-tree
symlink   0777 ./usr/share/holo/files/02-holoscripts/etc/link-through-link.conf.holoscript
/usr/bin/uniq
```

This works fine on systems where `uniq` is its own binary coming from GNU
coreutils. On systems using Busybox, however, `/usr/bin/uniq` is just another
symlink to `/bin/busybox`. Busybox then fails because it sees the holoscript's
filename as `argv[0]` and thus cannot infer which applet to use.

We therefore added the scripts in this directory as an additional indirection
that makes the symlink holoscripts work on Busybox systems.

```
$ cat test/binwrap/uniq
#!/bin/sh
exec uniq "$@"

$ grep uniq test/files/02-holoscripts/source-tree
symlink   0777 ./usr/share/holo/files/02-holoscripts/etc/link-through-link.conf.holoscript
../../../../../../../../../binwrap/uniq
```
