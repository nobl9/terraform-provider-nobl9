.DEFAULT_GOAL := help
MAKEFLAGS += --silent --no-print-directory

TEST ?= $$(go list ./... | grep -v 'vendor')
HOSTNAME = nobl9.com
NAMESPACE = nobl9
NAME = nobl9
PROVIDER_ADDRESS = $(HOSTNAME)/$(NAMESPACE)/$(NAME)
BIN_DIR = ./bin
BINARY = $(BIN_DIR)/terraform-provider-$(NAME)
GIT_VERSION := $(shell git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null | sed 's/^v//')
VERSION ?= $(or $(GIT_VERSION),0.0.0)
VERSION_PKG := "$(shell go list -m)/internal/version"
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
REVISION ?= $(shell git rev-parse --short=8 HEAD)
LDFLAGS += -s -w \
	-X $(VERSION_PKG).BuildVersion=$(VERSION) \
	-X $(VERSION_PKG).BuildGitBranch=$(BRANCH) \
	-X $(VERSION_PKG).BuildGitRevision=$(REVISION)
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_INSTALL_DIR = $(HOME)/.terraform.d/plugins/$(PROVIDER_ADDRESS)/$(VERSION)/$(OS_ARCH)

# renovate datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v2.13.2

# Check if the program is present in $PATH and install otherwise.
# ${1} - oneOf{binary,yarn}
# ${2} - program name
define _ensure_installed
	LOCAL_BIN_DIR=$(BIN_DIR) ./scripts/ensure_installed.sh "${1}" "${2}"
endef

# Install Go binary using 'go install' with an output directory set via $GOBIN.
# ${1} - repository url
define _install_go_binary
	GOBIN=$(realpath $(BIN_DIR)) go install "${1}"
endef

# Print Makefile target step description.
# ${1} - step description
define _print_step
	printf -- '------\n%s...\n' "${1}"
endef

.PHONY: install/provider
## Install provider locally and print its Terraform declaration.
install/provider: build
	$(call _print_step,Installing provider $(VERSION) locally)
	mkdir -p "$(PROVIDER_INSTALL_DIR)"
	mv "$(BINARY)" "$(PROVIDER_INSTALL_DIR)"
	printf -- '%s\n' \
		'' \
		'Use this provider declaration in your Terraform configuration:' \
		'' \
		'terraform {' \
		'  required_providers {' \
		'    $(NAME) = {' \
		'      source  = "$(PROVIDER_ADDRESS)"' \
		'      version = "$(VERSION)"' \
		'    }' \
		'  }' \
		'}'

.PHONY: build
## Build provider binary.
build:
	$(call _print_step,Building provider binary)
	go build -ldflags="$(LDFLAGS)" -o $(BINARY)

.PHONY: test test/unit test/acc
## Run all tests.
test: test/unit test/acc

## Run Go unit tests.
test/unit:
	$(call _print_step,Running Go unit tests)
	go test -race -cover ./...

## Run Terraform acceptance tests.
test/acc:
	# Why? If we run acceptance tests without the terraform binary already installed,
	# the testing framework will try to install it for each test and it will lead to
	# system-level errors "text file busy".
	# See: https://github.com/hashicorp/terraform-plugin-testing/issues/429.
	$(call _print_step,Checking for Terraform binary)
	@which terraform > /dev/null 2>&1 || (echo "Error: terraform binary not found in PATH. Please install Terraform first." && exit 1)
	$(call _print_step,Running Terraform acceptance tests)
	TF_ACC=1 go test $(TEST) -ldflags="$(LDFLAGS)" -v $(TESTARGS) -timeout 60m nobl9/

.PHONY: release-dry-run
## Run Goreleaser in dry-run mode.
release-dry-run:
	$(call _print_step,Running Goreleaser in dry run mode)
	goreleaser release --snapshot --skip-publish --clean

.PHONY: check check/vet check/lint check/spell check/trailing check/markdown check/format check/generate
## Run all checks.
check: check/vet check/lint check/spell check/trailing check/markdown check/format check/generate

## Run 'go vet' on the whole project.
check/vet:
	$(call _print_step,Running go vet)
	go vet ./...

## Run golangci-lint all-in-one linter with configuration defined inside .golangci.yml.
check/lint:
	$(call _print_step,Running golangci-lint)
	$(call _ensure_installed,binary,golangci-lint)
	$(BIN_DIR)/golangci-lint run

## Check spelling, rules are defined in cspell.json.
check/spell:
	$(call _print_step,Verifying spelling)
	$(call _ensure_installed,yarn,cspell)
	yarn --silent cspell --no-progress '**/**'

## Check for trailing whitespaces in any of the projects' files.
check/trailing:
	$(call _print_step,Looking for trailing whitespaces)
	yarn --silent check-trailing-whitespaces

## Check markdown files for potential issues with markdownlint.
check/markdown:
	$(call _print_step,Verifying Markdown files)
	$(call _ensure_installed,yarn,markdownlint)
	yarn --silent markdownlint '**/*.md' -i node_modules -i docs

## Verify if the auto generated code has been committed.
check/generate:
	$(call _print_step,Checking if generated code matches the provided definitions)
	./scripts/check-generate.sh

## Verify if the files are formatted.
## You must first commit the changes, otherwise it won't detect the diffs.
check/format:
	$(call _print_step,Checking if files are formatted)
	./scripts/check-formatting.sh

.PHONY: generate generate/code
## Auto generate files.
generate: generate/code

## Generate Golang code.
generate/code:
	echo "Generating Go code..."
	go generate ./...

.PHONY: format format/go format/cspell
## Format files.
format: format/go format/cspell

## Format Go files.
format/go:
	echo "Formatting Go files..."
	$(call _ensure_installed,binary,golangci-lint)
	$(BIN_DIR)/golangci-lint fmt

## Format cspell config file.
format/cspell:
	echo "Formatting cspell.yaml configuration (words list)..."
	$(call _ensure_installed,yarn,yaml)
	yarn --silent format-cspell-config

.PHONY: install install/yarn install/golangci-lint
## Install all dev dependencies.
install: install/yarn install/golangci-lint

## Install JS dependencies with yarn.
install/yarn:
	echo "Installing yarn dependencies..."
	yarn --silent install

## Install golangci-lint (https://golangci-lint.run).
install/golangci-lint:
	echo "Installing golangci-lint..."
	curl -sSfL https://golangci-lint.run/install.sh |\
 		sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)

.PHONY: help
## Print this help message.
help:
	./scripts/makefile-help.awk $(MAKEFILE_LIST)
