This test checks all the conflicts that can arise during scanning of stacked entities.

* `group:stacked` and `user:stacked` have two maximally conflicting definitions
  in `01-first.toml` and `02-second.toml`.

* `group:valid` and `user:valid` are valid definitions in `01-first.toml`.
  Through their presence, we check that an error in one definition does not
  affect intact definitions

And since we're at it anyway, we also check the error case where an entity
definition is missing the required `name` attribute. By placing these broken
groups in `01-first.toml` (while the merge conflicts arise in `02-second.toml`),
we also check that the entity scanner keeps going and reports all errors at
once.
