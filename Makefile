pkg = github.com/holocm/holo
mans = holorc.5 holo-plugin-interface.7 holo-test.7 holo.8 holo-files.8 holo-run-scripts.8 holo-ssh-keys.8 holo-users-groups.8

default: build/holo
default: $(addprefix build/man/,$(mans))
.PHONY: default

GO            := go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w
GO_TESTFLAGS  := -covermode=count
GO_DEPS       := $(GO) list -f '{{.ImportPath}}{{"\n"}}{{join .Deps "\n"}}'

# which packages to test with static checkers
allpkgs := $(shell go list ./...)
# which files to test with static checkers (this contains a list of globs)
allfiles := $(addsuffix /*.go,$(patsubst $(shell go list .),.,$(shell go list ./...)))
# to get around weird Makefile syntax restrictions, we need variables containing a space and comma
space := $(null) $(null)
comma := ,

build build/man:
	@mkdir -p $@

.version: FORCE
	./util/find_version.sh | util/write-ifchanged $@

cmd/holo/version.go: .version
	printf 'package entrypoint\n\nfunc init() {\n\tversion = "%s"\n}\n' "$$(cat $<)" > $@

build/holo: FORCE cmd/holo/version.go | build
	$(GO) build -o $@ $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' $(pkg)
build/holo.test: build/holo main_test.go
	$(GO) test -c -o $@ $(GO_TESTFLAGS) -coverpkg=$(subst $(space),$(comma),$(allpkgs)) $(pkg)

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod .version | build/man
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $$(cat .version)" \
		$< $@

test: check # just a synonym
check: default test/cov.html test/cov.func.txt
test/cov.cov: clean-tests build/holo.test
	@if s="$$(gofmt -l $(allfiles) 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@if s="$$(golint   $(allpkgs)  2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	@$(GO) test $(GO_TESTFLAGS) -coverprofile=test/cov/holo-output.cov $(pkg)/cmd/holo/internal
	@$(GO) test $(GO_TESTFLAGS) -coverprofile=test/cov/ssh-keys-output.cov $(pkg)/cmd/holo-ssh-keys/impl
	@HOLO_BINARY="$$PWD/build/holo.test" HOLO_TEST_COVERDIR=$$PWD/test/cov ./util/holo-test-help
	@\
		export HOLO_BINARY=../../../build/holo.test && \
		export HOLO_TEST_COVERDIR=$(abspath test/cov) && \
		export HOLO_TEST_SCRIPTPATH=../../../util && \
		$(foreach p,files run-scripts ssh-keys users-groups generators,\
			ln -sfT ../build/holo.test test/holo-$p && \
			./util/holo-test holo-$p $(sort $(wildcard test/$p/??-*)) && ) \
		true
	util/gocovcat.go test/cov/*.cov > test/cov.cov
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
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/generators"
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

clean: clean-tests
	rm -fr -- build/
	rm -f -- .version cmd/holo/version.go
clean-tests:
	rm -fr -- test/*/*/target
	rm -f -- test/*/*/{tree,{colored-,}{apply,apply-force,diff,scan}-output}
	rm -f -- test/cov.* test/cov/* test/holo-*

vendor: FORCE
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: test check install clean clean-tests vendor
.PHONY: FORCE
