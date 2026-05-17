package main

import (
	"os"

	"github.com/bruh0987/gli-tools/internal/build"
	"github.com/bruh0987/gli-tools/internal/gli"
)

var version = "dev"
var commit = "unknown"

func main() {
	build.Version = version
	build.Commit = commit
	os.Exit(gli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
