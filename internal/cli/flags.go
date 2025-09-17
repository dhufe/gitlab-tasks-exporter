package cli

import (
	"flag"
	"fmt"
	"os"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
)

func ParseFlags() (*config.Config, error) {
	cfg := &config.Config{}

	var milestoneTitle string

	flag.StringVar(&cfg.GitLabURL, "gitlab-url", "https://gitlab.com", "GitLab URL")
	flag.StringVar(&cfg.GitLabToken, "gitlab-token", "", "GitLab Access Token")
	flag.StringVar(&cfg.ProjectPath, "project", "", "GitLab Projekt Pfad (z.B. 'user/repo')")
	flag.StringVar(&milestoneTitle, "milestone", "*", "Milestone Titel oder '*' für alle")

	flag.StringVar(&cfg.TodoistToken, "todoist-token", "", "Todoist API Token")
	flag.StringVar(&cfg.TodoistProject, "todoist-project", "", "Todoist Projekt Name (optional)")
	flag.BoolVar(&cfg.TodoistAPI, "todoist-api", false, "Direkt zu Todoist exportieren")

	flag.BoolVar(&cfg.Structured, "structured", false, "Strukturierter Export mit Sections")
	flag.BoolVar(&cfg.Markdown, "markdown", false, "Export als Markdown statt CSV")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output Datei (Standard: gitlab_issues.csv)")

	flag.Parse()

	// Milestone handling
	if milestoneTitle != "*" && milestoneTitle != "" {
		cfg.MilestoneTitle = &milestoneTitle
	}

	// Environment variables als Fallback
	if cfg.GitLabToken == "" {
		cfg.GitLabToken = os.Getenv("GITLAB_TOKEN")
	}

	if cfg.TodoistToken == "" {
		cfg.TodoistToken = os.Getenv("TODOIST_TOKEN")
	}

	if err := cfg.Validate(); err != nil {
		flag.Usage()
		return nil, err
	}

	return cfg, nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `GitLab zu Todoist Exporter

Usage: %s [OPTIONS]

Beispiele:
  # CSV Export aller Issues
  %s -project "user/repo" -gitlab-token "glpat-xxxx"
  
  # Bestimmter Milestone als Markdown
  %s -project "user/repo" -milestone "v1.0" -markdown
  
  # Direkt zu Todoist exportieren
  %s -project "user/repo" -todoist-api -todoist-token "xxxx"

Optionen:
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Environment Variables:
  GITLAB_TOKEN    GitLab Access Token
  TODOIST_TOKEN   Todoist API Token
`)
	}
}
