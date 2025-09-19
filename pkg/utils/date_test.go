package utils

import "testing"

func TestConvertToTodoistDate(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},                                        // empty
		{"2024-02-15", "2024-02-15"},                    // already YYYY-MM-DD
		{"2024-02-15T10:00:00Z", "2024-02-15"},          // ISO UTC
		{"2024-02-15T10:00:00.123Z", "2024-02-15"},      // ISO UTC ms
		{"2024-02-15T10:00:00+07:00", "2024-02-15"},     // ISO TZ
		{"2024-02-15T10:00:00.000+07:00", "2024-02-15"}, // ISO TZ ms
		{"15/02/2024", "15/02/2024"},                    // unknown format → passthrough
	}

	for i, c := range cases {
		if got := ConvertToTodoistDate(c.in); got != c.want {
			t.Fatalf("case %d: ConvertToTodoistDate(%q) = %q, want %q", i, c.in, got, c.want)
		}
	}
}

func TestFormatDateForDisplay(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "Kein Datum"},
		{"2024-02-15", "15.02.2024"},
		{"invalid", "invalid"},
	}
	for i, c := range cases {
		if got := FormatDateForDisplay(c.in); got != c.want {
			t.Fatalf("case %d: FormatDateForDisplay(%q) = %q, want %q", i, c.in, got, c.want)
		}
	}
}
