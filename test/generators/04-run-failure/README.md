Tests that an error during a generator run aborts Holo entirely.

Any error during the generator phase MUST be fatal because we cannot continue
with an apply or even just a scan if we cannot be sure that we are seeing all
resource files.
