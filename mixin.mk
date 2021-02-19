PKG = get.porter.sh/porter
SHELL = bash

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory

COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git describe --tags --exact-match &> /dev/null && echo latest || echo canary)

LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
GO = GO111MODULE=on go
XBUILD = CGO_ENABLED=0 GO111MODULE=on $(GO) build -ldflags '$(LDFLAGS)'
BINDIR ?= bin/mixins/$(MIXIN)

CLIENT_PLATFORM ?= $(shell go env GOOS)
CLIENT_ARCH ?= $(shell go env GOARCH)
RUNTIME_PLATFORM ?= linux
RUNTIME_ARCH ?= amd64
# NOTE: When we add more to the build matrix, update the regex for porter mixins feed generate
SUPPORTED_PLATFORMS = linux darwin windows
SUPPORTED_ARCHES = amd64

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

.PHONY: build
build: build-client build-runtime

build-runtime:
	mkdir -p $(BINDIR)/runtimes
	GOARCH=$(RUNTIME_ARCH) GOOS=$(RUNTIME_PLATFORM) $(XBUILD) -o $(BINDIR)/runtimes/$(MIXIN)-runtime$(FILE_EXT) ./cmd/$(MIXIN)

build-client:
	mkdir -p $(BINDIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(MIXIN)$(FILE_EXT) ./cmd/$(MIXIN)

xbuild-all:
	$(foreach OS, $(SUPPORTED_PLATFORMS), \
		$(foreach ARCH, $(SUPPORTED_ARCHES), \
				$(MAKE) $(MAKE_OPTS) CLIENT_PLATFORM=$(OS) CLIENT_ARCH=$(ARCH) MIXIN=$(MIXIN) xbuild -f mixin.mk; \
		))
	cp -R $(BINDIR)/$(VERSION) $(BINDIR)/latest

xbuild: $(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT):
	mkdir -p $(dir $@)
	GOOS=$(CLIENT_PLATFORM) GOARCH=$(CLIENT_ARCH) $(XBUILD) -o $@ ./cmd/$(MIXIN)

clean:
	-rm -r $(BINDIR)/latest
	-rm -fr bin/mixins/$(MIXIN)
