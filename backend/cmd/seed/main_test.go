package main

import "testing"

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantReset bool
		wantError bool
	}{
		{name: "default", args: nil},
		{name: "reset", args: []string{"--reset-passwords"}, wantReset: true},
		{name: "unknown", args: []string{"--force"}, wantError: true},
		{name: "too many", args: []string{"--reset-passwords", "extra"}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseOptions(test.args)
			if (err != nil) != test.wantError {
				t.Fatalf("parseOptions() error = %v, wantError = %v", err, test.wantError)
			}
			if got.resetPasswords != test.wantReset {
				t.Fatalf("resetPasswords = %v, want %v", got.resetPasswords, test.wantReset)
			}
		})
	}
}
