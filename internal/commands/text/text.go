package text

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func Run(args []string, out io.Writer, errOut io.Writer) int {
	args = normalizeArgs(args)
	flags := flag.NewFlagSet("text", flag.ContinueOnError)
	flags.SetOutput(errOut)

	showHelp := flags.Bool("help", false, "show help")
	showHelpShort := flags.Bool("h", false, "show help")
	style := flags.String("style", "standard", "style name")
	list := flags.Bool("list", false, "list available styles")
	preview := flags.String("preview", "", "render every style for this text")
	exportFormat := flags.String("export", "text", "export format: text, go, js, ts, py, json, html, rust, csharp, sh")
	copyOutput := flags.Bool("copy", false, "copy generated output to clipboard")
	outFile := flags.String("out", "", "write generated output to file")
	width := flags.Int("width", 80, "reserved output width")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *showHelp || *showHelpShort {
		printHelp(out)
		return 0
	}
	if *list {
		printStyles(out)
		return 0
	}
	if *preview != "" {
		if err := printPreview(out, *preview, *width); err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
		return 0
	}

	parts := flags.Args()
	if len(parts) == 0 {
		return interactive(out, errOut, *width)
	}

	input := strings.Join(parts, " ")
	if strings.TrimSpace(input) == "" {
		fmt.Fprintln(errOut, "text cannot be empty")
		return 2
	}

	rendered, err := Render(input, *style, *width)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}

	result, err := Export(rendered, *exportFormat)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}

	if *outFile != "" {
		if err := os.WriteFile(*outFile, []byte(result), 0o644); err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
	}
	if *copyOutput {
		if err := Copy(result); err != nil {
			fmt.Fprintf(errOut, "copy failed: %v\n", err)
		} else {
			fmt.Fprintf(errOut, "copied as %s\n", *exportFormat)
		}
	}

	fmt.Fprint(out, result)
	if !strings.HasSuffix(result, "\n") {
		fmt.Fprintln(out)
	}
	return 0
}

func normalizeArgs(args []string) []string {
	valueFlags := map[string]bool{
		"--style": true, "--preview": true, "--export": true, "--out": true, "--width": true,
	}
	boolFlags := map[string]bool{
		"-h": true, "--help": true, "--list": true, "--copy": true,
	}

	var flagArgs []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			name := arg
			if before, _, ok := strings.Cut(arg, "="); ok {
				name = before
			}
			if valueFlags[name] {
				flagArgs = append(flagArgs, arg)
				if !strings.Contains(arg, "=") && i+1 < len(args) {
					i++
					flagArgs = append(flagArgs, args[i])
				}
				continue
			}
			if boolFlags[name] {
				flagArgs = append(flagArgs, arg)
				continue
			}
		}
		if boolFlags[arg] {
			flagArgs = append(flagArgs, arg)
			continue
		}
		positional = append(positional, arg)
	}
	return append(flagArgs, positional...)
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "Generate ASCII text art.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gli text [text] [flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  gli text Hello")
	fmt.Fprintln(out, "  gli text \"Hello world\" --style slant")
	fmt.Fprintln(out, "  gli text --preview Hello")
	fmt.Fprintln(out, "  gli text \"Hello\" --export go --copy")
	fmt.Fprintln(out, "  gli text \"Hello\" --style block --out banner.txt")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  -h, --help          show this help")
	fmt.Fprintln(out, "      --style name    style name, default standard")
	fmt.Fprintln(out, "      --list          list available styles")
	fmt.Fprintln(out, "      --preview text  render all styles for quick comparison")
	fmt.Fprintln(out, "      --export fmt    text, go, js, ts, py, json, html, rust, csharp, sh")
	fmt.Fprintln(out, "      --copy          copy generated output to clipboard")
	fmt.Fprintln(out, "      --out file      write generated output to file")
	fmt.Fprintln(out, "      --width number  reserved output width, default 80")
}

func printStyles(out io.Writer) {
	names := StyleNames()
	for _, name := range names {
		fmt.Fprintln(out, name)
	}
}

func printPreview(out io.Writer, input string, width int) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("preview text cannot be empty")
	}
	for _, name := range StyleNames() {
		rendered, err := Render(input, name, width)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "== %s ==\n%s\n", name, rendered)
	}
	return nil
}

func interactive(out io.Writer, errOut io.Writer, width int) int {
	reader := bufio.NewReader(os.Stdin)
	styles := StyleNames()
	styleIndex := 0

	fmt.Fprint(out, "Text: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Fprintln(errOut, "text cannot be empty")
		return 2
	}

	for {
		clearScreen(out)
		rendered, err := Render(input, styles[styleIndex], width)
		if err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
		fmt.Fprintf(out, "style: %s (%d/%d)\n\n%s\n", styles[styleIndex], styleIndex+1, len(styles), rendered)
		fmt.Fprintln(out, "[n] next  [p] previous  [/] search  [t] text  [e] export+copy  [c] copy text  [q] quit")
		fmt.Fprint(out, "> ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		switch command {
		case "n", "":
			styleIndex = (styleIndex + 1) % len(styles)
		case "p":
			styleIndex--
			if styleIndex < 0 {
				styleIndex = len(styles) - 1
			}
		case "/":
			fmt.Fprint(out, "Search style: ")
			query, _ := reader.ReadString('\n')
			query = strings.ToLower(strings.TrimSpace(query))
			for i, name := range styles {
				if strings.Contains(name, query) {
					styleIndex = i
					break
				}
			}
		case "t":
			fmt.Fprint(out, "Text: ")
			next, _ := reader.ReadString('\n')
			next = strings.TrimSpace(next)
			if next != "" {
				input = next
			}
		case "c":
			if err := Copy(rendered); err != nil {
				fmt.Fprintf(errOut, "copy failed: %v\n", err)
			} else {
				fmt.Fprintln(out, "copied as text")
			}
			wait(reader, out)
		case "e":
			fmt.Fprint(out, "Export format: ")
			format, _ := reader.ReadString('\n')
			format = strings.TrimSpace(format)
			exported, err := Export(rendered, format)
			if err != nil {
				fmt.Fprintln(errOut, err)
			} else if err := Copy(exported); err != nil {
				fmt.Fprintf(errOut, "copy failed: %v\n", err)
			} else {
				fmt.Fprintf(out, "copied as %s\n", format)
			}
			wait(reader, out)
		case "q":
			return 0
		}
	}
}

func wait(reader *bufio.Reader, out io.Writer) {
	fmt.Fprint(out, "Press Enter...")
	_, _ = reader.ReadString('\n')
}

func clearScreen(out io.Writer) {
	fmt.Fprint(out, "\033[2J\033[H")
}

func StyleNames() []string {
	names := make([]string, 0, len(styles))
	for name := range styles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
