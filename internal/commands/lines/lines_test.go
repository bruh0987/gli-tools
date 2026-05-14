package lines

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCountsLinesByExtension(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "README.md", "# Title\ntext")
	writeFile(t, dir, "nested/data.json", "{\n  \"ok\": true\n}\n")

	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{dir}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0: %s", code, errOut.String())
	}

	output := out.String()
	for _, want := range []string{".go", ".md", ".json"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %s: %s", want, output)
		}
	}
}

func TestRunExcludesExtensions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	writeFile(t, dir, "README.md", "# Title\n")

	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"--exclude", "md", dir}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0: %s", code, errOut.String())
	}

	output := out.String()
	if strings.Contains(output, ".md") {
		t.Fatalf("output should not include .md: %s", output)
	}
	if !strings.Contains(output, ".go") {
		t.Fatalf("output missing .go: %s", output)
	}
}

func TestRunTopFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "small.go", "one\n")
	writeFile(t, dir, "large.go", "one\ntwo\nthree\n")

	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"--top", "1", dir}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0: %s", code, errOut.String())
	}

	output := out.String()
	if !strings.Contains(output, "large.go") {
		t.Fatalf("output missing large.go: %s", output)
	}
	if strings.Contains(output, "small.go") {
		t.Fatalf("output should not include small.go in top 1: %s", output)
	}
}

func TestRunHonorsGitignore(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".gitignore", "ignored/\n*.log\n")
	writeFile(t, dir, "keep.go", "package keep\n")
	writeFile(t, dir, "ignored/nope.go", "package nope\n")
	writeFile(t, dir, "debug.log", "noise\n")

	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"--gitignore", dir}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0: %s", code, errOut.String())
	}

	output := out.String()
	if strings.Contains(output, ".log") {
		t.Fatalf("output should not include ignored log file: %s", output)
	}
	if !strings.Contains(output, ".go") {
		t.Fatalf("output missing kept go file: %s", output)
	}
}

func writeFile(t *testing.T, root string, name string, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
