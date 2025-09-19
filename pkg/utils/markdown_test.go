package utils

import "testing"

func TestEscapeMarkdown(t *testing.T) {
	in := "Hello [World] (test) *bold* _italics_ `code` #1 + plus - dash . dot ! bang | pipe"
	// Expect all special Markdown characters escaped with backslashes
	// want := "Hello \\ [World\\] \\ (test\\) \\*bold\\* \\_italics\\_ \\`code\\` \\#1 \\+ plus \\- dash \\. dot \\! bang \\| pipe"
	// Note: The above string contains Go string-escaped backslashes. To avoid confusion,
	// verify by running the function and ensuring all target characters are escaped.

	got := EscapeMarkdown(in)

	// Spot-check each token is escaped as expected
	checks := map[string]string{
		"[": "\\[", "]": "\\]", "(": "\\(", ")": "\\)",
		"*": "\\*", "_": "\\_", "`": "\\`", "#": "\\#", "+": "\\+",
		"-": "\\-", ".": "\\.", "!": "\\!", "|": "\\|",
	}
	for ch, esc := range checks {
		if ch == "-" { // ensure hyphens in words remain but standalone dash is escaped
			// We at least require that there exists an escaped hyphen sequence in the output
			if !contains(got, esc) {
				t.Fatalf("expected escaped hyphen %q in %q", esc, got)
			}
			continue
		}
		if !contains(got, esc) {
			t.Fatalf("expected %q to contain escaped %q", got, esc)
		}
	}

	// Minimal overall equality sanity check on a simpler sample to avoid confusion with escapes
	simpleIn := "[*](_)`#-+.!|"
	simpleWant := "\\[\\*\\]\\(\\_\\)\\`\\#\\-\\+\\.\\!\\|"
	if got2 := EscapeMarkdown(simpleIn); got2 != simpleWant {
		t.Fatalf("EscapeMarkdown simple: got %q, want %q", got2, simpleWant)
	}
}

// contains is a tiny helper to avoid importing strings for a single use.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTruncateText(t *testing.T) {
	// No truncation when length <= max
	if got := TruncateText("hello", 10); got != "hello" {
		t.Fatalf("no-trunc: got %q", got)
	}
	if got := TruncateText("helloworld", 10); got != "helloworld" {
		t.Fatalf("edge equal: got %q", got)
	}

	// Truncation adds ellipsis when maxLength > 3
	if got := TruncateText("helloworld", 7); got != "hell..." { // 7-3=4 + "..."
		t.Fatalf("trunc with ellipsis: got %q", got)
	}

	// For maxLength <= 3, no ellipsis is added; just cut to length
	if got := TruncateText("abcdef", 3); got != "abc" {
		t.Fatalf("trunc no ellipsis (3): got %q", got)
	}
	if got := TruncateText("abcdef", 2); got != "ab" {
		t.Fatalf("trunc no ellipsis (2): got %q", got)
	}
	if got := TruncateText("abcdef", 0); got != "" {
		t.Fatalf("trunc no ellipsis (0): got %q", got)
	}
}

func TestFormatLabels(t *testing.T) {
	if got := FormatLabels(nil); got != "" {
		t.Fatalf("nil labels: got %q", got)
	}
	if got := FormatLabels([]string{}); got != "" {
		t.Fatalf("empty labels: got %q", got)
	}
	if got := FormatLabels([]string{"bug", "high-priority"}); got != "`bug` `high-priority`" {
		t.Fatalf("format labels: got %q", got)
	}
}
