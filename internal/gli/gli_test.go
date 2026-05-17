package gli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsSplashWithoutCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run(nil, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	if !strings.Contains(out.String(), "gli tools") {
		t.Fatalf("splash output missing title: %q", out.String())
	}
}

func TestRunPrintsHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"-h"}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	output := out.String()
	if !strings.Contains(output, "gli <command> <flags>") {
		t.Fatalf("help output missing usage: %q", output)
	}
	if !strings.Contains(output, "lines") || !strings.Contains(output, "update") {
		t.Fatalf("help output missing commands: %q", output)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"nope"}, &out, &errOut); code != 2 {
		t.Fatalf("Run() = %d, want 2", code)
	}

	if !strings.Contains(errOut.String(), `unknown command "nope"`) {
		t.Fatalf("error output missing unknown command: %q", errOut.String())
	}
}

func TestRunDispatchesHello(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"hello", "-name", "gli"}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	if got := strings.TrimSpace(out.String()); got != "hello, gli" {
		t.Fatalf("output = %q, want %q", got, "hello, gli")
	}
}
