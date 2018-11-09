COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --dirty='+dev' --abbrev=0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

PKG = github.com/deislabs/porter
LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
XBUILD = CGO_ENABLED=0 go build -a -tags netgo -ldflags '$(LDFLAGS)'

build: porter exec
	cp -R templates bin/

porter:
	$(XBUILD) -o bin/porter ./cmd/porter

exec:
	mkdir -p bin/mixins/exec
	$(XBUILD) -o bin/mixins/exec/exec ./cmd/exec

test: build
	go test ./...
	./bin/porter version
	./bin/porter help
	./bin/porter run --action install --file templates/porter.yaml

.PHONY: docs
docs:
	hugo --source docs/

docs-preview:
	hugo serve --source docs/
