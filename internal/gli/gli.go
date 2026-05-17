package gli

import (
	"flag"
	"fmt"
	"io"
	"sort"

	"github.com/bruh0987/gli-tools/internal/build"
	"github.com/bruh0987/gli-tools/internal/commands/hello"
	"github.com/bruh0987/gli-tools/internal/commands/lines"
	"github.com/bruh0987/gli-tools/internal/commands/reload"
	"github.com/bruh0987/gli-tools/internal/commands/update"
)

type Command struct {
	Name        string
	Description string
	Run         func(args []string, out io.Writer, errOut io.Writer) int
}

func Run(args []string, out io.Writer, errOut io.Writer) int {
	commands := map[string]Command{
		"hello": {
			Name:        "hello",
			Description: "Print a greeting.",
			Run:         hello.Run,
		},
		"lines": {
			Name:        "lines",
			Description: "Count lines by file extension.",
			Run:         lines.Run,
		},
		"reload": {
			Name:        "reload",
			Description: "Refresh shell PATH helpers.",
			Run:         reload.Run,
		},
		"update": {
			Name:        "update",
			Description: "Update gli from GitHub.",
			Run:         update.Run,
		},
	}

	flags := flag.NewFlagSet("gli", flag.ContinueOnError)
	flags.SetOutput(errOut)

	showVersion := flags.Bool("version", false, "print version and exit")
	showVersionShort := flags.Bool("v", false, "print version and exit")
	showHelp := flags.Bool("help", false, "print help and exit")
	showHelpShort := flags.Bool("h", false, "print help and exit")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *showVersion || *showVersionShort {
		fmt.Fprintf(out, "gli %s\n", build.DisplayVersion())
		return 0
	}

	rest := flags.Args()
	if *showHelp || *showHelpShort {
		printHelp(out, commands)
		return 0
	}
	if len(rest) == 0 {
		printSplash(out, commands)
		return 0
	}

	name := rest[0]
	command, ok := commands[name]
	if !ok {
		fmt.Fprintf(errOut, "unknown command %q\n\n", name)
		printHelp(errOut, commands)
		return 2
	}

	return command.Run(rest[1:], out, errOut)
}

func printSplash(out io.Writer, commands map[string]Command) {
	fmt.Fprintf(out, "gli tools %s\n", build.DisplayVersion())
	fmt.Fprintf(out, "%s\n", build.RepoURL)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gli <command> <flags>")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Run `gli -h` to see all %d commands.\n", len(commands))
	fmt.Fprintln(out, "Run `gli update` to update from GitHub main.")
}

func printHelp(out io.Writer, commands map[string]Command) {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintln(out, "gli is a collection of small, fast CLI tools.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gli <command> <flags>")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Commands:")
	for _, name := range names {
		command := commands[name]
		fmt.Fprintf(out, "  %-12s %s\n", command.Name, command.Description)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Global flags:")
	fmt.Fprintln(out, "  -h, --help      print help and exit")
	fmt.Fprintln(out, "  -v, --version   print version and exit")
}
