SHELL = bash
PKG = get.porter.sh/porter

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory
PORTER_REGISTRY ?= localhost:5000

CLIENT_PLATFORM = $(shell go env GOOS)
CLIENT_ARCH = $(shell go env GOARCH)
CLIENT_GOPATH = $(shell go env GOPATH)
RUNTIME_PLATFORM = linux
RUNTIME_ARCH = amd64
PORTER_UPDATE_TEST_FILES ?=

GO = go
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
build:
	go run mage.go -v Build

test:
	go run mage.go -v Test

test-unit:
	go run mage.go -v TestUnit

test-integration:
	go run mage.go -v TestIntegration

test-smoke:
	go run mage.go -v TestSmoke

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

install:
	go run mage.go install

setup-dco:
	@scripts/setup-dco/setup.sh

clean:
	go run mage.go clean


