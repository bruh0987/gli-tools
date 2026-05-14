package gli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsHelpWithoutCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run(nil, &out, &errOut, "test"); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	if !strings.Contains(out.String(), "gli <command> <flags>") {
		t.Fatalf("help output missing usage: %q", out.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"nope"}, &out, &errOut, "test"); code != 2 {
		t.Fatalf("Run() = %d, want 2", code)
	}

	if !strings.Contains(errOut.String(), `unknown command "nope"`) {
		t.Fatalf("error output missing unknown command: %q", errOut.String())
	}
}

func TestRunDispatchesHello(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"hello", "-name", "gli"}, &out, &errOut, "test"); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	if got := strings.TrimSpace(out.String()); got != "hello, gli" {
		t.Fatalf("output = %q, want %q", got, "hello, gli")
	}
}
