SHELL = bash
PKG = get.porter.sh/porter

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory
REGISTRY ?= getporterci

CLIENT_PLATFORM = $(shell go env GOOS)
CLIENT_ARCH = $(shell go env GOARCH)
CLIENT_GOPATH = $(shell go env GOPATH)
RUNTIME_PLATFORM = linux
RUNTIME_ARCH = amd64
BASEURL_FLAG ?=
PORTER_UPDATE_TEST_FILES ?=

GO = GO111MODULE=on go
LOCAL_PORTER = PORTER_HOME=$(PWD)/bin $(PWD)/bin/porter

# Add ~/go/bin to PATH, works for everything _except_ shell commands
HAS_GOBIN_IN_PATH := $(shell re='(:|^)$(CLIENT_GOPATH)/bin/?(:|$$)'; if [[ "$${PATH}" =~ $${re} ]];then echo $${GOPATH}/bin;fi)
ifndef HAS_GOBIN_IN_PATH
export PATH := ${CLIENT_GOPATH}/bin:${PATH}
endif

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

INT_MIXINS = exec

.PHONY: build
build: build-porter docs-gen build-mixins clean-packr get-mixins

build-porter: generate
	go run mage.go -v BuildAll $(PKG) porter bin

build-porter-client: generate
	go run mage.go -v BuildClient $(PKG) porter bin
	$(MAKE) $(MAKE_OPTS) clean-packr

build-mixins: generate
	go run mage.go -v BuildAll $(PKG) exec bin/mixins/exec

generate: packr2
	$(GO) mod tidy
	$(GO) generate ./...

HAS_PACKR2 := $(shell command -v packr2)
HAS_GOBIN_IN_PATH := $(shell re='(:|^)$(CLIENT_GOPATH)/bin/?(:|$$)'; if [[ "$${PATH}" =~ $${re} ]];then echo $${GOPATH}/bin;fi)
packr2:
ifndef HAS_PACKR2
ifndef HAS_GOBIN_IN_PATH
	$(error "$(CLIENT_GOPATH)/bin is not in path and packr2 is not installed. Install packr2 or add "$(CLIENT_GOPATH)/bin to your path")
endif
	curl -SLo /tmp/packr.tar.gz https://github.com/gobuffalo/packr/releases/download/v2.6.0/packr_2.6.0_$(CLIENT_PLATFORM)_$(CLIENT_ARCH).tar.gz
	cd /tmp && tar -xzf /tmp/packr.tar.gz
	install /tmp/packr2 $(CLIENT_GOPATH)/bin/
endif

xbuild-all: xbuild-porter xbuild-mixins

xbuild-porter: generate
	go run mage.go -v XBuildAll $(PKG) porter bin

xbuild-mixins: generate
	go run mage.go -v XBuildAll $(PKG) exec bin/mixins/exec

get-mixins:
	go run mage.go GetMixins

verify:
	@echo 'verify does nothing for now but keeping it as a placeholder for a bit'

test: build test-unit test-integration test-smoke

test-unit:
	PORTER_UPDATE_TEST_FILES=$(PORTER_UPDATE_TEST_FILES) $(GO) test ./...

test-integration:
	go run mage.go TestIntegration

test-smoke:
	go run mage.go testSmoke

.PHONY: docs
docs:
	hugo --source docs/ $(BASEURL_FLAG)

docs-gen:
	$(GO) run --tags=docs ./cmd/porter docs

docs-preview: docs-stop-preview
	@docker run -d -v $$PWD:/src -p 1313:1313 --name porter-docs -w /src/docs \
	klakegg/hugo:0.78.1-ext-alpine server -D -F --noHTTPCache --watch --bind=0.0.0.0
	# Wait for the documentation web server to finish rendering
	@until docker logs porter-docs | grep -m 1  "Web Server is available"; do : ; done
	@open "http://localhost:1313/docs/"

docs-stop-preview:
	@docker rm -f porter-docs &> /dev/null || true

publish: publish-bin publish-mixins publish-images

publish-bin:
	go run mage.go -v PublishPorter

publish-mixins:
	go run mage.go -v PublishMixinFeed exec

.PHONY: build-images
build-images:
	go run mage.go -v BuildImages

.PHONY: publish-images
publish-images:
	go run mage.go -v PublishImages

start-local-docker-registry:
	@docker run -d -p 5000:5000 --name registry registry:2

stop-local-docker-registry:
	@if $$(docker inspect registry > /dev/null 2>&1); then \
		docker rm -f registry ; \
	fi

# all-bundles loops through all items under the dir provided by the first argument
# and if the item is a sub-directory containing a porter.yaml file,
# runs the make target(s) provided by the second argument
define all-bundles
	@for dir in $$(ls -1 $(1)); do \
		if [[ -e "$(1)/$$dir/porter.yaml" ]]; then \
			BUNDLE=$$dir make $(MAKE_OPTS) $(2) || exit $$? ; \
		fi ; \
	done
endef

EXAMPLES_DIR := examples

.PHONY: lint-examples
lint-examples:
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),lint-examples)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && $(LOCAL_PORTER) lint
endif

.PHONY: build-examples
build-examples:
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),build-examples)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && $(LOCAL_PORTER) build
endif

.PHONY: publish-examples
publish-examples:
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),publish-examples)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && $(LOCAL_PORTER) publish --registry $(REGISTRY)
endif

SCHEMA_VERSION     := cnab-core-1.0.1
BUNDLE_SCHEMA      := bundle.schema.json
DEFINITIONS_SCHEMA := definitions.schema.json

define fetch-schema
	@curl -L --fail --silent --show-error -o /tmp/$(1) \
		https://cnab.io/schema/$(SCHEMA_VERSION)/$(1)
endef

fetch-schemas: fetch-bundle-schema fetch-definitions-schema

fetch-bundle-schema:
	$(call fetch-schema,$(BUNDLE_SCHEMA))

fetch-definitions-schema:
	$(call fetch-schema,$(DEFINITIONS_SCHEMA))

HAS_AJV := $(shell command -v ajv)
ajv:
ifndef HAS_AJV
	npm install -g ajv-cli@3.3.0
endif

.PHONY: validate-bundle
validate-examples: fetch-schemas ajv
ifndef BUNDLE
	$(call all-bundles,$(EXAMPLES_DIR),validate-examples)
else
	cd $(EXAMPLES_DIR)/$(BUNDLE) && \
		ajv test -s /tmp/$(BUNDLE_SCHEMA) -r /tmp/$(DEFINITIONS_SCHEMA) -d .cnab/bundle.json --valid
endif

install:
	go run mage.go install

setup-dco:
	@scripts/setup-dco/setup.sh

clean:
	go run mage.go clean

clean-packr: packr2
	cd cmd/porter && packr2 clean
	cd pkg/porter && packr2 clean
	cd pkg/pkgmgmt/feed && packr2 clean
	$(foreach MIXIN, $(INT_MIXINS), \
		`cd pkg/$(MIXIN) && packr2 clean`; \
	)
