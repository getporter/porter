SHELL = bash

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory

REGISTRY ?= $(USER)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

KUBECONFIG  ?= $(HOME)/.kube/config
DUFFLE_HOME ?= bin/.duffle
PORTER_HOME ?= bin

CLIENT_PLATFORM = $(shell go env GOOS)
CLIENT_ARCH = $(shell go env GOARCH)
RUNTIME_PLATFORM = linux
RUNTIME_ARCH = amd64
BASEURL_FLAG ?= 

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

MIXIN_TAG ?= canary
HELM_MIXIN_URL = https://deislabs.blob.core.windows.net/porter/mixins/helm/$(MIXIN_TAG)/helm
AZURE_MIXIN_URL = https://deislabs.blob.core.windows.net/porter/mixins/azure/$(MIXIN_TAG)/azure

build: build-client build-runtime azure helm

build-runtime: generate
	$(MAKE) $(MAKE_OPTS) build-runtime MIXIN=porter -f mixin.mk BINDIR=bin
	$(MAKE) $(MAKE_OPTS) build-runtime MIXIN=exec -f mixin.mk
	$(MAKE) $(MAKE_OPTS) build-runtime MIXIN=kubernetes -f mixin.mk

build-client: generate
	$(MAKE) $(MAKE_OPTS) build-client MIXIN=porter -f mixin.mk BINDIR=bin
	$(MAKE) $(MAKE_OPTS) build-client MIXIN=exec -f mixin.mk
	$(MAKE) $(MAKE_OPTS) build-client MIXIN=kubernetes -f mixin.mk

generate: packr2
	go generate ./...

HAS_PACKR2 := $(shell command -v packr2)
packr2:
ifndef HAS_PACKR2
	go get -u github.com/gobuffalo/packr/v2/packr2
endif

xbuild-all:
	$(MAKE) $(MAKE_OPTS) xbuild-all MIXIN=porter -f mixin.mk BINDIR=bin
	$(MAKE) $(MAKE_OPTS) xbuild-all MIXIN=exec -f mixin.mk
	$(MAKE) $(MAKE_OPTS) xbuild-all MIXIN=kubernetes -f mixin.mk

xbuild-runtime:
	$(MAKE) $(MAKE_OPTS) xbuild-runtime MIXIN=porter -f mixin.mk BINDIR=bin
	$(MAKE) $(MAKE_OPTS) xbuild-runtime MIXIN=exec -f mixin.mk
	$(MAKE) $(MAKE_OPTS) xbuild-runtime MIXIN=kubernetes -f mixin.mk

xbuild-client:
	$(MAKE) $(MAKE_OPTS) xbuild-client MIXIN=porter -f mixin.mk BINDIR=bin
	$(MAKE) $(MAKE_OPTS) xbuild-client MIXIN=exec -f mixin.mk
	$(MAKE) $(MAKE_OPTS) xbuild-client MIXIN=kubernetes -f mixin.mk

bin/mixins/helm/helm:
	mkdir -p bin/mixins/helm
	curl -f -o bin/mixins/helm/helm $(HELM_MIXIN_URL)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)
	chmod +x bin/mixins/helm/helm
	bin/mixins/helm/helm version

bin/mixins/helm/helm-runtime:
	mkdir -p bin/mixins/helm
	curl -f -o bin/mixins/helm/helm-runtime $(HELM_MIXIN_URL)-runtime-$(RUNTIME_PLATFORM)-$(RUNTIME_ARCH)
	chmod +x bin/mixins/helm/helm-runtime

helm: bin/mixins/helm/helm bin/mixins/helm/helm-runtime

bin/mixins/azure/azure:
	mkdir -p bin/mixins/azure
	curl -f -o bin/mixins/azure/azure $(AZURE_MIXIN_URL)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)
	chmod +x bin/mixins/azure/azure
	bin/mixins/azure/azure version

bin/mixins/azure/azure-runtime:
	mkdir -p bin/mixins/azure
	curl -f -o bin/mixins/azure/azure-runtime $(AZURE_MIXIN_URL)-runtime-$(RUNTIME_PLATFORM)-$(RUNTIME_ARCH)
	chmod +x bin/mixins/azure/azure-runtime

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
	hugo --source docs/ $(BASEURL_FLAG)

docs-preview:
	hugo serve --source docs/

prep-install-scripts:
	mkdir -p bin/$(VERSION)
	sed 's|UNKNOWN|$(PERMALINK)|g' scripts/install/install-mac.sh > bin/$(VERSION)/install-mac.sh
	sed 's|UNKNOWN|$(PERMALINK)|g' scripts/install/install-linux.sh > bin/$(VERSION)/install-linux.sh
	sed 's|UNKNOWN|$(PERMALINK)|g' scripts/install/install-windows.ps1 > bin/$(VERSION)/install-windows.ps1

publish: prep-install-scripts
	$(MAKE) $(MAKE_OPTS) publish MIXIN=exec -f mixin.mk
	$(MAKE) $(MAKE_OPTS) publish MIXIN=kubernetes -f mixin.mk
	# AZURE_STORAGE_CONNECTION_STRING will be used for auth in the following commands
	if [[ "$(PERMALINK)" == "latest" ]]; then \
	az storage blob upload-batch -d porter/$(VERSION) -s bin/$(VERSION); \
	fi
	az storage blob upload-batch -d porter/$(PERMALINK) -s bin/$(VERSION)

install: build
	mkdir -p $(HOME)/.porter
	cp -R bin/* $(HOME)/.porter/
	ln -f -s $(HOME)/.porter/porter /usr/local/bin/porter

clean:
	-rm -fr bin/
	-rm -fr cnab/
	-rm Dockerfile porter.yaml
	-duffle uninstall PORTER-HELLO
	-duffle uninstall PORTER-WORDPRESS --credentials ci
	-helm delete --purge porter-ci-mysql
	-helm delete --purge porter-ci-wordpress
