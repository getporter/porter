COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --dirty='+dev' --abbrev=0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

PKG = github.com/deislabs/porter
LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
XBUILD = GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags '$(LDFLAGS)'

REGISTRY ?= $(USER)

ifeq ($(OS),Windows_NT)
	TARGET = $(PROJECT).exe
	SHELL  = cmd.exe
	CHECK  = where.exe
else
	TARGET = $(PROJECT)
	SHELL  = bash
	CHECK  = command -v
endif

build: porter exec
	cp -R templates bin/

porter:
	$(XBUILD) -o bin/porter ./cmd/porter
	GOOS=linux $(XBUILD) -o bin/porter-runtime ./cmd/porter
	mkdir -p bin/mixins/porter
	cp bin/porter* bin/mixins/porter/

exec:
	mkdir -p bin/mixins/exec
	$(XBUILD) -o bin/mixins/exec/exec ./cmd/exec
	GOOS=linux $(XBUILD) -o bin/mixins/exec/exec-runtime ./cmd/exec

test: test-unit test-cli

test-unit: build
	go test ./...

test-cli: clean build
	./bin/porter help
	./bin/porter version
	./bin/porter init
	sed -i 's/porter-hello:latest/$(REGISTRY)\/porter-hello:latest/g' porter.yaml
	./bin/porter build
	duffle install PORTER-HELLO -f bundle.json --insecure

.PHONY: docs
docs:
	hugo --source docs/

docs-preview:
	hugo serve --source docs/

clean:
	-rm -fr bin/
	-rm -fr cnab/
	-rm Dockerfile porter.yaml
	-duffle uninstall PORTER-HELLO

.PHONY: lint
lint:
	golangci-lint run --config ./golangci.yml

HAS_DEP          := $(shell $(CHECK) dep)
HAS_GOLANGCI     := $(shell $(CHECK) golangci-lint)

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
ifndef HAS_GOLANGCI
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin
endif
	dep ensure -vendor-only -v