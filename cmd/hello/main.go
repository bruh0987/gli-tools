package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bruh0987/gli-tools/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags := flag.NewFlagSet("hello", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	showVersion := flags.Bool("version", false, "print version and exit")
	name := flags.String("name", "world", "name to greet")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	app := cli.App{
		Name:    "hello",
		Version: version,
		Out:     os.Stdout,
		Err:     os.Stderr,
	}

	if *showVersion {
		app.PrintVersion()
		return 0
	}

	fmt.Fprintf(os.Stdout, "hello, %s\n", *name)
	return 0
}
