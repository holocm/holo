#!/bin/bash

CURRENT_DIR="$(dirname $0)"

# check distribution and install appropriate integrations
[ -f /etc/os-release ] && source /etc/os-release || source /usr/lib/os-release
DIST_IDS="$(echo "$ID $ID_LIKE" | tr ' ' ',')"

case ",$DIST_IDS," in
    *,arch,*)
        set -ex
        install -D -m 0644 "${CURRENT_DIR}/alpm.hook"    "${DESTDIR}/usr/share/libalpm/hooks/01-holo-resolve-pacnew.hook"
        install -D -m 0755 "${CURRENT_DIR}/alpm-hook.sh" "${DESTDIR}/usr/lib/holo/alpm-hook-resolve-pacnew"
        ;;
    *)
        echo No specialized integration for this distribution.
        ;;
esac
