default: prepare-build
default: build/holo build/holo-files
default: build/man/holorc.5 build/man/holo-plugin-interface.7 build/man/holo-test.7 build/man/holo.8 build/man/holo-files.8

VERSION := $(shell git describe --tags --dirty)

prepare-build:
	@mkdir -p build/man
build/holo: src/holo/main.go src/holo/*/*.go
	go build --ldflags "-X main.version=$(VERSION)" -o $@ $<
build/holo-files: src/holo-files/main.go src/holo-files/*/*.go
	go build -o $@ $<

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $(VERSION)" \
		$< $@

test: check # just a synonym
check: default
	@go test ./src/holo/impl
	@env HOLO_BINARY=../../build/holo bash src/holo-test holo $(sort $(wildcard test/??-*))

install: default src/holorc src/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -D -m 0644 src/holorc             "$(DESTDIR)/etc/holorc"
	install -D -m 0755 build/holo             "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 build/holo-files       "$(DESTDIR)/usr/lib/holo/holo-files"
	install -D -m 0755 src/holo-test          "$(DESTDIR)/usr/bin/holo-test"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
	install -D -m 0644 build/man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 build/man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 build/man/holo-files.8            "$(DESTDIR)/usr/share/man/man8/holo-files.8"
	install -D -m 0644 build/man/holo-test.7             "$(DESTDIR)/usr/share/man/man7/holo-test.7"
	install -D -m 0644 build/man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"

.PHONY: prepare-build test check install
