default: prepare-build build/holo-users-groups build/man/holo-users-groups.8

VERSION := $(shell ./util/find_version.sh)
# force people to use golangvend
GOCC := env GOPATH=$(CURDIR)/.gopath go

prepare-build:
	@mkdir -p build/man
build/holo-users-groups: cmd/holo-users-groups/*.go
	$(GOCC) build --ldflags "-s -w" -o $@ github.com/holocm/holo-users-groups/cmd/holo-users-groups

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) --center="Configuration Management" \
		--release="holo-users-groups $(VERSION)" \
		$< $@

test: check # just a synonym
check: default
	@holo-test holo-users-groups $(sort $(wildcard test/??-*))

install: default conf/holorc.holo-users-groups
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/users-groups"
	install -D -m 0755 build/holo-users-groups       "$(DESTDIR)/usr/lib/holo/holo-users-groups"
	install -D -m 0644 conf/holorc.holo-users-groups "$(DESTDIR)/etc/holorc.d/20-users-groups"
	install -D -m 0644 build/man/holo-users-groups.8 "$(DESTDIR)/usr/share/man/man8/holo-users-groups.8"

.PHONY: prepare-build test check install
