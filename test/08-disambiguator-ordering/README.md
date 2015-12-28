This testcase tests an obscure edge case in disambiguator This testcase tests
obscure edge cases in disambiguator sorting: If a target has source files with
disambiguators like e.g. `01-foo` and `01-foo-bar` (i.e. one is a prefix of the
other), then the source files may be sorted differently at different layers of
the application, thereby resulting in a wrong application order, even when the
correct one is reported by `holo scan`.
