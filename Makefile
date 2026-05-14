BINARY_DIR := bin
TOOL := gli

.PHONY: all build test fmt clean

all: test build

build:
	@mkdir -p $(BINARY_DIR)
	go build -trimpath -ldflags="-s -w" -o $(BINARY_DIR)/$(TOOL) ./cmd/$(TOOL)

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './bin/*' -not -path './dist/*')

clean:
	rm -rf $(BINARY_DIR) dist
