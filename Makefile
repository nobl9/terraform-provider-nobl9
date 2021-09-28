TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=nobl9.com
NAMESPACE=nobl9
NAME=nobl9
BINARY=terraform-provider-${NAME}
VERSION=0.0.1
OS_ARCH=darwin_amd64

default: install

build:
	go build -ldflags "-X github.com/nobl9/terraform-provider-nobl9/nobl9.Version=$(VERSION)-rc" -o ${BINARY}

release-dry-run:
	goreleaser release --snapshot --skip-publish --rm-dist

release:
	goreleaser release

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

