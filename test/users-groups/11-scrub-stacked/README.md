In this testcase, the system supplies some user, which we assign to an
additional group through a user definition like

```toml
[[user]]
name = "foo"
groups = ["sys"]
```

Then this definition is removed. The next apply should restore the original set
of auxiliary groups. There are two such users in this testcase, one without and
one with auxiliary groups in the base image.
