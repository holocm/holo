This testcase checks the basic behavior with single, plain repo files and
target files of all kinds (regular or symlink).

    /etc/plain-over-plain.conf          # stock config is plain file, repo has plain file
    /etc/link-over-plain.conf           # stock config is plain file, repo has link file
    /etc/plain-over-link.conf           # stock config is link file, repo has plain file
    /etc/link-over-link.conf            # stock config is link file, repo has link file

Also, some error cases are tested:

* `/etc/stock-file-missing.conf` has a repo file, but not a stock config file.
* `/etc/stock-file-is-directory.conf` has a repo file, but the target is a
  directory (and thus not a manageable file).

We also check if files not conforming to the repo file naming pattern are
ignored correctly by the scanner.

* `repo/not-a-repo-file.conf` is not within a subdirectory of the repo directory.

Furthermore, provisioning script functionality is tested:

    /usr/share/holo/provision/01-successful.sh          # prints output, exits successfully
    /usr/share/holo/provision/02-failing.sh             # prints output, exits with failure
    /usr/share/holo/provision/03-successful-nooutput.sh # does not print, exits successfully
    /usr/share/holo/provision/04-failing-nooutput.sh    # does not print, exits with failure
