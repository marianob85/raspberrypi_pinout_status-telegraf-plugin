next_version :=  $(shell cat build_version.txt)
tag := $(shell git describe --exact-match --tags 2>git_describe_error.tmp; rm -f git_describe_error.tmp)
branch := $(shell git rev-parse --abbrev-ref HEAD)
commit := $(shell git rev-parse --short=8 HEAD)
glibc_version := 2.17
plugin_name := raspberrypi_pinout_status-telegraf-plugin

ifdef NIGHTLY
	version := $(next_version)
	rpm_version := nightly
	rpm_iteration := 0
	deb_version := nightly
	deb_iteration := 0
	tar_version := nightly
else ifeq ($(tag),)
	version := $(next_version)
	rpm_version := $(version)~$(commit)-0
	rpm_iteration := 0
	deb_version := $(version)~$(commit)-0
	deb_iteration := 0
	tar_version := $(version)~$(commit)
else ifneq ($(findstring -rc,$(tag)),)
	version := $(word 1,$(subst -, ,$(tag)))
	version := $(version:v%=%)
	rc := $(word 2,$(subst -, ,$(tag)))
	rpm_version := $(version)-0.$(rc)
	rpm_iteration := 0.$(subst rc,,$(rc))
	deb_version := $(version)~$(rc)-1
	deb_iteration := 0
	tar_version := $(version)~$(rc)
else
	version := $(tag:v%=%)
	rpm_version := $(version)-1
	rpm_iteration := 1
	deb_version := $(version)-1
	deb_iteration := 1
	tar_version := $(version)
endif

MAKEFLAGS += --no-print-directory
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
HOSTGO := env -u GOOS -u GOARCH -u GOARM -- go

LDFLAGS := $(LDFLAGS) -X main.commit=$(commit) -X main.branch=$(branch) -X main.goos=$(GOOS) -X main.goarch=$(GOARCH)
ifneq ($(tag),)
	LDFLAGS += -X main.version=$(version)
endif

# Go built-in race detector works only for 64 bits architectures.
ifneq ($(GOARCH), 386)
	race_detector := -race
endif


GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l -s $(filter-out plugins/parsers/influx/machine.go, $(GOFILES)))

prefix ?= /usr/local
bindir ?= $(prefix)/bin
sysconfdir ?= $(prefix)/etc
pkgdir ?= build/dist

.PHONY: all
all:
	@$(MAKE) deps
	@$(MAKE) $(plugin_name)

.PHONY: help
help:
	@echo 'Targets:'
	@echo '  all        - download dependencies and compile telegraf binary'
	@echo '  deps       - download dependencies'
	@echo '  $(plugin_name)   - compile telegraf binary'
	@echo '  test       - run short unit tests'
	@echo '  fmt        - format source files'
	@echo '  tidy       - tidy go modules'
	@echo '  check-deps - check docs/LICENSE_OF_DEPENDENCIES.md'
	@echo '  clean      - delete build artifacts'
	@echo ''
	@echo 'Package Targets:'
	@$(foreach dist,$(dists),echo "  $(dist)";)

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: $(plugin_name)
$(plugin_name):
	go build -ldflags "$(LDFLAGS)" ./cmd/$(plugin_name)

# Used by dockerfile builds
.PHONY: go-install
go-install:
	go install -mod=mod -ldflags "-w -s $(LDFLAGS)" ./cmd/$(plugin_name)

.PHONY: test
test:
	go test -v -short $(race_detector) ./...

.PHONY: fmt
fmt:
	@gofmt -s -w $(filter-out plugins/parsers/influx/machine.go, $(GOFILES))

.PHONY: fmtcheck
fmtcheck:
	@if [ ! -z "$(GOFMT)" ]; then \
		echo "[ERROR] gofmt has found errors in the following files:"  ; \
		echo "$(GOFMT)" ; \
		echo "" ;\
		echo "Run make fmt to fix them." ; \
		exit 1 ;\
	fi

.PHONY: test-windows
test-windows:
	go test -short ./...

.PHONY: vet
vet:
	@echo 'go vet $$(go list ./... | grep -v ./plugins/parsers/influx)'
	@go vet $$(go list ./... | grep -v ./plugins/parsers/influx) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "go vet has found suspicious constructs. Please remediate any reported errors"; \
		echo "to fix them before submitting code for review."; \
		exit 1; \
	fi

.PHONY: tidy
tidy:
	go mod verify
	go mod tidy
	@if ! git diff --quiet go.mod go.sum; then \
		echo "please run go mod tidy and check in changes"; \
		exit 1; \
	fi

.PHONY: check
check: fmtcheck vet
	@$(MAKE) --no-print-directory tidy

.PHONY: test-all
test-all: fmtcheck vet
	go test $(race_detector) ./...

.PHONY: check-deps
check-deps:
	./scripts/check-deps.sh

.PHONY: clean
clean:
	rm -f $(plugin_name)
	rm -f $(plugin_name).exe
	rm -rf build

.PHONY: install
install: $(buildbin)
	@mkdir -pv $(DESTDIR)$(bindir)
	@mkdir -pv $(DESTDIR)$(sysconfdir)
	@if [ $(GOOS) != "windows" ]; then mkdir -pv $(DESTDIR)$(sysconfdir)/telegraf; fi
	@cp -fv $(buildbin) $(DESTDIR)$(bindir)
	@if [ $(GOOS) != "windows" ]; then cp -fv etc/$(plugin_name).config $(DESTDIR)$(sysconfdir)/telegraf/$(plugin_name).config$(conf_suffix); fi
	@if [ $(GOOS) = "windows" ]; then cp -fv etc/$(plugin_name).config $(DESTDIR)/$(plugin_name).config; fi

