#!/bin/bash

set -euo pipefail

# enumerate files that are managed by holo-files
holo scan --short | grep '^file:' | cut -d: -f2- | sort > /tmp/files-managed-by-holo

# changed files come from stdin (but without leading slash!); find files
# managed by Holo among these
sort | sed 's+^+/+' | comm -12 - /tmp/files-managed-by-holo | sed 's/^/file:/' | xargs -r holo apply

# cleanup
rm /tmp/files-managed-by-holo
