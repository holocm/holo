pkg = github.com/holocm/holo
mans = holorc.5 holo-plugin-interface.7 holo-test.7 holo.8 holo-files.8 holo-run-scripts.8 holo-ssh-keys.8 holo-users-groups.8

default: build/holo
default: $(addprefix build/man/,$(mans))
.PHONY: default

GO            := GOPATH=$(CURDIR)/.go-workspace GOBIN=$(CURDIR)/build go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w
GO_TESTFLAGS  := -covermode=count
GO_DEPS       := $(GO) list -f '{{.ImportPath}}{{"\n"}}{{join .Deps "\n"}}'

# go_pkgs and go_dirs are obvious.
go_pkgs = $(sort $(filter-out $(pkg)/vendor%,$(filter $(pkg) $(pkg)/%,$(shell $(GO_DEPS) $(pkg)))))
go_dirs = $(patsubst $(pkg)%,.%,$(go_pkgs))
# go_srcs is like go_dirs, but for tools that crawl directories; it is
# the minimal list of files and directories that will include all of
# our Go sources, without just saying "." (we don't want to just say
# "." because that would crawl ".git" and ".go-workspace" and other
# things that it should skip).
go_srcs = $(sort $(foreach f,$(patsubst .,*.go,$(patsubst ./%,%,$(go_dirs))),$(firstword $(subst /, ,$f))))

build build/man:
	@mkdir -p $@

.version: FORCE
	./util/find_version.sh | util/write-ifchanged $@

cmd/holo/version.go: .version
	printf 'package entrypoint\n\nfunc init() {\n\tversion = "%s"\n}\n' "$$(cat $<)" > $@

build/holo: FORCE cmd/holo/version.go | build
	$(GO) install $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' $(pkg)
build/holo.test: build/holo main_test.go
	$(GO) test -c -o $@ $(GO_TESTFLAGS) -coverpkg $$(echo $(go_pkgs) | tr ' ' ,) $(pkg)

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod .version | build/man
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $$(cat .version)" \
		$< $@

test: check # just a synonym
.PHONY: test

check: default
check: check-gofmt
check: check-golint
check: test/cov.html test/cov.func.txt
.PHONY: check

check-gofmt:
	@if s="$$(gofmt -l $(go_srcs) 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
check-golint:
	@if s="$$(golint   $(go_dirs) 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
%/go-test.cov: clean-tests
	@$(GO) test $(GO_TESTFLAGS) -coverprofile=$@ $(pkg)/$*
check-holo-test-help: clean-tests build/holo.test util/holo-test-help
	@HOLO_BINARY="$$PWD/build/holo.test" HOLO_TEST_COVERDIR=$$PWD/test/cov ./util/holo-test-help
test/%/check: test/%/source-tree clean-tests build/holo.test util/holo-test
	@\
		export HOLO_BINARY=../../../build/holo.test && \
		export HOLO_TEST_COVERDIR=$(abspath test/cov) && \
		export HOLO_TEST_SCRIPTPATH=../../../util && \
		ln -sfT ../build/holo.test test/holo-$(firstword $(subst /, ,$*)) && \
		./util/holo-test holo-$(firstword $(subst /, ,$*)) $(@D)
.PHONY: check-% test/%/check

test/cov.cov: $(sort $(shell find $(filter-out %.go,$(go_srcs)) -name '*_test.go' -printf '%h/go-test.cov\n'))
test/cov.cov: check-holo-test-help
test/cov.cov: $(sort $(patsubst %/source-tree,%/check,$(wildcard test/*/??-*/source-tree)))
test/cov.cov: util/gocovcat.go
	util/gocovcat.go test/cov/*.cov $(filter %.cov,$^) > $@
%.html: %.cov
	$(GO) tool cover -html $< -o $@
%.func.txt: %.cov
	$(GO) tool cover -func $< -o $@

DIST_IDS = $(shell [ -f /etc/os-release ] && source /etc/os-release || source /usr/lib/os-release; echo "$$ID $$ID_LIKE")

install: default conf/holorc conf/holorc.holo-files util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/run-scripts"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/ssh-keys"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/users-groups"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/users-groups/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/users-groups/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/lib/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/run-scripts"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/ssh-keys"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/users-groups"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
	install -D -m 0644 conf/holorc.holo-run-scripts "$(DESTDIR)/etc/holorc.d/95-holo-run-scripts"
	install -D -m 0644 conf/holorc.holo-ssh-keys "$(DESTDIR)/etc/holorc.d/25-ssh-keys"
	install -D -m 0644 conf/holorc.holo-users-groups "$(DESTDIR)/etc/holorc.d/20-users-groups"
	install -D -m 0755 build/holo             "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 cmd/holo-run-scripts   "$(DESTDIR)/usr/lib/holo/holo-run-scripts"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
	install -D -m 0644 build/man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 build/man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 build/man/holo-files.8            "$(DESTDIR)/usr/share/man/man8/holo-files.8"
	install -D -m 0644 build/man/holo-run-scripts.8      "$(DESTDIR)/usr/share/man/man8/holo-run-scripts.8"
	install -D -m 0644 build/man/holo-ssh-keys.8         "$(DESTDIR)/usr/share/man/man8/holo-ssh-keys.8"
	install -D -m 0644 build/man/holo-users-groups.8     "$(DESTDIR)/usr/share/man/man8/holo-users-groups.8"
	install -D -m 0644 build/man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	ln -sfT ../../bin/holo "$(DESTDIR)/usr/lib/holo/holo-files"
	ln -sfT ../../bin/holo "$(DESTDIR)/usr/lib/holo/holo-ssh-keys"
	ln -sfT ../../bin/holo "$(DESTDIR)/usr/lib/holo/holo-users-groups"
ifneq ($(filter arch,$(DIST_IDS)),)
	install -D -m 0644 util/distribution-integration/alpm.hook    "$(DESTDIR)/usr/share/libalpm/hooks/01-holo-resolve-pacnew.hook"
	install -D -m 0755 util/distribution-integration/alpm-hook.sh "$(DESTDIR)/usr/share/libalpm/scripts/holo-resolve-pacnew"
endif
.PHONY: install

clean: clean-tests
	rm -fr -- build/ .go-workspace/pkg/
	rm -f -- .version cmd/holo/version.go
clean-tests:
	rm -fr -- test/*/*/target
	rm -f -- test/*/*/{tree,{colored-,}{apply,apply-force,diff,scan}-output}
	rm -f -- test/cov.* test/cov/* test/holo-*
	find -name go-test.cov -delete
.PHONY: clean clean-tests

vendor: FORCE
	@# vendoring by https://github.com/holocm/golangvend
	golangvend
.PHONY: vendor

.PHONY: FORCE
