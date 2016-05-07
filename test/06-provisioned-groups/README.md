This test is mostly identical to `01-groups`, but all entities have been
provisioned already. The test confirms the behavior of `holo diff` in the
presence of provisioned images.

`group:wronggid` still has the wrong GID, but this time the base image is empty
and the provisioned image matches the definition, thus the change has been made
by a user.
