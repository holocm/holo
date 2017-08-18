#!/bin/sh
echo "Running provisioning script 02-failing.sh"
echo "This is output on stdout"
sleep 0.1 # ensure that output arrives in the correct order
echo "This is output on stderr" >&2
sleep 0.1 # ensure that output arrives in the correct order
echo "Done with 02-failing.sh, exiting with code 1"
exit 1
