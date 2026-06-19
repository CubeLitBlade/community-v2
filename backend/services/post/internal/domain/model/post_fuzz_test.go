package model

import (
	"strings"
	"testing"
)

//nolint:revive // for test only
func FuzzCleanString(f *testing.F) {
	f.Add("hello")
	f.Add("  hello  ")
	f.Add("")
	f.Add("   ")
	f.Add("\t\n")

	f.Fuzz(func(t *testing.T, orig string) {
		input := orig
		inputPtr := &input

		out := cleanString(inputPtr)

		if out == nil {
			if strings.TrimSpace(orig) != "" {
				t.Errorf("cleanString returned nil for non-blank input: %q", orig)
			}

			return
		}

		if *out != strings.TrimSpace(orig) {
			t.Errorf("cleanString result %q != TrimSpace %q", *out, strings.TrimSpace(orig))
		}

		if *out == "" {
			t.Errorf("cleanString returned empty string pointer for input: %q", orig)
		}
	})
}
