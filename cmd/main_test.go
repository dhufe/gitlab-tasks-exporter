package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMainHelper is executed in a separate subprocess to call main() safely.
// It resets the default flag set and reconstructs os.Args based on the env var
// GO_HELPER_ARGS to avoid interference with the testing package's flags.
func TestMainHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Reset the global flag set so our app's flags can parse cleanly
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Rebuild os.Args as if the app was run directly
	helperArgs := os.Getenv("GO_HELPER_ARGS")
	if helperArgs != "" {
		os.Args = append([]string{"cmd"}, strings.Fields(helperArgs)...)
	} else {
		os.Args = []string{"cmd"}
	}

	// Call the real main; it will call os.Exit(...)
	main()
}

// runMain is a helper to spawn the current test binary and execute TestMainHelper
// which in turn calls the program's main().
func runMain(t *testing.T, args []string, extraEnv map[string]string) (output string, exitCode int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run", "TestMainHelper")

	// Pass down environment, override with our specific variables
	env := os.Environ()
	env = append(env,
		"GO_WANT_HELPER_PROCESS=1",
		"GO_HELPER_ARGS="+strings.Join(args, " "),
		// Disable godotenv so tests don't pick up a local .env file
		"GODOTENV_DISABLE=1",
	)
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	// Capture combined stdout/stderr
	out, err := cmd.CombinedOutput()
	output = string(out)

	if err == nil {
		return output, 0
	}

	// Extract exit code
	if exitErr, ok := err.(*exec.ExitError); ok {
		// On Unix-like systems this is the exit code
		return output, exitErr.ExitCode()
	}

	// Fallback: treat as unknown failure
	return output, -1
}

func TestMain_HelpFlag_ExitsZeroAndPrintsUsage(t *testing.T) {
	out, code := runMain(t, []string{"--help"}, map[string]string{})

	if code != 0 {
		t.Fatalf("expected exit code 0 for --help, got %d. Output: %s", code, out)
	}

	// Basic sanity checks for usage text
	if !strings.Contains(out, "GitLab zu Todoist Exporter") &&
		!strings.Contains(out, "VERWENDUNG:") {
		t.Fatalf("expected usage text in output, got: %s", out)
	}
}

func TestMain_ParseFlagsError_ExitsOneAndPrintsMessage(t *testing.T) {
	// Ensure required config is not present; prevent accidental .env use
	env := map[string]string{
		"GITLAB_TOKEN":    "", // empty
		"PROJECT_PATH":    "", // empty
		"TODOIST_TOKEN":   "",
		"TODOIST_PROJECT": "",
		"GITLAB_URL":      "", // default will be used but irrelevant
	}

	out, code := runMain(t, nil, env)

	if code != 1 {
		t.Fatalf("expected exit code 1 for parse error, got %d. Output: %s", code, out)
	}

	if !strings.Contains(out, "Fehler beim Parsen der Flags") {
		t.Fatalf("expected parse error message, got: %s", out)
	}
}

func TestMain_ExportError_ExitsOneAndPrintsMessage(t *testing.T) {
	// Provide minimal config to pass validation, but an unreachable GitLab URL
	// so the exporter fails quickly during connection validation.
	env := map[string]string{
		"GITLAB_TOKEN": "dummy-token",
		"PROJECT_PATH": "user/project",
		// Port 9 is typically closed; connection should fail immediately
		"GITLAB_URL":  "http://127.0.0.1:9",
		"TODOIST_API": "false",          // ensure we don't try Todoist API
		"OUTPUT_FILE": "test-output.md", // irrelevant, but harmless
		"VERBOSE":     "false",
	}

	out, code := runMain(t, nil, env)

	if code != 1 {
		t.Fatalf("expected exit code 1 for export error, got %d. Output: %s", code, out)
	}

	if !strings.Contains(out, "Export fehlgeschlagen") {
		t.Fatalf("expected export error message, got: %s", out)
	}
}
