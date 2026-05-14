package hello

import (
	"flag"
	"fmt"
	"io"
)

func Run(args []string, out io.Writer, errOut io.Writer) int {
	flags := flag.NewFlagSet("hello", flag.ContinueOnError)
	flags.SetOutput(errOut)

	name := flags.String("name", "world", "name to greet")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	fmt.Fprintf(out, "hello, %s\n", *name)
	return 0
}
