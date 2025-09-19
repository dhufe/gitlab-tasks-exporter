package config

import (
	"testing"
)

// helper to construct a config with a clean environment.
func newConfigWithEnv(t *testing.T, env map[string]string) *Config {
	t.Helper()

	// Ensure godotenv does not load a developer's local .env
	t.Setenv("GODOTENV_DISABLE", "1")

	// Clear all relevant variables first (empty → defaults will be used)
	keys := []string{
		"GITLAB_TOKEN", "GITLAB_URL", "PROJECT_PATH", "MILESTONE_TITLE",
		"TODOIST_TOKEN", "TODOIST_PROJECT", "TODOIST_API", "OUTPUT_FILE", "VERBOSE",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}

	// Apply overrides for this test
	for k, v := range env {
		t.Setenv(k, v)
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
	return cfg
}

func TestNewConfig_Defaults_NoEnv(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{})

	if cfg.GitLabToken != "" {
		t.Errorf("expected empty GitLabToken, got %q", cfg.GitLabToken)
	}
	if cfg.GitLabURL != "https://gitlab.com" {
		t.Errorf("expected default GitLabURL, got %q", cfg.GitLabURL)
	}
	if cfg.ProjectPath != "" {
		t.Errorf("expected empty ProjectPath, got %q", cfg.ProjectPath)
	}
	if cfg.MilestoneTitle != nil {
		t.Errorf("expected nil MilestoneTitle, got %v", cfg.MilestoneTitle)
	}
	if cfg.TodoistToken != "" {
		t.Errorf("expected empty TodoistToken, got %q", cfg.TodoistToken)
	}
	if cfg.TodoistProject != "GitLab Issues" {
		t.Errorf("expected default TodoistProject, got %q", cfg.TodoistProject)
	}
	if cfg.TodoistAPI {
		t.Errorf("expected TodoistAPI false by default")
	}
	if cfg.OutputFile != "gitlab_issues.md" {
		t.Errorf("expected default OutputFile, got %q", cfg.OutputFile)
	}
	if cfg.Verbose {
		t.Errorf("expected Verbose false by default")
	}
}

func TestNewConfig_WithEnvValues(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{
		"GITLAB_TOKEN":    "glpat-123",
		"GITLAB_URL":      "https://example.gitlab.local/",
		"PROJECT_PATH":    "user/repo",
		"MILESTONE_TITLE": "v1.2.3",
		"TODOIST_TOKEN":   "todo-xyz",
		"TODOIST_PROJECT": "My Todoist Project",
		"TODOIST_API":     "true",
		"OUTPUT_FILE":     "out.md",
		"VERBOSE":         "true",
	})

	if cfg.GitLabToken != "glpat-123" {
		t.Errorf("GitLabToken mismatch: %q", cfg.GitLabToken)
	}
	if cfg.GitLabURL != "https://example.gitlab.local/" {
		t.Errorf("GitLabURL mismatch: %q", cfg.GitLabURL)
	}
	if cfg.ProjectPath != "user/repo" {
		t.Errorf("ProjectPath mismatch: %q", cfg.ProjectPath)
	}
	if cfg.MilestoneTitle == nil || *cfg.MilestoneTitle != "v1.2.3" {
		t.Fatalf("MilestoneTitle mismatch: %#v", cfg.MilestoneTitle)
	}
	if cfg.TodoistToken != "todo-xyz" {
		t.Errorf("TodoistToken mismatch: %q", cfg.TodoistToken)
	}
	if cfg.TodoistProject != "My Todoist Project" {
		t.Errorf("TodoistProject mismatch: %q", cfg.TodoistProject)
	}
	if !cfg.TodoistAPI {
		t.Errorf("expected TodoistAPI true")
	}
	if cfg.OutputFile != "out.md" {
		t.Errorf("OutputFile mismatch: %q", cfg.OutputFile)
	}
	if !cfg.Verbose {
		t.Errorf("expected Verbose true")
	}
}

func TestValidate_MissingGitLabToken(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{
		// Only project path set, token missing
		"PROJECT_PATH": "user/repo",
	})

	if err := cfg.Validate(); err == nil || err.Error() != "GitLab Token fehlt (GITLAB_TOKEN)" {
		t.Fatalf("expected missing token error, got: %v", err)
	}
}

func TestValidate_MissingProjectPath(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{
		"GITLAB_TOKEN": "glpat-123",
	})

	if err := cfg.Validate(); err == nil || err.Error() != "GitLab Projekt-Pfad fehlt (PROJECT_PATH)" {
		t.Fatalf("expected missing project path error, got: %v", err)
	}
}

func TestValidate_TodoistAPINeedsToken(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{
		"GITLAB_TOKEN": "glpat-123",
		"PROJECT_PATH": "user/repo",
		"TODOIST_API":  "true",
		// TODOIST_TOKEN intentionally missing
	})

	if err := cfg.Validate(); err == nil || err.Error() != "todoist Token fehlt für API-Export (TODOIST_TOKEN)" {
		t.Fatalf("expected todoist token required error, got: %v", err)
	}
}

func TestGetGitLabBaseURL_TrimSuffix(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{
		"GITLAB_URL": "https://example.com/",
	})

	got := cfg.GetGitLabBaseURL()
	want := "https://example.com"
	if got != want {
		t.Fatalf("GetGitLabBaseURL(): got %q, want %q", got, want)
	}
}

func TestGetTodoistBaseURL(t *testing.T) {
	cfg := newConfigWithEnv(t, map[string]string{})
	if got := cfg.GetTodoistBaseURL(); got != "https://api.todoist.com/rest/v2" {
		t.Fatalf("GetTodoistBaseURL(): got %q", got)
	}
}
