This test checks the success case for stacked entities.

* `group:stacked` has a minimal definition in `01-first.toml` and a complete
  definition in `02-second.toml`.
* `user:stacked` is the same, but for auxiliary groups (where we merge both
  sets), we check both entries only existing in one definition, and an entry
  existing in both definitions (which should only appear once in the merge
  result).
