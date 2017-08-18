#!/bin/sh
echo "Running provisioning script 01-successful.sh"
echo "This is output on stdout"
sleep 0.1 # ensure that output arrives in the correct order
echo "This is output on stderr" >&2
sleep 0.1 # ensure that output arrives in the correct order
echo "Done with 01-successful.sh, exiting with code 0"
