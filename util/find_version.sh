#!/bin/sh
#
# When run in the repository root directory, prints the application version.
#

if [ -d .git ]; then
    # best option: use git-describe
    git describe --tags --dirty
else
    # second-best option: when running inside an unpacked release tarball, the
    # root directory's name indicates the version, e.g. "holo-ssh-keys-1.2"
    root_basename="$(basename "$(readlink -f .)")"
    if [[ $root_basename =~ holo-ssh-keys-* ]]; then
        echo ${root_basename/holo-ssh-keys-/}
    else
        echo "Cannot determine application version. The root directory basename should look like 'holo-ssh-keys-1.2.3', but actually is '$root_basename'." >&2
        exit 1
    fi
fi
