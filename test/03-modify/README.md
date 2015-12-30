This test checks scenarios of key sets changing and being reprovisioned:

* `user1` has two key sets provisioned. One of them has been deleted from the
  resource dir, so it shall be removed from `authorized_keys` too.
* `user2` has a set of two keys provisioned. The source file is changed by
  removing one key and adding a new one. This change shall be propagated to
  `authorized_keys`.
