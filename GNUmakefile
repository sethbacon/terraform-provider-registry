HOSTNAME     = registry.terraform.io
NAMESPACE    = terraform-registry
NAME         = registry
BINARY       = terraform-provider-${NAME}
VERSION      = 0.1.0
OS_ARCH      = $(shell go env GOOS)_$(shell go env GOARCH)

default: build

.PHONY: build
build:
	go build -o ${BINARY} .

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

.PHONY: test
test:
	go test ./... -v -count=1 -timeout 10m

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./internal/provider/... -v -count=1 -timeout 120m

.PHONY: fmt
fmt:
	gofmt -s -w .
	@which goimports > /dev/null 2>&1 && goimports -w . || true

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: docs
docs:
	go generate ./...

.PHONY: clean
clean:
	rm -f ${BINARY}

.PHONY: tidy
tidy:
	go mod tidy
