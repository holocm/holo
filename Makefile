bins = holo holo-files
mans = holorc.5 holo-plugin-interface.7 holo-test.7 holo.8 holo-files.8

default: prepare-build
default: $(addprefix build/,$(bins))
default: $(addprefix build/man/,$(mans))
.PHONY: default

VERSION := $(shell ./util/find_version.sh)

GO            := GOPATH=$(CURDIR)/.go-workspace GOBIN=$(CURDIR)/build go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w -X main.version=$(VERSION)

prepare-build:
	@mkdir -p build/man

$(addprefix %/,$(bins)): FORCE
	$(GO) install $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' $(addprefix github.com/holocm/holo/cmd/,$(bins))

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $(VERSION)" \
		$< $@

test: check # just a synonym
check: default clean-tests
	@if s="$$(gofmt -l cmd 2>/dev/null)"                        && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@if s="$$(find cmd -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	@$(GO) test github.com/holocm/holo/cmd/holo/internal
	@env HOLO_BINARY=../../build/holo bash src/holo-test holo $(sort $(wildcard test/??-*))

install: default conf/holorc conf/holorc.holo-files src/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
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
	env DESTDIR=$(DESTDIR) ./src/distribution-integration/install.sh

clean: clean-tests
	rm -fr -- build/holo build/holo-files build/man .go-workspace/pkg
clean-tests:
	rm -fr -- test/*/{target,tree,{colored-,}{apply,apply-force,diff,scan}-output}

.PHONY: prepare-build test check install clean clean-tests
.PHONY: FORCE
