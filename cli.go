package main

// cli.go
import (
	"flag"
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

	flag.StringVar(&config.TodoistToken, "todoist-token", getEnv("TODOIST_TOKEN", ""), "Todoist API Token")
	useAPI := flag.Bool("todoist-api", false, "Use Todoist API instead of CSV export")
	flag.StringVar(&config.TodoistProject, "todoist-project", getEnv("TODOIST_PROJECT", ""), "Todoist Project Name (will be created if not exists)")

	flag.Parse()

	if *milestoneTitle != "" {
		config.MilestoneTitle = milestoneTitle
	}

	if *assignedUser != "" {
		config.AssignedUser = assignedUser
	}

	config.Structured = *structured
	config.ExportMarkdown = *exportMarkdown
	config.UseAPI = *useAPI

	if config.Token == "" || config.ProjectPath == "" {
		log.Fatal("GITLAB_TOKEN und PROJECT_PATH müssen gesetzt sein")
	}

	if config.UseAPI && config.TodoistToken == "" {
		log.Fatal("TODOIST_TOKEN muss gesetzt sein wenn --todoist-api verwendet wird")
	}
	return config
}
