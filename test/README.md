Test suite for `holo`
=====================

The `holo` binary is tested by running its subcommands on fabricated chroot
directories and comparing the output and resulting filesystem with the expected
results.

The test cases are run with `holo-test`, which is documented in its manpage at
`doc/holo-test.7.pod`. To run all test cases, do the following from the
repository root:

    ./util/holo-test holo ./test/??-*
