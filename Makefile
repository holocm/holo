default: prepare-build build/man/holo-run-scripts.8

VERSION := $(shell git describe --tags --dirty)

prepare-build:
	@mkdir -p build/man

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) --center="Configuration Management" \
		--release="holo-run-scripts $(VERSION)" \
		$< $@

test: check # just a synonym
check: default
	@holo-test holo-run-scripts $(sort $(wildcard test/??-*))

install: default src/holorc.holoscript
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/run-scripts"
	install -D -m 0755 src/holo-run-scripts         "$(DESTDIR)/usr/lib/holo/holo-run-scripts"
	install -D -m 0755 src/holorc.holoscript        "$(DESTDIR)/usr/share/holo/files/95-holo-run-scripts/etc/holorc.holoscript"
	install -D -m 0644 build/man/holo-run-scripts.8 "$(DESTDIR)/usr/share/man/man8/holo-run-scripts.8"
