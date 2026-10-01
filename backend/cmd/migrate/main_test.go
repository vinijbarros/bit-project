package main

import "testing"

func TestParseActionAcceptsSupportedCommands(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"up", "status", "down", "version", " UP "} {
		action := action
		t.Run(action, func(t *testing.T) {
			t.Parallel()

			if _, err := parseAction([]string{action}); err != nil {
				t.Fatalf("parseAction(%q) error = %v", action, err)
			}
		})
	}
}

func TestParseActionRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	tests := [][]string{
		nil,
		{"up", "status"},
		{"redo"},
	}
	for _, args := range tests {
		if _, err := parseAction(args); err == nil {
			t.Fatalf("parseAction(%q) error = nil, want error", args)
		}
	}
}
