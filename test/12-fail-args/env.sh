holo_wrapper_BINARY=$HOLO_BINARY
holo_wrapper() {
	case "$1" in
		apply|diff|scan)
			set -- "$@" bogus-selector
			;;
	esac
	$holo_wrapper_BINARY "$@"
}
HOLO_BINARY=holo_wrapper
