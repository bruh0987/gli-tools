package text

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
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
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return interactiveRaw(out, errOut, width)
	}
	return interactiveLine(out, errOut, width)
}

func interactiveLine(out io.Writer, errOut io.Writer, width int) int {
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

func interactiveRaw(out io.Writer, errOut io.Writer, width int) int {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	styles := StyleNames()
	styleIndex := 0
	input := editPromptRaw(out, "Text", "")
	if strings.TrimSpace(input) == "" {
		fmt.Fprintln(errOut, "\r\ntext cannot be empty")
		return 2
	}

	for {
		renderInteractive(out, input, styles[styleIndex], styleIndex, len(styles), width, "", nil)
		key := readKey()
		switch key {
		case "n", "right", "down", " ":
			styleIndex = (styleIndex + 1) % len(styles)
		case "p", "left", "up":
			styleIndex--
			if styleIndex < 0 {
				styleIndex = len(styles) - 1
			}
		case "/":
			styleIndex = searchRaw(out, input, styles, styleIndex, width)
		case "t":
			next := editPromptRaw(out, "Text", input)
			if strings.TrimSpace(next) != "" {
				input = next
			}
		case "c":
			rendered, _ := Render(input, styles[styleIndex], width)
			if err := Copy(rendered); err != nil {
				showMessageRaw(out, "copy failed: "+err.Error())
			} else {
				showMessageRaw(out, "copied as text")
			}
		case "e":
			format := exportPickerRaw(out, input, styles[styleIndex], width)
			if format != "" {
				rendered, _ := Render(input, styles[styleIndex], width)
				exported, err := Export(rendered, format)
				if err != nil {
					showMessageRaw(out, err.Error())
				} else if err := Copy(exported); err != nil {
					showMessageRaw(out, "copy failed: "+err.Error())
				} else {
					showMessageRaw(out, "copied as "+format)
				}
			}
		case "q", "esc", "ctrl-c":
			fmt.Fprint(out, "\r\n")
			return 0
		}
	}
}

func renderInteractive(out io.Writer, input string, styleName string, index int, total int, width int, mode string, options []string) {
	clearScreen(out)
	rendered, err := Render(input, styleName, width)
	if err != nil {
		rendered = err.Error() + "\n"
	}
	fmt.Fprintf(out, "text: %s\r\nstyle: %s (%d/%d)\r\n\r\n%s\r\n", input, styleName, index+1, total, crlf(rendered))
	fmt.Fprintln(out, "[n/p] style  [/] search  [t] text  [e] export+copy  [c] copy  [q] quit")
	if mode != "" {
		fmt.Fprintf(out, "\r\n%s\r\n", mode)
	}
	for _, option := range options {
		fmt.Fprintf(out, "  %s\r\n", option)
	}
}

func searchRaw(out io.Writer, input string, styles []string, current int, width int) int {
	query := ""
	selected := current
	for {
		matches := matchingStyles(styles, query)
		if len(matches) > 0 {
			selected = matches[0]
		}
		options := make([]string, 0, min(8, len(matches)))
		for _, index := range matches[:min(8, len(matches))] {
			prefix := " "
			if index == selected {
				prefix = ">"
			}
			options = append(options, prefix+" "+styles[index])
		}
		renderInteractive(out, input, styles[selected], selected, len(styles), width, "search: "+query, options)
		key := readKey()
		switch key {
		case "enter":
			return selected
		case "esc":
			return current
		case "backspace":
			if len(query) > 0 {
				query = query[:len(query)-1]
			}
		default:
			if len(key) == 1 {
				query += key
			}
		}
	}
}

func exportPickerRaw(out io.Writer, input string, styleName string, width int) string {
	formats := []string{"text", "go", "js", "ts", "py", "json", "html", "rust", "csharp", "sh"}
	selected := 0
	for {
		options := make([]string, 0, len(formats))
		for i, format := range formats {
			prefix := " "
			if i == selected {
				prefix = ">"
			}
			options = append(options, fmt.Sprintf("%s %d %s", prefix, i+1, format))
		}
		renderInteractive(out, input, styleName, 0, 1, width, "export format: arrows/n/p, number, enter, esc", options)
		key := readKey()
		switch key {
		case "enter":
			return formats[selected]
		case "esc":
			return ""
		case "n", "right", "down":
			selected = (selected + 1) % len(formats)
		case "p", "left", "up":
			selected--
			if selected < 0 {
				selected = len(formats) - 1
			}
		default:
			if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
				index := int(key[0] - '1')
				if index < len(formats) {
					return formats[index]
				}
			}
			if key == "0" {
				return formats[9]
			}
		}
	}
}

func editPromptRaw(out io.Writer, label string, initial string) string {
	value := initial
	for {
		clearScreen(out)
		fmt.Fprintf(out, "%s: %s\r\n\r\nEnter to accept, Esc to cancel\r\n", label, value)
		key := readKey()
		switch key {
		case "enter":
			return value
		case "esc":
			return initial
		case "backspace":
			if len(value) > 0 {
				value = value[:len(value)-1]
			}
		default:
			if len(key) == 1 {
				value += key
			}
		}
	}
}

func showMessageRaw(out io.Writer, message string) {
	fmt.Fprintf(out, "\r\n%s\r\npress any key...", message)
	_ = readKey()
}

func readKey() string {
	var b [3]byte
	n, err := os.Stdin.Read(b[:1])
	if err != nil || n == 0 {
		return ""
	}
	switch b[0] {
	case 3:
		return "ctrl-c"
	case 13, 10:
		return "enter"
	case 27:
		os.Stdin.Read(b[1:2])
		if b[1] == '[' {
			os.Stdin.Read(b[2:3])
			switch b[2] {
			case 'A':
				return "up"
			case 'B':
				return "down"
			case 'C':
				return "right"
			case 'D':
				return "left"
			}
		}
		return "esc"
	case 8, 127:
		return "backspace"
	default:
		return string(b[0])
	}
}

func matchingStyles(styles []string, query string) []int {
	var matches []int
	query = strings.ToLower(query)
	for i, style := range styles {
		if query == "" || strings.Contains(style, query) {
			matches = append(matches, i)
		}
	}
	return matches
}

func crlf(value string) string {
	return strings.ReplaceAll(value, "\n", "\r\n")
}

func wait(reader *bufio.Reader, out io.Writer) {
	fmt.Fprint(out, "Press Enter...")
	_, _ = reader.ReadString('\n')
}

func clearScreen(out io.Writer) {
	fmt.Fprint(out, "\033[2J\033[H")
}
