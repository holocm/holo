This test checks the handling of user entites.

* `user:new` is not yet in the user database and will be created.
* `user:minimal` is the same, but it has no attributes except its name, to
  check that the user is created correctly when no extra attributes are given.
* `user:existing` is already in the user database, so nothing changes.
* `user:wronguid` is already in the user database, but its UID is wrong and
  must be corrected.
* `user:wronghome`, `user:wronggroup`, `user:wronggroups`, `user:wrongshell` are similar.

A legacy JSON definition file is also included to check that these are not
picked up by Holo anymore.
