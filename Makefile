BUILD_TARGETS += $(patsubst cmd/%,.gopath/bin/%, $(wildcard cmd/*))         # build all binaries
BUILD_TARGETS += $(patsubst doc/%.pod,.gopath/man/%, $(wildcard doc/*.pod)) # build all manpages

default: $(BUILD_TARGETS)

VERSION := $(shell ./util/find_version.sh)

GO            := env GOPATH=$(CURDIR)/.gopath GOBIN=$(CURDIR)/.gopath/bin go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w

.gopath/bin/holo: cmd/holo/main.go lib/holo/*.go
	$(GO) install $(GO_BUILDFLAGS) --ldflags "$(GO_LDFLAGS) -X github.com/holocm/holo/cmd/holo.version=$(VERSION)" github.com/holocm/holo/cmd/holo
.gopath/bin/holo-files: cmd/holo-files/main.go lib/holo-files/*/*.go
	$(GO) install $(GO_BUILDFLAGS) --ldflags "$(GO_LDFLAGS)" github.com/holocm/holo/cmd/holo-files

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
.gopath/man/%: doc/%.pod
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $(VERSION)" \
		$< $@

test: check # just a synonym
check: default clean-tests
	@if s="$$(gofmt -l cmd lib 2>/dev/null)"                        && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@if s="$$(find cmd lib -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	@$(GO) test ./lib/holo
	@env HOLO_BINARY=../../.gopath/bin/holo bash util/holo-test holo $(sort $(wildcard test/??-*))

install: default conf/holorc conf/holorc.holo-files util/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
	install -D -m 0755 .gopath/bin/holo       "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 .gopath/bin/holo-files "$(DESTDIR)/usr/lib/holo/holo-files"
	install -D -m 0755 util/holo-test         "$(DESTDIR)/usr/bin/holo-test"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
	install -D -m 0644 .gopath/man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 .gopath/man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 .gopath/man/holo-files.8            "$(DESTDIR)/usr/share/man/man8/holo-files.8"
	install -D -m 0644 .gopath/man/holo-test.7             "$(DESTDIR)/usr/share/man/man7/holo-test.7"
	install -D -m 0644 .gopath/man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	env DESTDIR=$(DESTDIR) ./util/distribution-integration/install.sh

clean: clean-tests
	rm -fr -- $(BUILD_TARGETS)
clean-tests:
	rm -fr -- test/*/{target,tree,{colored-,}{apply,apply-force,diff,scan}-output}

.PHONY: test check install clean clean-tests
