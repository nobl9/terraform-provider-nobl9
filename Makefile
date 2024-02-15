.DEFAULT_GOAL := help
MAKEFLAGS += --silent --no-print-directory

TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=nobl9.com
NAMESPACE=nobl9
NAME=nobl9
BIN_DIR=./bin
BINARY=$(BIN_DIR)/terraform-provider-$(NAME)
VERSION=0.23.0-beta
BUILD_FLAGS="-X github.com/nobl9/terraform-provider-nobl9/nobl9.Version=$(VERSION)"
OS_ARCH?=linux_amd64

# renovate datasource=github-releases depName=securego/gosec
GOSEC_VERSION := v2.18.2
# renovate datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v1.55.2
# renovate datasource=go depName=golang.org/x/vuln/cmd/govulncheck
GOVULNCHECK_VERSION := v1.0.4
# renovate datasource=go depName=golang.org/x/tools/cmd/goimports
GOIMPORTS_VERSION := v0.18.0
# renovate datasource=github-releases depName=segmentio/golines
GOLINES_VERSION := v0.9.0

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

# Print Makefile target step description for check.
# Only print 'check' steps this way, and not dependent steps, like 'install'.
# ${1} - step description
define _print_check_step
	printf -- '------\n%s...\n' "${1}"
endef

.PHONY: install
## Install provider locally.
install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

.PHONY: build
## Build provider binary.
build:
	go build -ldflags $(BUILD_FLAGS) -o $(BINARY)

.PHONY: test
## Run Go unit tests.
test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

.PHONY: testacc
## Run acceptance tests.
testacc:
	cd nobl9 && TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

.PHONY: release-dry-run
## Run Goreleaser in dry-run mode.
release-dry-run:
	goreleaser release --snapshot --skip-publish --rm-dist

.PHONY: check check/vet check/lint check/gosec check/spell check/trailing check/markdown check/format check/generate check/vulns
## Run all checks.
check: check/vet check/lint check/gosec check/spell check/trailing check/markdown check/format check/generate check/vulns

## Run 'go vet' on the whole project.
check/vet:
	$(call _print_check_step,Running go vet)
	go vet ./...

## Run golangci-lint all-in-one linter with configuration defined inside .golangci.yml.
check/lint:
	$(call _print_check_step,Running golangci-lint)
	$(call _ensure_installed,binary,golangci-lint)
	$(BIN_DIR)/golangci-lint run

## Check for security problems using gosec, which inspects the Go code by scanning the AST.
check/gosec:
	$(call _print_check_step,Running gosec)
	$(call _ensure_installed,binary,gosec)
	$(BIN_DIR)/gosec -exclude-generated -quiet ./...

## Check spelling, rules are defined in cspell.json.
check/spell:
	$(call _print_check_step,Verifying spelling)
	$(call _ensure_installed,yarn,cspell)
	yarn --silent cspell --no-progress '**/**'

## Check for trailing whitespaces in any of the projects' files.
check/trailing:
	$(call _print_check_step,Looking for trailing whitespaces)
	yarn --silent check-trailing-whitespaces

## Check markdown files for potential issues with markdownlint.
check/markdown:
	$(call _print_check_step,Verifying Markdown files)
	$(call _ensure_installed,yarn,markdownlint)
	yarn --silent markdownlint '*.md' --disable MD010 # MD010 does not handle code blocks well.

## Check for potential vulnerabilities across all Go dependencies.
check/vulns:
	$(call _print_check_step,Running govulncheck)
	$(call _ensure_installed,binary,govulncheck)
	$(BIN_DIR)/govulncheck ./...

## Verify if the auto generated code has been committed.
check/generate:
	$(call _print_check_step,Checking if generated code matches the provided definitions)
	./scripts/check-generate.sh

## Verify if the files are formatted.
## You must first commit the changes, otherwise it won't detect the diffs.
check/format:
	$(call _print_check_step,Checking if files are formatted)
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
	$(call _ensure_installed,binary,goimports)
	$(call _ensure_installed,binary,golines)
	go fmt ./...
	$(BIN_DIR)/goimports -local=github.com/nobl9/terraform-provider-nobl9 -w .
	$(BIN_DIR)/golines --ignore-generated -m 120 -w .

## Format cspell config file.
format/cspell:
	echo "Formatting cspell.yaml configuration (words list)..."
	$(call _ensure_installed,yarn,yaml)
	yarn --silent format-cspell-config

.PHONY: install install/yarn install/golangci-lint install/gosec install/govulncheck install/goimports
## Install all dev dependencies.
install: install/yarn install/golangci-lint install/gosec install/govulncheck install/goimports

## Install JS dependencies with yarn.
install/yarn:
	echo "Installing yarn dependencies..."
	yarn --silent install

## Install golangci-lint (https://golangci-lint.run).
install/golangci-lint:
	echo "Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh |\
 		sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)

## Install gosec (https://github.com/securego/gosec).
install/gosec:
	echo "Installing gosec..."
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh |\
 		sh -s -- -b $(BIN_DIR) $(GOSEC_VERSION)

## Install govulncheck (https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck).
install/govulncheck:
	echo "Installing govulncheck..."
	$(call _install_go_binary,golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION))

## Install goimports (https://pkg.go.dev/golang.org/x/tools/cmd/goimports).
install/goimports:
	echo "Installing goimports..."
	$(call _install_go_binary,golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION))

## Install golines (https://github.com/segmentio/golines).
install/golines:
	echo "Installing golines..."
	$(call _install_go_binary,github.com/segmentio/golines@$(GOLINES_VERSION))

.PHONY: help
## Print this help message.
help:
	./scripts/makefile-help.awk $(MAKEFILE_LIST)
