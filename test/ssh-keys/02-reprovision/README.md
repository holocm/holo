This test is identical to `test/01-provision`, but the `.ssh/authorized_keys`
of the test users are those generated during `test/01-provision`. We therefore
check that the next `holo apply` does not change anything.
