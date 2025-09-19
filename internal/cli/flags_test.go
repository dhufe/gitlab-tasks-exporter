package cli

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Test helper that runs in a subprocess and calls ParseFlags safely.
func TestHelperProcess_ParseFlags(t *testing.T) {
	if os.Getenv("GO_WANT_PARSEFLAGS_HELPER") != "1" {
		return
	}

	// Reset global flags and args so our CLI can parse cleanly.
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	helperArgs := os.Getenv("GO_HELPER_ARGS")
	if helperArgs != "" {
		os.Args = append([]string{"gitlab-exporter"}, strings.Fields(helperArgs)...)
	} else {
		os.Args = []string{"gitlab-exporter"}
	}

	cfg, err := ParseFlags()

	// If ParseFlags returns an error (e.g., validation failed), signal with exit code 2
	if err != nil {
		// Prefix to make assertions stable in parent process
		_, err := os.Stderr.WriteString("PARSE_ERROR: " + err.Error() + "\n")
		if err != nil {
			return
		}
		os.Exit(2)
		return
	}

	// Serialize a subset of the config for assertions
	out := struct {
		GitLabURL      string `json:"gitlab_url"`
		ProjectPath    string `json:"project_path"`
		MilestoneTitle string `json:"milestone_title"`
		TodoistToken   string `json:"todoist_token"`
		TodoistProject string `json:"todoist_project"`
		TodoistAPI     bool   `json:"todoist_api"`
		OutputFile     string `json:"output_file"`
		Verbose        bool   `json:"verbose"`
	}{
		GitLabURL:      cfg.GitLabURL,
		ProjectPath:    cfg.ProjectPath,
		TodoistToken:   cfg.TodoistToken,
		TodoistProject: cfg.TodoistProject,
		TodoistAPI:     cfg.TodoistAPI,
		OutputFile:     cfg.OutputFile,
		Verbose:        cfg.Verbose,
	}
	if cfg.MilestoneTitle != nil {
		out.MilestoneTitle = *cfg.MilestoneTitle
	}

	b, _ := json.Marshal(out)
	_, err = os.Stdout.WriteString("CFG:" + string(b) + "\n")
	if err != nil {
		return
	}
	os.Exit(0)
}

// runParseFlags runs ParseFlags in a subprocess so we can capture exit code and output
// even when ParseFlags calls os.Exit (e.g., for --help).
func runParseFlags(t *testing.T, args []string, env map[string]string) (output string, exitCode int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelperProcess_ParseFlags")

	// Start with current env, then override.
	e := os.Environ()

	// Ensure godotenv won't load a local .env
	e = append(e, "GODOTENV_DISABLE=1")

	// Pass args for the helper
	e = append(e, "GO_WANT_PARSEFLAGS_HELPER=1")
	e = append(e, "GO_HELPER_ARGS="+strings.Join(args, " "))

	// Clear and set relevant variables to make behavior deterministic
	keys := []string{
		"GITLAB_TOKEN", "GITLAB_URL", "PROJECT_PATH", "MILESTONE_TITLE",
		"TODOIST_TOKEN", "TODOIST_PROJECT", "TODOIST_API", "OUTPUT_FILE", "VERBOSE",
	}
	for _, k := range keys {
		e = append(e, k+"=")
	}

	for k, v := range env {
		e = append(e, k+"="+v)
	}

	cmd.Env = e

	out, err := cmd.CombinedOutput()
	output = string(out)

	if err == nil {
		return output, 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return output, ee.ExitCode()
	}
	return output, -1
}

func TestParseFlags_Help_PrintsUsageAndExitsZero(t *testing.T) {
	out, code := runParseFlags(t, []string{"--help"}, nil)

	if code != 0 {
		t.Fatalf("expected exit code 0 for --help, got %d. Output: %s", code, out)
	}

	// Basic usage content checks
	if !strings.Contains(out, "VERWENDUNG:") || !strings.Contains(out, "CLI-OPTIONEN:") {
		t.Fatalf("expected usage text with VERWENDUNG and CLI-OPTIONEN, got: %s", out)
	}
}

func TestParseFlags_ValidationRunsBeforeCLIFlags(t *testing.T) {
	// Intentionally provide required values only via CLI flags, not env.
	// Current implementation validates BEFORE applying CLI flags, so we expect a validation error.
	out, code := runParseFlags(t,
		[]string{"--gitlab-token", "glpat-123", "--project-path", "user/repo"},
		map[string]string{},
	)

	if code != 2 {
		t.Fatalf("expected exit code 2 for validation error, got %d. Output: %s", code, out)
	}

	if !strings.Contains(out, "PARSE_ERROR:") || !strings.Contains(out, "GitLab Token fehlt") {
		t.Fatalf("expected validation error about missing GitLab token, got: %s", out)
	}
}

func TestParseFlags_AppliesCLIOverridesAfterValidation(t *testing.T) {
	// Provide minimal valid env so validation passes, then assert flags override values.
	env := map[string]string{
		"GITLAB_TOKEN":    "env-token",
		"PROJECT_PATH":    "env/project",
		"GITLAB_URL":      "https://gitlab.com",
		"TODOIST_API":     "false",
		"OUTPUT_FILE":     "env.md",
		"TODOIST_PROJECT": "GitLab Issues",
	}

	args := []string{
		"--gitlab-url", "https://example.internal/",
		"--milestone", "Release-1",
		"--todoist", "--todoist-token", "t-123",
		"--todoist-project", "MyProj",
		"--output", "out.md",
		"--verbose",
	}

	out, code := runParseFlags(t, args, env)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. Output: %s", code, out)
	}

	// Extract JSON payload following the CFG: prefix
	idx := strings.Index(out, "CFG:")
	if idx == -1 {
		t.Fatalf("expected CFG: JSON in output, got: %s", out)
	}
	payload := strings.TrimSpace(out[idx+4:])

	var got struct {
		GitLabURL      string `json:"gitlab_url"`
		ProjectPath    string `json:"project_path"`
		MilestoneTitle string `json:"milestone_title"`
		TodoistToken   string `json:"todoist_token"`
		TodoistProject string `json:"todoist_project"`
		TodoistAPI     bool   `json:"todoist_api"`
		OutputFile     string `json:"output_file"`
		Verbose        bool   `json:"verbose"`
	}

	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("failed to decode config JSON: %v. Raw: %s", err, payload)
	}

	// Assert overrides took effect
	if got.GitLabURL != "https://example.internal/" {
		t.Errorf("GitLabURL not overridden, got %q", got.GitLabURL)
	}
	if got.MilestoneTitle != "Release-1" {
		t.Errorf("MilestoneTitle not set, got %q", got.MilestoneTitle)
	}
	if !got.TodoistAPI {
		t.Errorf("expected TodoistAPI true after --todoist")
	}
	if got.TodoistToken != "t-123" {
		t.Errorf("TodoistToken not overridden, got %q", got.TodoistToken)
	}
	if got.TodoistProject != "MyProj" {
		t.Errorf("TodoistProject not overridden, got %q", got.TodoistProject)
	}
	if got.OutputFile != "out.md" {
		t.Errorf("OutputFile not overridden, got %q", got.OutputFile)
	}
	if !got.Verbose {
		t.Errorf("expected Verbose true after --verbose")
	}
}
