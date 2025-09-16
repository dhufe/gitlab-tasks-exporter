package main

// cli.go
import (
	"flag"
	"fmt"
)

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.GitLabURL, "url", getEnv("GITLAB_URL", "https://gitlab.com"), "GitLab URL")
	flag.StringVar(&config.Token, "token", getEnv("GITLAB_TOKEN", ""), "GitLab Private Token")
	flag.StringVar(&config.ProjectPath, "project", getEnv("PROJECT_PATH", ""), "Project Path (username/project)")
	flag.StringVar(&config.OutputFile, "output", getEnv("OUTPUT_FILE", "todoist_import.csv"), "Output CSV file")

	milestoneTitle := flag.String("milestone", getEnv("MILESTONE_TITLE", ""), "Milestone Title (optional, use '*' for all milestones)")
	assignedUser := flag.String("assigned", getEnv("ASSIGNED_USER", ""), "Filter by assigned user (username)")
	structured := flag.Bool("structured", false, "Create structured export with projects")

	flag.Parse()

	if *milestoneTitle != "" {
		config.MilestoneTitle = milestoneTitle
	}

	if *assignedUser != "" {
		config.AssignedUser = assignedUser
	}

	config.Structured = *structured
	fmt.Println(config)
	return config
}
