MIXIN = porter

COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --dirty='+dev' --abbrev=0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

PKG = github.com/deislabs/porter
LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
XBUILD = GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags '$(LDFLAGS)'
BINDIR = bin/mixins/$(MIXIN)

CLIENT_PLATFORM = $(shell go env GOOS)
CLIENT_ARCH = $(shell go env GOARCH)
RUNTIME_PLATFORM = linux
RUNTIME_ARCH = amd64
SUPPORTED_CLIENT_PLATFORMS = linux darwin windows
SUPPORTED_CLIENT_ARCHES = amd64 386

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

REGISTRY ?= $(USER)

KUBECONFIG ?= $(HOME)/.kube/config
DUFFLE_HOME ?= bin/.duffle
PORTER_HOME ?= bin

HELM_MIXIN_URL = https://deislabs.blob.core.windows.net/porter/mixins/helm/v0.1.0-ralpha.1+aviation/helm
AZURE_MIXIN_URL = https://deislabs.blob.core.windows.net/porter/mixins/azure/v0.1.0-ralpha.1+aviation/azure

build: build-client build-runtime azure helm

build-runtime:
	mkdir -p $(BINDIR)
	GOARCH=$(RUNTIME_ARCH) GOOS=$(RUNTIME_PLATFORM) go build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(MIXIN)-runtime$(FILE_EXT) ./cmd/$(MIXIN)
	cp $(BINDIR)/$(MIXIN)-runtime$(FILE_EXT) bin/

build-client:
	mkdir -p $(BINDIR)
	go build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(MIXIN)$(FILE_EXT) ./cmd/$(MIXIN)
	cp $(BINDIR)/$(MIXIN)$(FILE_EXT) bin/
	cp -R templates bin/

build-all: xbuild-runtime $(addprefix build-for-,$(SUPPORTED_CLIENT_PLATFORMS))
	cp -R templates bin/

build-for-%:
	$(MAKE) CLIENT_PLATFORM=$* xbuild-client

xbuild-runtime:
	GOARCH=$(RUNTIME_ARCH) GOOS=$(RUNTIME_PLATFORM) $(XBUILD) -o $(BINDIR)/$(VERSION)/$(MIXIN)-runtime-$(RUNTIME_PLATFORM)-$(RUNTIME_ARCH)$(FILE_EXT) ./cmd/$(MIXIN)

xbuild-client: $(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT):
	mkdir -p $(dir $@)
	$(XBUILD) -o $@ ./cmd/$(MIXIN)

exec:
	mkdir -p bin/mixins/exec
	$(XBUILD) -o bin/mixins/exec/exec ./cmd/exec
	GOOS=linux $(XBUILD) -o bin/mixins/exec/exec-runtime ./cmd/exec

bin/mixins/helm/helm:
	mkdir -p bin/mixins/helm
	curl -o bin/mixins/helm/helm $(HELM_MIXIN_URL)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)

bin/mixins/helm/helm-runtime:
	mkdir -p bin/mixins/helm
	curl -o bin/mixins/helm/helm-runtime $(HELM_MIXIN_URL)-runtime-$(RUNTIME_PLATFORM)-$(RUNTIME_ARCH)

helm: bin/mixins/helm/helm bin/mixins/helm/helm-runtime

bin/mixins/azure/azure:
	mkdir -p bin/mixins/azure
	curl -o bin/mixins/azure/azure $(AZURE_MIXIN_URL)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)

bin/mixins/azure/azure-runtime:
	mkdir -p bin/mixins/azure
	curl -o bin/mixins/azure/azure-runtime $(AZURE_MIXIN_URL)-runtime-$(RUNTIME_PLATFORM)-$(RUNTIME_ARCH)

azure: bin/mixins/azure/azure bin/mixins/azure/azure-runtime

test: clean test-unit test-cli

test-unit: build
	go test ./...

test-cli: clean build init-duffle-home-for-ci init-porter-home-for-ci
	export KUBECONFIG
	export PORTER_HOME
	export DUFFLE_HOME

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
	duffle install PORTER-WORDPRESS -f bundle.json --credentials ci --insecure --home $(DUFFLE_HOME)

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
