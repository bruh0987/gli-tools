BINARY_DIR := bin
TOOLS := hello

.PHONY: all build test fmt clean

all: test build

build:
	@mkdir -p $(BINARY_DIR)
	@for tool in $(TOOLS); do \
		echo "building $$tool"; \
		go build -trimpath -ldflags="-s -w" -o $(BINARY_DIR)/$$tool ./cmd/$$tool; \
	done

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './bin/*' -not -path './dist/*')

clean:
	rm -rf $(BINARY_DIR) dist
