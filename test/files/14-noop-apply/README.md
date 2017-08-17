This testcase checks that nothing happens (and no `--force` is required) when
Holo intends to change the contents of the target file, but finds the file to
have been updated to that state by someone else already. In this test, the
update is triggered by a .pacnew file.
