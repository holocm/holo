#!/bin/bash
#
# Usage: tree-to-dump.sh <directory>
#
# Reproducibly prints a serialization of all directories, files and symlinks
# below <directory>, that can later be rematerialized with dump-to-tree.sh.
#

set -euo pipefail
LC_ALL=C

SEPARATOR="$(echo "----------------------------------------"; echo ">")"
SEPARATOR="----------------------------------------"

dump() {
  local DIR_PATH="$1"
  local PREFIX="$2"

  for ENTRY in "${DIR_PATH}"/* "${DIR_PATH}"/.ssh; do
    if [ -L "${ENTRY}" ]; then
      echo "symlink   0777 ${PREFIX}/$(basename "${ENTRY}")"
      readlink "${ENTRY}"
      echo "${SEPARATOR}"

    elif [ -f "${ENTRY}" ]; then
      echo "file      $(stat -c 0%a "${ENTRY}") ${PREFIX}/$(basename "${ENTRY}")"
      cat "${ENTRY}"
      echo "${SEPARATOR}"

    elif [ -d "${ENTRY}" ]; then
      local NEXT_PREFIX="${PREFIX}/$(basename "${ENTRY}")"
      local CONTENTS="$(dump "${ENTRY}" "${NEXT_PREFIX}")"
      # omit directories with default mode (0755) whose existence is implied by children
      local MODE="$(stat -c 0%a "${ENTRY}")"
      if [ -z "${CONTENTS}" -o "${MODE}" != "0755" ]; then
        echo "directory ${MODE} ${NEXT_PREFIX}/"
        echo "${SEPARATOR}"
      fi
      if [ -n "${CONTENTS}" ]; then
        echo "${CONTENTS}"
      fi
    fi
  done
}

dump "${1:-.}" "."
