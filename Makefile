COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --dirty='+dev' --abbrev=0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

PKG = github.com/deislabs/porter
LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
XBUILD = GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags '$(LDFLAGS)'

REGISTRY ?= $(USER)

KUBECONFIG ?= $(HOME)/.kube/config
DUFFLE_HOME ?= bin/.duffle
PORTER_HOME ?= bin

.EXPORT_ALL_VARIABLES:

build: porter exec helm
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

helm:
	mkdir -p bin/mixins/helm
	$(XBUILD) -o bin/mixins/helm/helm ./cmd/helm
	GOOS=linux $(XBUILD) -o bin/mixins/helm/helm-runtime ./cmd/helm

test: clean test-unit test-cli

test-unit: build
	go test ./...

test-cli: clean build init-duffle-home-for-ci init-porter-home-for-ci
	./bin/porter help
	./bin/porter version

	# Verify our default template bundle
	./bin/porter create
	sed -i 's/porter-hello:latest/$(REGISTRY)\/porter-hello:latest/g' porter.yaml
	./bin/porter build
	duffle install PORTER-HELLO -f bundle.json --insecure

	# Verify a bundle with dependencies
	cp build/testdata/bundles/wordpress/porter.yaml .
	sed -i 's/porter-wordpress:latest/$(REGISTRY)\/porter-wordpress:latest/g' porter.yaml
	./bin/porter build
	duffle install PORTER-WORDPRESS -f bundle.json --credentials ci --insecure

init-duffle-home-for-ci:
	duffle init --home $(DUFFLE_HOME)
	cp -R build/testdata/credentials $(DUFFLE_HOME)
	sed -i 's|KUBECONFIGPATH|$(KUBECONFIG)|g' $(DUFFLE_HOME)/credentials/ci.yaml

init-porter-home-for-ci:
	#porter init
	cp -R build/testdata/bundles $(PORTER_HOME)

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
	-duffle uninstall PORTER-WORDPRESS --credentials ci
	-helm delete --purge porter-ci-mysql
	-helm delete --purge porter-ci-wordpress
