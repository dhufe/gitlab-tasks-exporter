package utils

import (
	"strings"
)

// EscapeMarkdown escaped spezielle Markdown-Zeichen
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		".", "\\.",
		"!", "\\!",
		"|", "\\|",
	)
	return replacer.Replace(text)
}

// TruncateText kürzt Text auf maximale Länge
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	if maxLength <= 3 {
		return text[:maxLength]
	}

	return text[:maxLength-3] + "..."
}

// FormatLabels formatiert Labels als Markdown-Tags
func FormatLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}

	var formatted []string
	for _, label := range labels {
		formatted = append(formatted, "`"+label+"`")
	}

	return strings.Join(formatted, " ")
}
