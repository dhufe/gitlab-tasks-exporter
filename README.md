# GitLab Tasks Exporter 🚀

Export GitLab issues (and optionally merge requests) into either a Markdown report or directly into Todoist as tasks. Configure it via CLI flags, environment variables, or a .env file. Binaries are provided for common platforms, and you can also build from source. ✨

Current local date/time: 2025-10-02 16:35

## ✨ Features
- 📄 Export GitLab items to a Markdown file for reporting or sharing
- ✅ Export directly to Todoist via the Todoist API
- 🎯 Filter by milestone title
- ⚙️ Flexible configuration: CLI flags > environment variables > .env file
- 🐞 Verbose mode for easier troubleshooting
- 🧰 Cross‑platform builds via Makefile (Linux, macOS, Windows)

## 📦 How to use

### 1) Install 🧭
You can either use the prebuilt binaries in bin/ or build from source.

- Use an existing binary (example):
  ```text
  macOS (arm64): bin/gitlab-exporter_arm64
  macOS (amd64): bin/gitlab-exporter_darwin
  Linux (amd64): bin/gitlab-exporter_unix
  Windows (amd64): bin/gitlab-exporter.exe
  ```

- Build from source:
  - Prerequisites: Go (per go.mod, Go 1.25+)
  - Build:
    ```bash
    make build
    ```
  - The binary will be at:
    ```text
    bin/gitlab-exporter
    ```

### 2) Configure 🔧
You can configure via any of the following, with this priority:
1) CLI flags
2) Environment variables
3) .env file (loaded via github.com/joho/godotenv)

Quick setup helper:
- Create a .env and install dependencies:
  ```bash
  make setup
  ```

Environment variables supported:
```ini
# GitLab
GITLAB_TOKEN=GitLab API Token
GITLAB_URL=https://gitlab.com
PROJECT_PATH=user/repository
MILESTONE_TITLE=Optional milestone title filter

# Todoist
TODOIST_TOKEN=Todoist API Token
TODOIST_PROJECT=Todoist project name
TODOIST_API=false  # set to true to export to Todoist

# Output & Verbosity
OUTPUT_FILE=output.md
VERBOSE=false
```

### 3) Run ▶️
Basic help:
```bash
bin/gitlab-exporter --help
```

Common examples:
- Export to a Markdown file using .env configuration:
  ```bash
  bin/gitlab-exporter
  ```

- Export to Todoist using only CLI parameters:
  ```bash
  bin/gitlab-exporter --gitlab-token glpat-123 --project-path user/repo --todoist --todoist-token abc123
  ```

- Export for a specific milestone to a specific file:
  ```bash
  bin/gitlab-exporter --milestone "v1.0.0" --output milestone-v1.md
  ```

CLI flags (mirror the environment variables):
```text
--gitlab-token     GitLab API token
--gitlab-url       GitLab URL
--project-path     GitLab project path
--milestone        Milestone title filter
--todoist-token    Todoist API token
--todoist-project  Todoist project name
--todoist          Enable export to Todoist API (boolean flag)
--output           Output file for Markdown export
--verbose          Verbose mode
--help             Show usage
```

## 📦 Releases & Assets
Each release includes prebuilt archives as assets:
  - bin/releases/gitlab-exporter-linux-amd64.tar.gz (Linux amd64)
  - bin/releases/gitlab-exporter-darwin-amd64.tar.gz (macOS Intel)
  - bin/releases/gitlab-exporter-darwin-arm64.tar.gz (macOS Apple Silicon)
  - bin/releases/gitlab-exporter-windows-amd64.zip (Windows amd64)

Download from the Releases page: [Releases](../../releases)

## 🤝 How to contribute
Contributions are welcome! To get started:

1) Setup your environment
- Ensure you have Go installed (Go 1.25+)
- Clone the repository
- Run setup (creates .env and installs dependencies):
  ```bash
  make setup
  ```

2) Development workflow
- Format:
  ```bash
  make fmt
  ```
- Lint (optional if you have golangci-lint installed):
  ```bash
  make lint
  ```
- Run tests:
  ```bash
  make test  # or: go test ./...
  ```
- Build locally:
  ```bash
  make build
  ```
- Try the app:
  ```bash
  make run  # or run the binary in bin/
  ```

3) Git workflow
- Create a feature branch from main
- Commit small, focused changes with clear messages
- Ensure tests pass and code is formatted
- Open a pull request describing the motivation and changes

4) Notes
- If you add or change CLI flags or env vars, please update this README and any examples
- If your change affects behavior, add or update tests in the relevant internal/... packages

Thank you for contributing! 🙏

## 📜 License
This project is licensed under the MIT License. See the LICENSE file for details.