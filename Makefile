SHELL = bash

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory

REGISTRY ?= $(USER)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)

KUBECONFIG  ?= $(HOME)/.kube/config
PORTER_HOME = bin

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

INT_MIXINS = exec kubernetes
EXT_MIXINS = helm azure terraform
MIXIN_TAG ?= canary
MIXINS_URL = https://cdn.deislabs.io/porter/mixins

.PHONY: build
build: build-porter build-mixins clean-packr get-mixins

build-porter: generate
	$(MAKE) $(MAKE_OPTS) build MIXIN=porter -f mixin.mk BINDIR=bin

build-porter-client: generate
	$(MAKE) $(MAKE_OPTS) build-client MIXIN=porter -f mixin.mk BINDIR=bin

build-mixins: $(addprefix build-mixin-,$(INT_MIXINS))
build-mixin-%: generate
	$(MAKE) $(MAKE_OPTS) build MIXIN=$* -f mixin.mk

generate: packr2
	go generate ./...

HAS_PACKR2 := $(shell command -v packr2)
packr2:
ifndef HAS_PACKR2
	go get -u github.com/gobuffalo/packr/v2/packr2
endif

xbuild-all: xbuild-porter xbuild-mixins

xbuild-porter: generate
	$(MAKE) $(MAKE_OPTS) xbuild-all MIXIN=porter -f mixin.mk BINDIR=bin

xbuild-mixins: $(addprefix xbuild-mixin-,$(INT_MIXINS))
xbuild-mixin-%: generate
	$(MAKE) $(MAKE_OPTS) xbuild-all MIXIN=$* -f mixin.mk

get-mixins:
	$(foreach MIXIN, $(EXT_MIXINS), \
		bin/porter mixin install $(MIXIN) --version $(MIXIN_TAG) --url $(MIXINS_URL)/$(MIXIN); \
	)

verify: verify-vendor

verify-vendor: clean-packr dep
	dep check

HAS_DEP := $(shell command -v dep)
dep:
ifndef HAS_DEP
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
	dep version

test: clean-last-testrun test-unit test-integration test-cli

test-unit: build
	go test ./...

test-integration: build
	go test -v -timeout 20m -tags=integration ./tests

test-cli: clean-last-testrun build init-porter-home-for-ci
	REGISTRY=$(REGISTRY) KUBECONFIG=$(KUBECONFIG) ./scripts/test/test-cli.sh

init-porter-home-for-ci:
	cp -R build/testdata/credentials $(PORTER_HOME)
	sed -i 's|KUBECONFIGPATH|$(KUBECONFIG)|g' $(PORTER_HOME)/credentials/ci.yaml
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

	# Generate the mixin feed
	az storage blob download -c porter -n atom.xml -f bin/atom.xml
	bin/porter mixins feed generate -d bin/mixins -f bin/atom.xml -t build/atom-template.xml
	az storage blob upload -c porter -n atom.xml -f bin/atom.xml

# all-bundles loops through all items under the dir provided by the first argument
# and if the item is a sub-directory containing a porter.yaml file,
# runs the make target(s) provided by the second argument
define all-bundles
	@for dir in $$(ls -1 $(1)); do \
		if [[ -e "$(1)/$$dir/porter.yaml" ]]; then \
			BUNDLE=$$dir make $(MAKE_OPTS) $(2) ; \
			exit_code=$$? ; \
			if [[ $$exit_code -ne 0 ]] ; then \
				exit $$exit_code ; \
			fi ; \
		fi ; \
	done
endef

EXAMPLES_DIR := examples

.PHONY: build-bundle
build-bundle:
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),build-bundle)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && porter build
endif

.PHONY: publish-bundle
publish-bundle:
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),publish-bundle)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && porter publish
endif

BUNDLE_SCHEMA_PATH ?= /tmp/bundle.schema.json

.PHONY: validate-bundle
validate-bundle: fetch-bundle-schema
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),validate-bundle)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && ajv test -s $(BUNDLE_SCHEMA_PATH) -d .cnab/bundle.json --valid
endif

fetch-bundle-schema:
	@curl --fail --silent --show-error -o $(BUNDLE_SCHEMA_PATH) \
		https://raw.githubusercontent.com/deislabs/cnab-spec/master/schema/bundle.schema.json

install:
	mkdir -p $(HOME)/.porter
	cp -R bin/mixins $(HOME)/.porter/
	cp bin/porter* $(HOME)/.porter/
	ln -f -s $(HOME)/.porter/porter /usr/local/bin/porter

clean: clean-mixins clean-last-testrun

clean-mixins:
	-rm -fr bin/

clean-last-testrun:
	-rm -fr cnab/ porter.yaml Dockerfile bundle.json

clean-packr: packr2
	cd pkg/porter && packr2 clean
	$(foreach MIXIN, $(INT_MIXINS), \
		`cd pkg/$(MIXIN) && packr2 clean`; \
	)
