package reload

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func Run(args []string, out io.Writer, errOut io.Writer) int {
	flags := flag.NewFlagSet("reload", flag.ContinueOnError)
	flags.SetOutput(errOut)

	showHelp := flags.Bool("help", false, "show help")
	showHelpShort := flags.Bool("h", false, "show help")
	check := flags.Bool("check", false, "show PATH diagnostics")
	printOnly := flags.Bool("print", false, "print only the reload command")
	jsonOutput := flags.Bool("json", false, "print diagnostics as JSON")
	spawn := flags.Bool("spawn", false, "open a fresh shell with the latest PATH")
	shell := flags.String("shell", detectShell(), "shell to target: powershell, cmd, bash, zsh, fish")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *showHelp || *showHelpShort {
		printHelp(out)
		return 0
	}
	if flags.NArg() > 0 {
		fmt.Fprintln(errOut, "usage: gli reload [flags]")
		return 2
	}

	info := getInfo(*shell)
	if *jsonOutput {
		return printJSON(out, errOut, info)
	}
	if *check {
		printCheck(out, info)
		return 0
	}
	if *spawn {
		if err := spawnShell(info); err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
		fmt.Fprintf(out, "opened a fresh %s shell\n", info.Shell)
		return 0
	}

	if *printOnly {
		fmt.Fprintln(out, info.Command)
		return 0
	}

	fmt.Fprintln(out, "A child process cannot directly change the PATH of its parent shell.")
	fmt.Fprintln(out, "Run this in your current shell to reload PATH:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, info.Command)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Useful checks:")
	fmt.Fprintln(out, "  gli reload --check")
	fmt.Fprintln(out, "  gli reload --print")
	fmt.Fprintln(out, "  gli reload --spawn")
	return 0
}

type info struct {
	Shell       string   `json:"shell"`
	Command     string   `json:"command"`
	CurrentPath string   `json:"current_path"`
	ReloadPath  string   `json:"reload_path"`
	GliPath     string   `json:"gli_path"`
	PathEntries []string `json:"path_entries"`
	GliOnPath   bool     `json:"gli_on_path"`
}

func getInfo(shell string) info {
	reloadPath := refreshedPath()
	gliPath, _ := exec.LookPath("gli")
	gliDir := ""
	if gliPath != "" {
		gliDir = filepath.Dir(gliPath)
	}
	return info{
		Shell:       normalizeShell(shell),
		Command:     reloadCommand(shell, gliDir),
		CurrentPath: os.Getenv("PATH"),
		ReloadPath:  reloadPath,
		GliPath:     gliPath,
		PathEntries: splitPath(reloadPath),
		GliOnPath:   gliPath != "",
	}
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "Refresh shell PATH helpers.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gli reload [flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  gli reload")
	fmt.Fprintln(out, "  gli reload --print")
	fmt.Fprintln(out, "  Invoke-Expression (gli reload --print)")
	fmt.Fprintln(out, "  gli reload --shell bash --print")
	fmt.Fprintln(out, "  gli reload --check")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  -h, --help        show this help")
	fmt.Fprintln(out, "      --check       show PATH diagnostics")
	fmt.Fprintln(out, "      --json        print diagnostics as JSON")
	fmt.Fprintln(out, "      --print       print only the reload command")
	fmt.Fprintln(out, "      --shell name  target shell: powershell, cmd, bash, zsh, fish")
	fmt.Fprintln(out, "      --spawn       open a fresh shell with the latest PATH")
}

func printCheck(out io.Writer, info info) {
	fmt.Fprintf(out, "shell: %s\n", info.Shell)
	fmt.Fprintf(out, "gli: %s\n", valueOrNone(info.GliPath))
	fmt.Fprintf(out, "gli on PATH: %t\n", info.GliOnPath)
	fmt.Fprintf(out, "PATH entries after reload: %d\n", len(info.PathEntries))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Reload command:")
	fmt.Fprintln(out, info.Command)
}

