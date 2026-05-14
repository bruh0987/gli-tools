package main

import (
	"os"

	"github.com/bruh0987/gli-tools/internal/gli"
)

var version = "dev"

func main() {
	os.Exit(gli.Run(os.Args[1:], os.Stdout, os.Stderr, version))
}
