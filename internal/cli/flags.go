package cli

import (
	"flag"
	"fmt"
	"os"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
)

// ParseFlags parst Command-Line Arguments und ENV-Konfiguration
func ParseFlags() (*config.Config, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	// 2. CLI-Flags definieren (überschreiben ENV-Werte)
	var (
		gitlabToken    = flag.String("gitlab-token", cfg.GitLabToken, "GitLab API Token (oder GITLAB_TOKEN)")
		gitlabURL      = flag.String("gitlab-url", cfg.GitLabURL, "GitLab URL (oder GITLAB_URL)")
		projectPath    = flag.String("project-path", cfg.ProjectPath, "GitLab Projekt-Pfad (oder PROJECT_PATH)")
		milestoneTitle = flag.String("milestone", "", "Milestone-Filter (oder MILESTONE_TITLE)")
		todoistToken   = flag.String("todoist-token", cfg.TodoistToken, "Todoist API Token (oder TODOIST_TOKEN)")
		todoistProject = flag.String("todoist-project", cfg.TodoistProject, "Todoist Projekt-Name (oder TODOIST_PROJECT)")
		todoistAPI     = flag.Bool("todoist", cfg.TodoistAPI, "Export zu Todoist API (oder TODOIST_API=true)")
		outputFile     = flag.String("output", cfg.OutputFile, "Output-Datei für Markdown-Export (oder OUTPUT_FILE)")
		verbose        = flag.Bool("verbose", cfg.Verbose, "Verbose-Modus (oder VERBOSE=true)")
		help           = flag.Bool("help", false, "Hilfe anzeigen")
	)

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	// 3. CLI-Flags anwenden (überschreiben .env-Werte)
	if *gitlabToken != "" {
		cfg.GitLabToken = *gitlabToken
	}
	if *gitlabURL != "" {
		cfg.GitLabURL = *gitlabURL
	}
	if *projectPath != "" {
		cfg.ProjectPath = *projectPath
	}
	if *milestoneTitle != "" {
		cfg.MilestoneTitle = milestoneTitle
	}
	if *todoistToken != "" {
		cfg.TodoistToken = *todoistToken
	}
	if *todoistProject != "" {
		cfg.TodoistProject = *todoistProject
	}
	cfg.TodoistAPI = *todoistAPI
	if *outputFile != "" {
		cfg.OutputFile = *outputFile
	}
	cfg.Verbose = *verbose

	return cfg, nil
}

func printUsage() {
	fmt.Println(`GitLab zu Todoist Exporter

VERWENDUNG:
  gitlab-exporter [OPTIONS]

KONFIGURATION:
  Die Konfiguration kann über CLI-Flags, Umgebungsvariablen oder .env-Datei erfolgen.
  Priorität: CLI-Flags > ENV-Variablen > .env-Datei

ENV-DATEI BEISPIEL (.env):
  GITLAB_TOKEN=glpat-your-token
  PROJECT_PATH=user/project
  TODOIST_TOKEN=your-todoist-token
  TODOIST_API=true

CLI-OPTIONEN:`)

	flag.PrintDefaults()

	fmt.Println(`
BEISPIELE:
  # Export zu Markdown-Datei (mit .env-Konfiguration)
  gitlab-exporter

  # Export zu Todoist mit CLI-Parametern
  gitlab-exporter --gitlab-token glpat-123 --project-path user/repo --todoist --todoist-token abc123

  # Nur bestimmtes Milestone
  gitlab-exporter --milestone "v1.0.0" --output milestone-v1.md

ENV-VARIABLEN:
  GITLAB_TOKEN     GitLab API Token
  GITLAB_URL       GitLab URL (default: https://gitlab.com)
  PROJECT_PATH     GitLab Projekt-Pfad (user/repository)
  MILESTONE_TITLE  Milestone-Filter
  TODOIST_TOKEN    Todoist API Token
  TODOIST_PROJECT  Todoist Projekt-Name
  TODOIST_API      Export zu Todoist (true/false)
  OUTPUT_FILE      Output-Datei für Markdown-Export
  VERBOSE          Verbose-Modus (true/false)`)
}