func printJSON(out io.Writer, errOut io.Writer, info info) int {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(info); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

func detectShell() string {
	if runtime.GOOS == "windows" {
		if strings.Contains(strings.ToLower(os.Getenv("PSModulePath")), "powershell") {
			return "powershell"
		}
		return "cmd"
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}
	parts := strings.FieldsFunc(shell, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) == 0 {
		return "bash"
	}
	return parts[len(parts)-1]
}

func normalizeShell(shell string) string {
	shell = strings.ToLower(strings.TrimSpace(shell))
	switch shell {
	case "pwsh", "powershell.exe", "pwsh.exe":
		return "powershell"
	case "cmd.exe":
		return "cmd"
	default:
		return shell
	}
}

func reloadCommand(shell string, gliDir string) string {
	switch normalizeShell(shell) {
	case "powershell":
		command := `$env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")`
		if gliDir != "" {
			command += `; if (($env:Path -split ';') -notcontains '` + escapePowerShellSingleQuoted(gliDir) + `') { $env:Path += ';` + escapePowerShellSingleQuoted(gliDir) + `' }`
		}
		return command
	case "cmd":
		return `for /f "tokens=2,*" %A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path') do set "MACHINE_PATH=%B" && for /f "tokens=2,*" %A in ('reg query "HKCU\Environment" /v Path') do set "USER_PATH=%B" && set "Path=%MACHINE_PATH%;%USER_PATH%"`
	case "fish":
		return "exec fish -l"
	case "zsh":
		return "exec zsh -l"
	default:
		return "exec ${SHELL:-bash} -l"
	}
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}

func refreshedPath() string {
	path := baseRefreshedPath()
	gliPath, err := exec.LookPath("gli")
	if err == nil && gliPath != "" {
		path = ensurePathEntry(path, filepath.Dir(gliPath))
	}
	return path
}

func baseRefreshedPath() string {
	if runtime.GOOS != "windows" {
		return os.Getenv("PATH")
	}
	if path := windowsDotNetPath(); path != "" {
		return path
	}
	machine := registryValue(`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "Path")
	if machine == "" {
		machine = os.Getenv("PATH")
	}
	user := registryValue(`HKCU\Environment`, "Path")
	if user == "" {
		return machine
	}
	if machine == "" {
		return user
	}
	return machine + ";" + user
}

func ensurePathEntry(path string, entry string) string {
	if entry == "" {
		return path
	}
	for _, existing := range splitPath(path) {
		if strings.EqualFold(existing, entry) {
			return path
		}
	}
	if path == "" {
		return entry
	}
	return path + string(os.PathListSeparator) + entry
}

func windowsDotNetPath() string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `[Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")`)
	data, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func registryValue(key string, value string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	cmd := exec.Command("reg", "query", key, "/v", value)
	data, err := cmd.Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		for i, field := range fields {
			if strings.HasPrefix(field, "REG_") && i+1 < len(fields) {
				return expandWindowsEnv(strings.Join(fields[i+1:], " "))
			}
		}
	}
	return ""
}

func expandWindowsEnv(value string) string {
	replacements := map[string]string{
		"%USERPROFILE%":       os.Getenv("USERPROFILE"),
		"%LOCALAPPDATA%":      os.Getenv("LOCALAPPDATA"),
		"%APPDATA%":           os.Getenv("APPDATA"),
		"%ProgramFiles%":      os.Getenv("ProgramFiles"),
		"%ProgramFiles(x86)%": os.Getenv("ProgramFiles(x86)"),
		"%SystemRoot%":        os.Getenv("SystemRoot"),
		"%SYSTEMROOT%":        os.Getenv("SystemRoot"),
		"%windir%":            os.Getenv("windir"),
		"%WINDIR%":            os.Getenv("windir"),
	}
	for name, replacement := range replacements {
		if replacement != "" {
			value = strings.ReplaceAll(value, name, replacement)
		}
	}
	return value
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, string(os.PathListSeparator))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func spawnShell(info info) error {
	env := append(os.Environ(), "PATH="+info.ReloadPath, "Path="+info.ReloadPath)
	switch info.Shell {
	case "powershell":
		cmd := exec.Command("powershell", "-NoExit")
		cmd.Env = env
		return cmd.Start()
	case "cmd":
		cmd := exec.Command("cmd")
		cmd.Env = env
		return cmd.Start()
	case "fish", "zsh", "bash":
		cmd := exec.Command(info.Shell, "-l")
		cmd.Env = env
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported shell for --spawn: %s", info.Shell)
	}
}

func valueOrNone(value string) string {
	if value == "" {
		return "[none]"
	}
	return value
}
