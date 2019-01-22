SHELL = bash

REGISTRY ?= $(USER)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

KUBECONFIG ?= $(HOME)/.kube/config
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

HELM_MIXIN_URL = https://deislabs.blob.core.windows.net/porter/mixins/helm/latest/helm
AZURE_MIXIN_URL = https://deislabs.blob.core.windows.net/porter/mixins/azure/latest/azure

build: build-client build-runtime azure helm

build-runtime:
	$(MAKE) build-runtime MIXIN=porter -f mixin.mk
	$(MAKE) build-runtime MIXIN=exec -f mixin.mk

build-client:
	$(MAKE) build-client MIXIN=porter -f mixin.mk
	$(MAKE) build-client MIXIN=exec -f mixin.mk
	cp bin/mixins/porter/porter$(FILE_EXT) bin/
	cp -R templates bin/

xbuild-all:
	$(MAKE) xbuild-all MIXIN=porter -f mixin.mk
	$(MAKE) xbuild-all MIXIN=exec -f mixin.mk
	cp -R templates bin/

xbuild-runtime:
	$(MAKE) xbuild-runtime MIXIN=porter -f mixin.mk
	$(MAKE) xbuild-runtime MIXIN=exec -f mixin.mk

xbuild-client:
	$(MAKE) xbuild-client MIXIN=porter -f mixin.mk
	$(MAKE) xbuild-client MIXIN=exec -f mixin.mk

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
	# commented out because azure version is borked
	#bin/mixins/azure/azure version

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

publish:
	$(MAKE) publish MIXIN=exec -f mixin.mk
	# AZURE_STORAGE_CONNECTION_STRING will be used for auth in the following commands
	if [[ "$(PERMALINK)" == "latest" ]]; then \
	az storage blob upload-batch -d porter/$(VERSION) -s bin/mixins/porter/$(VERSION); \
	az storage blob upload-batch -d porter/$(VERSION)/templates -s templates; \
	az storage blob upload-batch -d porter/$(VERSION) -s scripts/install; \
	fi
	az storage blob upload-batch -d porter/$(PERMALINK) -s bin/mixins/porter/$(VERSION)
	az storage blob upload-batch -d porter/$(PERMALINK)/templates -s templates
	az storage blob upload-batch -d porter/$(PERMALINK) -s scripts/install

clean:
	-rm -fr bin/
	-rm -fr cnab/
	-rm Dockerfile porter.yaml
	-duffle uninstall PORTER-HELLO
	-duffle uninstall PORTER-WORDPRESS --credentials ci
	-helm delete --purge porter-ci-mysql
	-helm delete --purge porter-ci-wordpress
