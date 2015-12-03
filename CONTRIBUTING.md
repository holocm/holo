How to contribute
=================

Holo uses GitHub's tools (the issue tracker, pull requests and release
management) for its development. So if you want to contribute, have a look at
the open issues, fork the repo, start hacking and submit pull requests.

If you have any questions concerning the code structure or internals, ask your
question as an issue and I'll do my best to explain everything to you.

Coding style
------------

Please run `gofmt` (or `goimports`) and
[`golint`](https://github.com/golang/lint) on your code before committing.

Branches
--------

Within Holo's branching model the `stable` branch is the current stable
release, and development for the next stable release happens on the `master`
branch. Therefore, users can always compile the `stable` branch to get the
latest bugfixes, without fear of unexpected instability.

Bugfixes should be developed on the stable branch, since forward-merging to
the development branch is always easier than cherry-picking back into the
stable branch.

Documentation
-------------

Documentation is written in POD (Perl's documentation format), since that
format has converters to manpage and HTML readily available on all
distributions (through the `pod2man` and `pod2html` executables included with
Perl).

The manpages are located in the `doc` directory and can be built with `make`.
