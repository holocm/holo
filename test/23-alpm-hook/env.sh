holo_topdir="$(readlink -f ../../)"
holo_wrapper() (
	TMPDIR="$(readlink -f -- "$TMPDIR")"
	cd "$HOLO_ROOT_DIR"
	HOLO_ROOT_DIR=.
	case "$1" in
		apply)
			printf '%s\n' \
			       etc/targetfile-deleted-with-pacsave.conf \
			       etc/targetfile-with-pacnew.conf \
			| PATH="usr/bin:$PATH" "${holo_topdir}/util/distribution-integration/alpm-hook.sh"
			;;
		*)
			exec usr/bin/holo "$@"
			;;
	esac
)
HOLO_BINARY=holo_wrapper
