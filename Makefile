default: prepare-build build/holo-ssh-keys build/man/holo-ssh-keys.8

VERSION := $(shell ./util/find_version.sh)

prepare-build:
	@mkdir -p build/man
build/holo-ssh-keys: cmd/holo-ssh-keys/main.go cmd/holo-ssh-keys/*/*.go
	go build --ldflags "-s -w" -o $@ $<

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) --center="Configuration Management" \
		--release="holo-ssh-keys $(VERSION)" \
		$< $@

test: check # just a synonym
check: default
	@go test ./cmd/holo-ssh-keys/impl
	@holo-test holo-ssh-keys $(sort $(wildcard test/??-*))

install: default cmd/holo-ssh-keys/holorc
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/ssh-keys"
	install -D -m 0755 build/holo-ssh-keys       "$(DESTDIR)/usr/lib/holo/holo-ssh-keys"
	install -D -m 0644 conf/holorc.holo-ssh-keys "$(DESTDIR)/etc/holorc.d/25-ssh-keys"
	install -D -m 0644 build/man/holo-ssh-keys.8 "$(DESTDIR)/usr/share/man/man8/holo-ssh-keys.8"

.PHONY: prepare-build test check install
