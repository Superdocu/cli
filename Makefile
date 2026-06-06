SPEC_URL ?= https://developers.superdocu.com/api-docs/v2/api.yaml
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -X main.version=$(VERSION)

.PHONY: build install gen test vet fmt clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/superdocu ./cmd/superdocu

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/superdocu

# Refresh the embedded OpenAPI spec from the published source, keeping the CLI
# in sync with the API. Rebuild afterwards.
gen:
	curl -fsSL $(SPEC_URL) -o openapi/api.yaml
	@echo "Updated openapi/api.yaml — rebuild with: make build"

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin dist
