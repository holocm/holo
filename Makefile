pkg = github.com/holocm/holo
bins = holo holo-files
mans = holorc.5 holo-plugin-interface.7 holo-test.7 holo.8 holo-files.8

default: prepare-build
default: $(addprefix build/,$(bins))
default: $(addprefix build/man/,$(mans))
.PHONY: default

GO            := GOPATH=$(CURDIR)/.go-workspace GOBIN=$(CURDIR)/build go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w
GO_TESTFLAGS  := -covermode=count
GO_DEPS       := $(GO) list -f '{{.ImportPath}}{{"\n"}}{{join .Deps "\n"}}'

prepare-build:
	@mkdir -p build/man

.version: FORCE
	./util/find_version.sh | util/write-ifchanged $@

cmd/holo/version.go: .version
	printf 'package main\n\nfunc init() {\n\tversion = "%s"\n}\n' "$$(cat $<)" > $@

$(addprefix %/,$(bins)): FORCE cmd/holo/version.go
	$(GO) install $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' $(addprefix $(pkg)/cmd/,$(bins))
build/%.test: build/% cmd/%/main_test.go
	$(GO) test -c -o $@ $(GO_TESTFLAGS) -coverpkg $$($(GO_DEPS) $(pkg)/cmd/$*|grep ^$(pkg)|tr '\n' ,|sed 's/,$$//') $(pkg)/cmd/$*

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod .version
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $$(cat .version)" \
		$< $@

test: check # just a synonym
check: default clean-tests $(foreach b,$(bins),build/$b.test)
	@if s="$$(gofmt -l cmd 2>/dev/null)"                        && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@if s="$$(find cmd -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	@$(GO) test $(GO_TESTFLAGS) $(pkg)/cmd/holo/internal
	@env HOLO_BINARY=../../build/holo bash util/holo-test holo $(sort $(wildcard test/??-*))

install: default conf/holorc conf/holorc.holo-files util/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
	install -D -m 0755 build/holo             "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 build/holo-files       "$(DESTDIR)/usr/lib/holo/holo-files"
	install -D -m 0755 util/holo-test         "$(DESTDIR)/usr/bin/holo-test"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
	install -D -m 0644 build/man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 build/man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 build/man/holo-files.8            "$(DESTDIR)/usr/share/man/man8/holo-files.8"
	install -D -m 0644 build/man/holo-test.7             "$(DESTDIR)/usr/share/man/man7/holo-test.7"
	install -D -m 0644 build/man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	env DESTDIR=$(DESTDIR) ./util/distribution-integration/install.sh

clean: clean-tests
	rm -fr -- build/ .go-workspace/pkg/
	rm -f -- .version cmd/holo/version.go
clean-tests:
	rm -fr -- test/*/{target,tree,{colored-,}{apply,apply-force,diff,scan}-output}

.PHONY: prepare-build test check install clean clean-tests
.PHONY: FORCE
