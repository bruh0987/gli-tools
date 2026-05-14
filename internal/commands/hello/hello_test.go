package hello

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRejectsUnknownFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"-nope"}, &out, &errOut); code != 2 {
		t.Fatalf("Run() = %d, want 2", code)
	}
}

func TestRunAcceptsName(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"-name", "gli"}, &out, &errOut); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	if got := strings.TrimSpace(out.String()); got != "hello, gli" {
		t.Fatalf("output = %q, want %q", got, "hello, gli")
	}
}
