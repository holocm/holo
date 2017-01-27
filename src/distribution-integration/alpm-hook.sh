#!/bin/bash

set -euo pipefail

comm -12 \
     <(# changed files come from stdin (but without leading slash!)
       sort | sed 's+^+/+') \
     <(# enumerate files that are managed by holo-files
       holo scan --short | sed -n 's/^file://p') \
| sed 's/^/file:/' | xargs -r holo apply
