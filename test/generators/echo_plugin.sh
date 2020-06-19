#!/bin/sh

case "$1" in
    info)
        echo MIN_API_VERSION=3
        echo MAX_API_VERSION=3
        ;;
    scan)
        # list executables in $HOLO_RESOURCE_DIR
        echo "ENTITY: env:$HOLO_RESOURCE_DIR"
        echo "ACTION: Printing"
        echo "found at: HOLO_RESOURCE_DIR"
        echo "SOURCE: $HOLO_RESOURCE_DIR/$FILENAME"
        ;;
    diff)
        ;;
    apply|force-apply)
        ;;
    *)
        exit 1
        ;;
esac