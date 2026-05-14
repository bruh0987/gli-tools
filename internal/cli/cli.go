package cli

import (
	"fmt"
	"io"
)

type App struct {
	Name        string
	Version     string
	Description string
	Out         io.Writer
	Err         io.Writer
}

func (a App) PrintVersion() {
	out := a.Out
	if out == nil {
		out = io.Discard
	}

	version := a.Version
	if version == "" {
		version = "dev"
	}

	fmt.Fprintf(out, "%s %s\n", a.Name, version)
}
