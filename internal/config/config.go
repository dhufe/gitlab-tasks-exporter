package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GitLabToken    string
	GitLabURL      string
	ProjectPath    string
	MilestoneTitle *string
	TodoistToken   string
	TodoistProject string
	TodoistAPI     bool
	OutputFile     string
	Verbose        bool
}

func NewConfig() (*Config, error) {
	// .env laden (ignoriere Fehler wenn Datei nicht existiert)
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Printf("⚠️  Warnung beim Laden der .env: %v\n", err)
	}

	cfg := &Config{
		GitLabToken:    getEnv("GITLAB_TOKEN", ""),
		GitLabURL:      getEnv("GITLAB_URL", "https://gitlab.com"),
		ProjectPath:    getEnv("PROJECT_PATH", ""),
		TodoistToken:   getEnv("TODOIST_TOKEN", ""),
		TodoistProject: getEnv("TODOIST_PROJECT", "GitLab Issues"),
		TodoistAPI:     getBoolEnv("TODOIST_API", false),
		OutputFile:     getEnv("OUTPUT_FILE", "gitlab_issues.md"),
		Verbose:        getBoolEnv("VERBOSE", false),
	}

	// Optional: MILESTONE_TITLE
	if milestone := os.Getenv("MILESTONE_TITLE"); milestone != "" {
		cfg.MilestoneTitle = &milestone
	}

	if cfg.Verbose {
		cfg.printDebugInfo()
	}

	return cfg, nil
}

func (c *Config) printDebugInfo() {
	fmt.Printf("🔧 Configuration loaded:\n")
	fmt.Printf("   GitLab URL: %s\n", c.GitLabURL)
	fmt.Printf("   Project Path: %s\n", c.ProjectPath)
	fmt.Printf("   Output File: %s\n", c.OutputFile)
	fmt.Printf("   Has GitLab Token: %t (length: %d)\n",
		c.GitLabToken != "", len(c.GitLabToken))
	fmt.Printf("   Has Todoist Token: %t\n", c.TodoistToken != "")
	if c.MilestoneTitle != nil {
		fmt.Printf("   Milestone Filter: %s\n", *c.MilestoneTitle)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.GitLabToken == "" {
		return fmt.Errorf("GitLab Token fehlt (GITLAB_TOKEN)")
	}
	if c.ProjectPath == "" {
		return fmt.Errorf("GitLab Projekt-Pfad fehlt (PROJECT_PATH)")
	}
	if c.TodoistAPI && c.TodoistToken == "" {
		return fmt.Errorf("todoist Token fehlt für API-Export (TODOIST_TOKEN)")
	}
	return nil
}

func (c *Config) GetGitLabBaseURL() string {
	return strings.TrimSuffix(c.GitLabURL, "/")
}

func (c *Config) GetTodoistBaseURL() string {
	return "https://api.todoist.com/rest/v2"
}
