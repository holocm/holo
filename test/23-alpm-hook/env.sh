holo_topdir="$(readlink -f ../../)"
holo_wrapper_BINARY="$(readlink -f -- "$HOLO_BINARY")"
holo_wrapper() (
	TMPDIR="$(readlink -f -- "$TMPDIR")"
	cd "$HOLO_ROOT_DIR"
	HOLO_ROOT_DIR=.
	case "$1" in
		apply)
			printf '%s\n' \
			       etc/targetfile-deleted-with-pacsave.conf \
			       etc/targetfile-with-pacnew.conf \
			| PATH="${holo_topdir}/build:$PATH" "${holo_topdir}/util/distribution-integration/alpm-hook.sh"
			;;
		*)
			exec $holo_wrapper_BINARY "$@"
			;;
	esac
)
HOLO_BINARY=holo_wrapper
