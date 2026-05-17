BINARY_DIR := bin
TOOL := gli
VERSION ?= 0.1.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: all build test fmt clean

all: test build

build:
	@mkdir -p $(BINARY_DIR)
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/$(TOOL) ./cmd/$(TOOL)

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './bin/*' -not -path './dist/*')

clean:
	rm -rf $(BINARY_DIR) dist
