When a plain repository file is applied, the results of previous application
steps are discarded. This testcase checks that in that case, the previous
application steps are not even executed. It does so by using the fact that if a
script is not executed, any errors that it causes will not be seen.

    target/etc/foo.conf
        store at target/var/lib/holo/files/base/etc/foo.conf
        passthru target/usr/share/holo/files/01-first/etc/foo.conf.holoscript
           apply target/usr/share/holo/files/02-second/etc/foo.conf
        passthru target/usr/share/holo/files/03-third/etc/foo.conf.holoscript

The `01-first` holoscript will fail with error output and exit code 1, but this
is not seen because `02-second` is a plain file that discards the previous
application steps.

As a control group (to see the effect of an actually failing application),
`/etc/bar.conf` has the same failing holoscript, but without any other
repository entries.
