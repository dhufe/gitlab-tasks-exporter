package main

// cli.go
import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func parseFlags() Config {
	var config Config

	// Lade .env wenn vorhanden
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file.")
	}

	flag.StringVar(&config.GitLabURL, "url", getEnv("GITLAB_URL", "https://gitlab.com"), "GitLab URL")
	flag.StringVar(&config.Token, "token", getEnv("GITLAB_TOKEN", ""), "GitLab Private Token")
	flag.StringVar(&config.ProjectPath, "project", getEnv("PROJECT_PATH", ""), "Project Path (username/project)")
	flag.StringVar(&config.OutputFile, "output", getEnv("OUTPUT_FILE", "todoist_import.csv"), "Output CSV file")

	milestoneTitle := flag.String("milestone", getEnv("MILESTONE_TITLE", ""), "Milestone Title (optional, use '*' for all milestones)")
	assignedUser := flag.String("assigned", getEnv("ASSIGNED_USER", ""), "Filter by assigned user (username)")
	structured := flag.Bool("structured", false, "Create structured export with projects")

	exportMarkdown := flag.Bool("markdown", false, "Markdown-Export für Notizen")
	flag.StringVar(&config.MarkdownFile, "markdown_file", getEnv("MARKDOWN_FILE", "issues-notes.md"), "Dateiname für Markdown-Export")
	flag.Parse()

	if *milestoneTitle != "" {
		config.MilestoneTitle = milestoneTitle
	}

	if *assignedUser != "" {
		config.AssignedUser = assignedUser
	}

	config.Structured = *structured
	config.ExportMarkdown = *exportMarkdown

	if config.Token == "" || config.ProjectPath == "" {
		log.Fatal("GITLAB_TOKEN und PROJECT_PATH müssen gesetzt sein")
	}
	fmt.Print(config)
	return config
}
