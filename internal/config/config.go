package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// GitLab
	GitLabToken    string
	GitLabURL      string
	ProjectPath    string
	MilestoneTitle *string

	// Todoist
	TodoistToken   string
	TodoistProject string
	TodoistAPI     bool

	// Output
	OutputFile string
	Verbose    bool
}

// NewConfig erstellt eine neue Konfiguration mit .env-Support
func NewConfig() (*Config, error) {
	cfg := &Config{
		GitLabURL: "https://gitlab.com", // Default
	}

	// 1. .env-Datei laden (falls vorhanden)
	if err := cfg.LoadFromEnvFile(".env"); err != nil {
		// .env optional - Fehler nur loggen, nicht stoppen
		if !os.IsNotExist(err) {
			fmt.Printf("⚠️  Warnung: .env-Datei konnte nicht gelesen werden: %v\n", err)
		}
	}

	// 2. System-Umgebungsvariablen laden (überschreibt .env)
	cfg.LoadFromEnv()

	return cfg, nil
}

// LoadFromEnvFile lädt Konfiguration aus .env-Datei
func (c *Config) LoadFromEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Fehler beim Schliessen der Datei: %s", filename)
		}
	}()

	fmt.Printf("📄 Lade Konfiguration aus %s...\n", filename)

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Leere Zeilen und Kommentare überspringen
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// KEY=VALUE parsen
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("⚠️  Ungültige Zeile %d in %s: %s\n", lineNumber, filename, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Anführungszeichen entfernen (optional)
		value = strings.Trim(value, "\"'")

		// Wert setzen (nur wenn noch nicht gesetzt)
		c.setEnvValue(key, value, false)
	}

	return scanner.Err()
}

// LoadFromEnv lädt Konfiguration aus System-Umgebungsvariablen
func (c *Config) LoadFromEnv() {
	envVars := map[string]*string{
		"GITLAB_TOKEN":    &c.GitLabToken,
		"GITLAB_URL":      &c.GitLabURL,
		"PROJECT_PATH":    &c.ProjectPath,
		"MILESTONE_TITLE": nil, // Special handling
		"TODOIST_TOKEN":   &c.TodoistToken,
		"TODOIST_PROJECT": &c.TodoistProject,
		"OUTPUT_FILE":     &c.OutputFile,
	}

	for envKey, configField := range envVars {
		if value := os.Getenv(envKey); value != "" {
			if configField != nil {
				*configField = value
			}
		}
	}

	// Special handling für MILESTONE_TITLE
	if milestone := os.Getenv("MILESTONE_TITLE"); milestone != "" {
		c.MilestoneTitle = &milestone
	}

	// Boolean-Werte
	if todoist := os.Getenv("TODOIST_API"); todoist != "" {
		c.TodoistAPI, _ = strconv.ParseBool(todoist)
	}

	if verbose := os.Getenv("VERBOSE"); verbose != "" {
		c.Verbose, _ = strconv.ParseBool(verbose)
	}
}

// setEnvValue setzt einen Konfigurationswert basierend auf dem Key
func (c *Config) setEnvValue(key, value string, overwrite bool) {
	switch key {
	case "GITLAB_TOKEN":
		if c.GitLabToken == "" || overwrite {
			c.GitLabToken = value
		}
	case "GITLAB_URL":
		if c.GitLabURL == "" || overwrite {
			c.GitLabURL = value
		}
	case "PROJECT_PATH":
		if c.ProjectPath == "" || overwrite {
			c.ProjectPath = value
		}
	case "MILESTONE_TITLE":
		if c.MilestoneTitle == nil || overwrite {
			c.MilestoneTitle = &value
		}
	case "TODOIST_TOKEN":
		if c.TodoistToken == "" || overwrite {
			c.TodoistToken = value
		}
	case "TODOIST_PROJECT":
		if c.TodoistProject == "" || overwrite {
			c.TodoistProject = value
		}
	case "TODOIST_API":
		if parsed, err := strconv.ParseBool(value); err == nil {
			c.TodoistAPI = parsed
		}
	case "OUTPUT_FILE":
		if c.OutputFile == "" || overwrite {
			c.OutputFile = value
		}
	case "VERBOSE":
		if parsed, err := strconv.ParseBool(value); err == nil {
			c.Verbose = parsed
		}
	}
}

// Rest der Config-Methoden bleiben gleich...
func (c *Config) Validate() error {
	if c.GitLabToken == "" {
		return fmt.Errorf("GitLab Token fehlt (--gitlab-token oder GITLAB_TOKEN)")
	}

	if c.ProjectPath == "" {
		return fmt.Errorf("GitLab Projekt-Pfad fehlt (--project-path oder PROJECT_PATH)")
	}

	if c.TodoistAPI && c.TodoistToken == "" {
		return fmt.Errorf("todoist Token fehlt für API-Export (--todoist-token oder TODOIST_TOKEN)")
	}

	return nil
}

func (c *Config) GetGitLabBaseURL() string {
	return strings.TrimSuffix(c.GitLabURL, "/")
}

func (c *Config) GetTodoistBaseURL() string {
	return "https://api.todoist.com/rest/v2"
}
