This test checks what happens when definitions for provisioned entities are
deleted.

* `holo apply` will report that `group:deleted` and `user:deleted` are
  orphaned.
* `holo apply --force` will delete these entities.
