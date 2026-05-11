HOSTNAME     = registry.terraform.io
NAMESPACE    = terraform-registry
NAME         = registry
BINARY       = terraform-provider-${NAME}
VERSION      = 0.3.2
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

# Regenerate Go types from the backend's OpenAPI 3 spec. Pulls the spec
# from a pinned backend image, patches it for strict OpenAPI 3 validators
# (path-level parameters, deduped enums, declared security schemes — all
# tracked as backend issues #359/#360/#361), and runs oapi-codegen to write
# internal/client/spec/models_gen.go.
#
# CI re-runs this on every PR and fails on `git diff` (see weekly-security.yml).
.PHONY: models-gen
models-gen:
	@echo "==> Pulling openapi3.json from pinned backend image..."
	./internal/client/spec/fetch-spec.sh
	@echo "==> Preprocessing spec for strict validators..."
	python3 internal/client/spec/preprocess.py \
		internal/client/spec/openapi3.json \
		internal/client/spec/openapi3-patched.json
	@echo "==> Generating Go types..."
	cd internal/client/spec && oapi-codegen -config oapi-codegen.yaml openapi3-patched.json
	gofmt -s -w internal/client/spec/models_gen.go
	@echo "==> Done. Generated $$(wc -l < internal/client/spec/models_gen.go) lines."
