#!/bin/sh

case "$1" in
    info)
        echo MIN_API_VERSION=3
        echo MAX_API_VERSION=3
        ;;
    scan)
        # one magical entity that dumps its resource dir
        echo "ENTITY: list-resource-dir"

        # one entity per txt file that prints its contents (this verifies that
        # the symlinks in the generated resource dir are valid)
        #
        # NOTE: `cut -c3-` removes the leading "./", e.g. "./file.txt" -> "file.txt"
        cd "${HOLO_RESOURCE_DIR}"
        find -L . -name \*.txt | sort | cut -c3- | while read FILE; do
            echo "ENTITY: print:${FILE}"
            echo "found at: ${HOLO_RESOURCE_DIR}/${FILE}"
            echo "SOURCE: ${HOLO_RESOURCE_DIR}/${FILE}"
        done
        ;;
    diff)
        ;;
    apply|force-apply)
        ENTITY_ID="$2"
        if [ "${ENTITY_ID}" = list-resource-dir ]; then
            "$(dirname "$0")/../../util/tree-to-dump.sh" "${HOLO_RESOURCE_DIR}"
        else
            FILE="${ENTITY_ID:6}" # strip "print:" prefix
            cat "${HOLO_RESOURCE_DIR}/${FILE}"
        fi
        ;;
    *)
        exit 1
        ;;
esac