# Telegraf build per platform.  This improves package performance by sharing
# the bin between deb/rpm/tar packages over building directly into the package
# directory.
$(buildbin):
	@mkdir -pv $(dir $@)
	go build -o $(dir $@) -ldflags "$(LDFLAGS)" ./cmd/$(plugin_name)

debs := $(plugin_name)_$(deb_version)_arm64.deb
debs += $(plugin_name)_$(deb_version)_armel.deb
debs += $(plugin_name)_$(deb_version)_armhf.deb

dists := $(debs) 

.PHONY: package
package: $(dists)

rpm_amd64 := amd64
rpm_386 := i386
rpm_s390x := s390x
rpm_ppc64le := ppc64le
rpm_arm5 := armel
rpm_arm6 := armv6hl
rpm_arm647 := aarch64
rpm_arch = $(rpm_$(GOARCH)$(GOARM))

.PHONY: $(rpms)
$(rpms):
	@$(MAKE) install
	@mkdir -p $(pkgdir)
	fpm --force \
		--log info \
		--architecture $(rpm_arch) \
		--input-type dir \
		--output-type rpm \
		--vendor InfluxData \
		--url https://github.com/marianob85/$(plugin_name) \
		--license MIT \
		--maintainer mariusz.brzeski@manobit.com \
		--config-files /etc/telegraf/$(plugin_name).config \
		--description "Plugin-driven server agent for reporting metrics into InfluxDB." \
		--depends coreutils \
		--depends shadow-utils \
		--name $(plugin_name) \
		--version $(version) \
		--iteration $(rpm_iteration) \
        --chdir $(DESTDIR) \
		--package $(pkgdir)/$@

deb_amd64 := amd64
deb_386 := i386
deb_s390x := s390x
deb_ppc64le := ppc64el
deb_arm5 := armel
deb_arm6 := armhf
deb_arm647 := arm64
deb_mips := mips
deb_mipsle := mipsel
deb_arch = $(deb_$(GOARCH)$(GOARM))

.PHONY: $(debs)
$(debs):
	@$(MAKE) install
	@mkdir -pv $(pkgdir)
	fpm --force \
		--log info \
		--architecture $(deb_arch) \
		--input-type dir \
		--output-type deb \
		--vendor InfluxData \
		--url https://github.com/marianob85/$(plugin_name) \
		--license MIT \
		--maintainer mariusz.brzeski@manobit.com \
		--config-files /etc/telegraf/$(plugin_name).config.sample \
		--description "Plugin-driven server agent for reporting metrics into InfluxDB." \
		--name $(plugin_name) \
		--version $(version) \
		--iteration $(deb_iteration) \
		--chdir $(DESTDIR) \
		--package $(pkgdir)/$@

.PHONY: $(zips)
$(zips):
	@$(MAKE) install
	@mkdir -p $(pkgdir)
	(cd $(dir $(DESTDIR)) && zip -r - ./*) > $(pkgdir)/$@

.PHONY: $(tars)
$(tars):
	@$(MAKE) install
	@mkdir -p $(pkgdir)
	tar --owner 0 --group 0 -czvf $(pkgdir)/$@ -C $(dir $(DESTDIR)) .

%armel.deb %armel.rpm %linux_armel.tar.gz: export GOOS := linux
%armel.deb %armel.rpm %linux_armel.tar.gz: export GOARCH := arm
%armel.deb %armel.rpm %linux_armel.tar.gz: export GOARM := 5

%armhf.deb %armv6hl.rpm %linux_armhf.tar.gz: export GOOS := linux
%armhf.deb %armv6hl.rpm %linux_armhf.tar.gz: export GOARCH := arm
%armhf.deb %armv6hl.rpm %linux_armhf.tar.gz: export GOARM := 6

%arm64.deb %aarch64.rpm %linux_arm64.tar.gz: export GOOS := linux
%arm64.deb %aarch64.rpm %linux_arm64.tar.gz: export GOARCH := arm64
%arm64.deb %aarch64.rpm %linux_arm64.tar.gz: export GOARM := 7

%.deb: export pkg := deb
%.deb: export prefix := /usr
%.deb: export conf_suffix := .sample
%.deb: export sysconfdir := /etc
%.rpm: export pkg := rpm
%.rpm: export prefix := /usr
%.rpm: export sysconfdir := /etc
%.tar.gz: export pkg := tar
%.tar.gz: export prefix := /usr
%.tar.gz: export sysconfdir := /etc
%.zip: export pkg := zip
%.zip: export prefix := /

%.deb %.rpm %.tar.gz %.zip: export DESTDIR = build/$(GOOS)-$(GOARCH)$(GOARM)$(cgo)-$(pkg)/$(plugin_name)-$(version)
%.deb %.rpm %.tar.gz %.zip: export buildbin = build/$(GOOS)-$(GOARCH)$(GOARM)$(cgo)/$(plugin_name)$(EXEEXT)
%.deb %.rpm %.tar.gz %.zip: export LDFLAGS = -w -s
