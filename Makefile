default:
	@echo 'Nothing to do; just continue with `make check` or `make install`'

test: check # just a synonym
check:
	@/usr/lib/holo/holo-test holo-run-scripts $(sort $(wildcard test/??-*))

install: src/holo-run-scripts src/holorc.holoscript
	install -D -m 0755 src/holo-run-scripts  "$(DESTDIR)/usr/lib/holo/holo-run-scripts"
	install -D -m 0755 src/holorc.holoscript "$(DESTDIR)/usr/share/holo/files/95-holo-run-scripts/etc/holorc.holoscript"
