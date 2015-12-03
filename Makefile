default: prepare-build build/holo-users-groups

prepare-build:
	@mkdir -p build/man
build/holo-users-groups: src/main.go src/*/*.go
	go build -o $@ $<

test: check # just a synonym
check:
	@holo-test holo-users-groups $(sort $(wildcard test/??-*))

install: build/holo-users-groups src/holorc.holoscript
	install -D -m 0755 build/holo-users-groups "$(DESTDIR)/usr/lib/holo/holo-users-groups"
	install -D -m 0755 src/holorc.holoscript   "$(DESTDIR)/usr/share/holo/files/01-holo-users-groups/etc/holorc.holoscript"

.PHONY: prepare-build test check install
