PKG = get.porter.sh/porter
SHELL = bash

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory

COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --match v* 2> /dev/null || echo v0)
PERMALINK ?= $(shell git describe --tags --exact-match --match v* &> /dev/null && echo latest || echo canary)

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
	go run mage.go -v BuildRuntime $(PKG) $(MIXIN) $(BINDIR)

build-client:
	go run mage.go -v BuildClient $(PKG) $(MIXIN) $(BINDIR)

xbuild-all:
	go run mage.go -v XBuildAll $(PKG) $(MIXIN) $(BINDIR)

