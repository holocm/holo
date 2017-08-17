There was a bug once that caused unchanged files to be reported during `holo
apply --force` (but not during `holo apply` without `--force`). This testcase
checks for this bug by having one file require `--force` while the other one is
unchanged.
