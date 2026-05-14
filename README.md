# gli-tools

A collection of small, fast, cross-platform CLI tools written in Go.

The project is intentionally simple:

- each command lives in `cmd/<tool-name>`
- shared code lives in `internal/`
- tools are built as single native binaries
- the standard library is preferred unless a dependency clearly earns its keep

## Requirements

- Go 1.25 or newer
- `make` for the convenience commands, or plain `go` commands on any platform

## Tools

| Tool | Description |
| --- | --- |
| `hello` | Minimal example command used as the project skeleton. |

## Build

Build every tool:

```sh
make build
```

Without `make`:

```sh
go build -trimpath -ldflags="-s -w" -o bin/hello ./cmd/hello
```

On Windows PowerShell, build the sample tool with:

```powershell
go build -trimpath -ldflags="-s -w" -o bin/hello.exe ./cmd/hello
```

## Test

```sh
make test
```

or:

```sh
go test ./...
```

## Add A Tool

Create a new folder under `cmd/`:

```text
cmd/my-tool/main.go
```

Keep command startup cheap:

- avoid global initialization that does I/O
- parse only the flags the command needs
- stream input and output where possible
- keep dependencies small and intentional

Then add the tool name to `TOOLS` in the `Makefile`.

## Cross-Compile

Go can build native binaries for other platforms with `GOOS` and `GOARCH`:

```sh
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/hello-linux-amd64 ./cmd/hello
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/hello-darwin-arm64 ./cmd/hello
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/hello-windows-amd64.exe ./cmd/hello
```

## License

MIT
