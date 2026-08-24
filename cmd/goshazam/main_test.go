package main

import (
	"bytes"
	"testing"
)

func TestCLICommands(t *testing.T) {
	cmd := newRootCommand()
	if cmd == nil {
		t.Fatalf("expected non-nil root command")
	}

	// Verify command registration
	commands := make(map[string]bool)
	for _, c := range cmd.Commands() {
		commands[c.Name()] = true
	}

	for _, expected := range []string{"recognize", "signature", "related", "upgrade"} {
		if !commands[expected] {
			t.Errorf("expected command %q to be registered", expected)
		}
	}

	// Verify version template output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("executing --version failed: %v", err)
	}

	out := buf.String()
	if out != version+"\n" {
		t.Errorf("expected version %q, got %q", version+"\n", out)
	}
}
