This test simulates the following scenario:

1. Entity is provisioned with some attributes not set (in this case, a group
   without GID and a user without comment or shell).
2. The GID is chosen by `groupadd`, the shell is chosen by `useradd`, and then
   some other program (outside of `holo apply`) sets the comment of the user.
3. Another entity definition is introduced that requires the GID and the user
   comment and shell to be set to some different value.

The next `holo apply` should complain about being able to apply these changes
with `--force` only.

Also, a group (`tty`) is added to a pre-existing user (`root`), and we check
that the base image is initialized from the pre-existing state properly.
