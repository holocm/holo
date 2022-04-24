#!/bin/bash
#
# Usage: dump-to-tree.sh <directory>
#
# Materializes the filesystem dump on stdin (in the format produced by
# tree-to-dump.sh) into the given filesystem directory. The directory will be
# created if it does not exist.
#

set -euo pipefail

mkdir -p "$1"
cd "$1"

LINE_NUMBER=0
read_and_inc() {
  LINE_NUMBER=$((LINE_NUMBER+1))
  read "$@"
}
read_and_inc_or_fail() {
  read_and_inc "$@" || fail "unexpected end of input"
}
fail() {
    echo "error: $@ on input line ${LINE_NUMBER}" >&2
    exit 1
}

while read_and_inc FILE_TYPE FILE_MODE FILE_PATH; do
  if [ -z "${FILE_TYPE}" -a -z "${FILE_MODE}" -a -z "${FILE_PATH}" ]; then
    continue # ignore empty lines
  fi
  if [ -z "${FILE_TYPE}" -o -z "${FILE_MODE}" -o -z "${FILE_PATH}" ]; then
    fail "unexpected end of entry header"
  fi
  if [[ "${FILE_PATH}" = /* ]]; then
    fail "absolute path not allowed in entry header"
  fi

  case "${FILE_TYPE}" in
    file)
      install -D -m "${FILE_MODE}" /dev/null "${FILE_PATH}"
      # header is followed by file content, terminated by a separator line like "---------------"
      while IFS='' read_and_inc_or_fail -r LINE; do
        if [[ "${LINE}" =~ ^-+$ ]]; then
          break
        fi
        echo "${LINE}" >> "${FILE_PATH}"
      done || true
      ;;

    symlink)
      # next line contains symlink target, followed by a separator line
      read_and_inc_or_fail TARGET
      [ -z "${TARGET}" ] && fail "missing symlink target"
      read_and_inc_or_fail SEPARATOR
      [[ "${SEPARATOR}" =~ ^-+$ ]] || fail "unexpected content line, expected separator"
      mkdir -p "$(dirname "${FILE_PATH}")"
      ln -sf "${TARGET}" "${FILE_PATH}"
      ;;

    directory)
      install -d -m "${FILE_MODE}" "${FILE_PATH}"
      read_and_inc_or_fail SEPARATOR
      [[ "${SEPARATOR}" =~ ^-+$ ]] || fail "unexpected content line, expected separator"
      ;;

    *)
      fail "unknown file type in entry header"
      ;;
  esac
done || true
