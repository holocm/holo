This testcase checks that when Holo is
 1. first provisioning a file that Holo has never touched before, and
 2. a system-update has already left a new stock version (a `.pacnew`
    file in this test) because of user modification,
that Holo correctly recognizes
 1. that the UpdatedTargetBase is the base version, and
 2. that `--force` should be required to wipe out the user
    modifications.
