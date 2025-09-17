package utils

import (
	"fmt"
	"time"
)

// ConvertToTodoistDate konvertiert GitLab ISO-Datum zu Todoist YYYY-MM-DD Format
func ConvertToTodoistDate(gitlabDate string) string {
	if gitlabDate == "" {
		return ""
	}

	// Mögliche GitLab Formate
	formats := []string{
		"2006-01-02T15:04:05Z",          // ISO UTC
		"2006-01-02T15:04:05.000Z",      // ISO UTC mit Millisekunden
		"2006-01-02T15:04:05+07:00",     // ISO mit Timezone
		"2006-01-02T15:04:05.000+07:00", // ISO mit Millisekunden + Timezone
		"2006-01-02",                    // Nur Datum
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, gitlabDate); err == nil {
			return parsedTime.Format("2006-01-02")
		}
	}

	fmt.Printf("⚠️  Unbekanntes Datumsformat: %s\n", gitlabDate)
	return gitlabDate
}

// FormatDateForDisplay formatiert Datum für schöne Anzeige
func FormatDateForDisplay(dateStr string) string {
	if dateStr == "" {
		return "Kein Datum"
	}

	if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
		return parsed.Format("02.01.2006")
	}

	return dateStr
}
