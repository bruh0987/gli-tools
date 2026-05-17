# gli-tools

A collection of small, fast, cross-platform CLI tools written in Go.

After installation, everything runs through one command:

```sh
gli <command> <flags>
```

The project is intentionally simple:

- `cmd/gli` builds the single executable
- each subcommand lives in `internal/commands/<command-name>`
- shared code lives in `internal/`
- tools are built as single native binaries
- the standard library is preferred unless a dependency clearly earns its keep

## Requirements

- Go 1.25 or newer
- `make` for the convenience commands, or plain `go` commands on any platform

## Tools

| Tool | Description |
| --- | --- |
| `gli hello` | Minimal example command used as the project skeleton. |
| `gli lines` | Count lines recursively and group them by file extension. |
| `gli reload` | Print or run helpers for refreshing shell `PATH`. |
| `gli update` | Replace the current `gli` binary from GitHub. |

## Usage

```sh
gli
gli -h
gli -v
gli --help
gli --version
gli <command> <flags>
```

`gli update` pulls and builds the latest `main` branch from GitHub. If the installed binary already matches the target commit, it exits without replacing itself. Use `--force` to rebuild and replace anyway.

To pin or downgrade, pass any Git ref:

```sh
gli update
gli update 3f96d40
gli update v0.1.0
gli update --dry-run main
gli update --force main
```

The update target can be a branch, tag, or commit hash.

### `gli reload`

Refresh helpers for when `PATH` changed after installing tools:

```sh
gli reload
gli reload --print
gli reload --check
gli reload --shell powershell --print
gli reload --spawn
```

A process cannot directly mutate the already-running parent shell, so the default output gives you the command to run in your current shell. On PowerShell, this works:

```powershell
Invoke-Expression (gli reload --print)
```

### `gli lines`

Count all lines under the current directory:

```sh
gli lines
```

Count another directory:

```sh
gli lines ./path/to/project
```

Show the largest files by line count:

```sh
gli lines --top 10 ./path/to/project
```

Honor the root `.gitignore` file:

```sh
gli lines --gitignore ./path/to/project
```

Exclude extensions:

```sh
gli lines --exclude md,json,txt ./path/to/project
```

## Install

### Windows PowerShell

Run:

```powershell
.\scripts\install-path.ps1
```

This builds `gli.exe` into `%USERPROFILE%\.gli\bin` and adds that folder to your user `PATH` if needed. Open a new terminal after running it.

### macOS And Linux

Run:

```sh
sh ./scripts/install-path.sh
```

This builds `gli` into `$HOME/.local/bin`. If that directory is not already on `PATH`, the script prints the line to add to your shell profile.

## Build

Build every tool:

```sh
make build
```

Without `make`:

```sh
go build -trimpath -ldflags="-s -w -X main.version=dev -X main.commit=$(git rev-parse --short HEAD)" -o bin/gli ./cmd/gli
```

On Windows PowerShell, build the sample tool with:

```powershell
go build -trimpath -ldflags="-s -w -X main.version=dev -X main.commit=$(git rev-parse --short HEAD)" -o bin/gli.exe ./cmd/gli
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

Create a new folder under `internal/commands/`:

```text
internal/commands/mytool/mytool.go
```

Keep command startup cheap:

- avoid global initialization that does I/O
- parse only the flags the command needs
- stream input and output where possible
- keep dependencies small and intentional

Then register the command in `internal/gli/gli.go`.

## Cross-Compile

Go can build native binaries for other platforms with `GOOS` and `GOARCH`:

```sh
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/gli-linux-amd64 ./cmd/gli
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/gli-darwin-arm64 ./cmd/gli
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/gli-windows-amd64.exe ./cmd/gli
```

## License

MIT
