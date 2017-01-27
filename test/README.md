Test suite for `holo`
=====================

The `holo` binary is tested by running its subcommands on fabricated chroot
directories and comparing the output and resulting filesystem with the expected
results.

The test cases are run with `holo-test`, which is documented in its manpage at
`doc/holo-test.7.pod`. To run all test cases, do the following from the
repository root:

    ./util/holo-test holo ./test/??-*

There are several test-hooks built in to the binaries:

 - `holo-files` also supports the `unittest` os-release ID, which disables
   package-manager specific functionality.

 - The environment variable `$HOLO_TEST_FLAGS` can be used to send flags to the
   testing subsystem (that is, the "testing" Go package). Documentation on
   supported test flags can be accessed by running `HOLO_TEST_FLAGS=-help
   ./bin/holo.test`.

 - The environment variable `$HOLO_TEST_COVERDIR` can be used to augment
   `$HOLO_TEST_FLAGS`. At process start up, the program will insert
   `-test.coverprofile $HOLO_TEST_COVERDIR/$UNIQ_FILE` into the flags passed to
   the testing subsystem, where `$UNIQ_FILE` is a identifier unique to that
   process invocation.
