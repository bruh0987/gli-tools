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

	dryRun := flags.Bool("dry-run", false, "print what would be updated without replacing the binary")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() > 1 {
		fmt.Fprintln(errOut, "usage: gli update [git-ref] [--dry-run]")
		return 2
	}

	ref := "main"
	if flags.NArg() == 1 {
		ref = flags.Arg(0)
	}

	if err := update(ref, *dryRun, out, errOut); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

func update(ref string, dryRun bool, out io.Writer, errOut io.Writer) error {
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
	if dryRun {
		fmt.Fprintln(out, "dry run: no files changed")
		return nil
	}

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
	if runtime.GOOS != "windows" {
		defer os.RemoveAll(workDir)
	}

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

	replacement := filepath.Join(workDir, executableName("gli-new"))
	ldflags := fmt.Sprintf("-s -w -X main.commit=%s", strings.TrimSpace(commit))
	if err := runIn(repoDir, errOut, "go", "build", "-trimpath", "-ldflags", ldflags, "-o", replacement, "./cmd/gli"); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
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
