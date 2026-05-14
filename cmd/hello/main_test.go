package main

import "testing"

func TestRunRejectsUnknownFlag(t *testing.T) {
	if code := run([]string{"-nope"}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunAcceptsName(t *testing.T) {
	if code := run([]string{"-name", "gli"}); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}
