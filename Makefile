TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=nobl9.com
NAMESPACE=nobl9
NAME=nobl9
BINARY=terraform-provider-${NAME}
VERSION=0.3.0
BUILD_FLAGS="-X github.com/nobl9/terraform-provider-nobl9/nobl9.Version=$(VERSION)"
OS_ARCH?=linux_amd64

default: install

build:
	go build -ldflags $(BUILD_FLAGS) -o ${BINARY}

release-dry-run:
	goreleaser release --snapshot --skip-publish --rm-dist

release:
	GOOS=darwin GOARCH=amd64 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -ldflags $(BUILD_FLAGS) -o ./bin/${BINARY}_${VERSION}_windows_amd64
	cd bin && for f in $$(ls);do \
  	  zip $$f.zip $$f; \
  	  rm $$f; \
  	done

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

