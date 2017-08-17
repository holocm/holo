This test checks the handling of group entities.

* `group:new` is not yet in the group database and will be created.
* `group:existing` is already in the group database, so nothing changes.
* `group:wronggid` is already in the group database, but its GID is wrong and
  must be corrected.

A legacy JSON definition file is also included to check that these are not
picked up by Holo anymore.
