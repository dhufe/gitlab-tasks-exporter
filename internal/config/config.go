package config

import "fmt"

type Config struct {
	// GitLab
	GitLabURL      string
	GitLabToken    string
	ProjectPath    string
	MilestoneTitle *string

	// Todoist
	TodoistToken   string
	TodoistProject string
	TodoistAPI     bool

	// Output
	Structured bool
	Markdown   bool
	OutputFile string
}

func (c *Config) Validate() error {
	if c.ProjectPath == "" {
		return fmt.Errorf("project path ist erforderlich")
	}

	if c.GitLabToken == "" {
		return fmt.Errorf("gitLab token ist erforderlich")
	}

	if c.TodoistAPI && c.TodoistToken == "" {
		return fmt.Errorf("todoist token ist erforderlich für API-Export")
	}

	return nil
}

func (c *Config) GetGitLabBaseURL() string {
	if c.GitLabURL == "" {
		return "https://gitlab.com"
	}
	return c.GitLabURL
}

func (c *Config) GetOutputFile() string {
	if c.OutputFile == "" {
		return "gitlab_issues.csv"
	}
	return c.OutputFile
}
