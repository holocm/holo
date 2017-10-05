There was a bug where applying a symlink over a file would always display
the `Working on file:/path/to/file` report, regardless of whether the target
was already provisioned. This test checks that `holo apply` stays silent in
this case when no changes are made.
