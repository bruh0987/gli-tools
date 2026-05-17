package update

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bruh0987/gli-tools/internal/build"
)

func Run(args []string, out io.Writer, errOut io.Writer) int {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(errOut)

	showHelp := flags.Bool("help", false, "show help")
	showHelpShort := flags.Bool("h", false, "show help")
	dryRun := flags.Bool("dry-run", false, "print what would be updated without replacing the binary")
	force := flags.Bool("force", false, "replace the binary even when it already matches the target")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *showHelp || *showHelpShort {
		printHelp(out)
		return 0
	}
	if flags.NArg() > 1 {
		fmt.Fprintln(errOut, "usage: gli update [git-ref] [--dry-run] [--force]")
		return 2
	}

	ref := "main"
	if flags.NArg() == 1 {
		ref = flags.Arg(0)
	}

	if err := update(ref, *dryRun, *force, out, errOut); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "Update gli from GitHub.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gli update [git-ref] [flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  gli update")
	fmt.Fprintln(out, "  gli update 2125448")
	fmt.Fprintln(out, "  gli update v0.1.0")
	fmt.Fprintln(out, "  gli update --force main")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "The git-ref can be a branch, tag, or commit hash. Without a ref, main is used.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  -h, --help     show this help")
	fmt.Fprintln(out, "      --dry-run  show what would be updated without replacing the binary")
	fmt.Fprintln(out, "      --force    replace the binary even when already on the target commit")
}

func update(ref string, dryRun bool, force bool, out io.Writer, errOut io.Writer) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}

	fmt.Fprintf(out, "repo: %s\n", build.RepoURL)
	fmt.Fprintf(out, "target: %s\n", ref)
	fmt.Fprintf(out, "binary: %s\n", self)

	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git is required for gli update")
	}
	if _, err := exec.LookPath("go"); err != nil {
		return errors.New("go is required for gli update")
	}

	workDir, err := os.MkdirTemp("", "gli-update-*")
	if err != nil {
		return err
	}
	keepWorkDir := false
	defer func() {
		if !keepWorkDir {
			_ = os.RemoveAll(workDir)
		}
	}()

	repoDir := filepath.Join(workDir, "repo")
	if err := run(errOut, "git", "clone", "--filter=blob:none", build.RepoGit, repoDir); err != nil {
		return err
	}
	if err := runIn(repoDir, errOut, "git", "checkout", ref); err != nil {
		return err
	}
	commit, err := outputIn(repoDir, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return err
	}
	targetCommit := strings.TrimSpace(commit)
	fmt.Fprintf(out, "target commit: %s\n", targetCommit)

	if !force && sameCommit(build.Commit, targetCommit) {
		fmt.Fprintln(out, "gli is already on the newest version for this target.")
		fmt.Fprintln(out, "Run `gli update --force` to rebuild and replace it anyway.")
		return nil
	}
	if dryRun {
		fmt.Fprintln(out, "dry run: no files changed")
		return nil
	}

	replacement := filepath.Join(workDir, executableName("gli-new"))
	ldflags := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s", build.Version, targetCommit)
	if err := runIn(repoDir, errOut, "go", "build", "-trimpath", "-ldflags", ldflags, "-o", replacement, "./cmd/gli"); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		keepWorkDir = true
		return replaceOnWindows(self, replacement, workDir, out)
	}

	if err := os.Chmod(replacement, 0o755); err != nil {
		return err
	}
	if err := os.Rename(replacement, self); err != nil {
		return err
	}

	fmt.Fprintln(out, "updated gli")
	return nil
}

func sameCommit(current string, target string) bool {
	current = strings.TrimSpace(current)
	target = strings.TrimSpace(target)
	if current == "" || current == "unknown" || target == "" {
		return false
	}
	return current == target || strings.HasPrefix(current, target) || strings.HasPrefix(target, current)
}

func run(out io.Writer, name string, args ...string) error {
	return runIn("", out, name, args...)
}

func runIn(dir string, out io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func outputIn(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	data, err := cmd.Output()
	return string(data), err
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func replaceOnWindows(self string, replacement string, workDir string, out io.Writer) error {
	script, err := writeWindowsReplaceScript(self, replacement, workDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/C", "start", "", "/B", script, strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Fprintln(out, "downloaded update; replacement will finish after this process exits")
	return nil
}

func writeWindowsReplaceScript(self string, replacement string, workDir string) (string, error) {
	script := filepath.Join(filepath.Dir(replacement), "replace-gli.cmd")
	content := strings.Join([]string{
		"@echo off",
		"cd /D %TEMP%",
		"set pid=%1",
		":wait",
		"tasklist /FI \"PID eq %pid%\" 2>NUL | find \"%pid%\" >NUL",
		"if not errorlevel 1 (",
		"  timeout /T 1 /NOBREAK >NUL",
		"  goto wait",
		")",
		"move /Y " + quoteCmd(replacement) + " " + quoteCmd(self) + " >NUL",
		"del \"%~f0\"",
		"rmdir /S /Q " + quoteCmd(workDir) + " >NUL 2>NUL",
		"",
	}, "\r\n")

	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		return "", err
	}
	_ = time.Now()
	return script, nil
}

func quoteCmd(path string) string {
	return `"` + strings.ReplaceAll(path, `"`, `\"`) + `"`
}
