package hello

import (
	"flag"
	"fmt"
	"io"
)

func Run(args []string, out io.Writer, errOut io.Writer) int {
	flags := flag.NewFlagSet("hello", flag.ContinueOnError)
	flags.SetOutput(errOut)

	showHelp := flags.Bool("help", false, "show help")
	showHelpShort := flags.Bool("h", false, "show help")
	name := flags.String("name", "world", "name to greet")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *showHelp || *showHelpShort {
		printHelp(out)
		return 0
	}

	fmt.Fprintf(out, "hello, %s\n", *name)
	return 0
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "Print a greeting.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gli hello [--name value]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  -h, --help       show this help")
	fmt.Fprintln(out, "      --name value name to greet")
}
