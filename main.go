// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

func main() {

	// Lade .env wenn vorhanden
	godotenv.Load()

	config := parseFlags()
	if milestone := getEnv("MILESTONE_TITLE", ""); milestone != "" {
		config.MilestoneTitle = &milestone
	}

	if config.Token == "" || config.ProjectPath == "" {
		log.Fatal("GITLAB_TOKEN und PROJECT_PATH müssen gesetzt sein")
	}

	exporter := NewGitLabExporter(config)

	fmt.Printf("Exportiere Issues aus %s...\n", config.ProjectPath)
	if config.MilestoneTitle != nil {
		fmt.Printf("Milestone: %s\n", *config.MilestoneTitle)
	}

	issues, err := exporter.GetAllIssues()
	if err != nil {
		log.Fatalf("Fehler beim Abrufen der Issues: %v", err)
	}

	fmt.Printf("Gefunden: %d Issues\n", len(issues))

	err = exporter.ExportToTodoistCSV(issues, config.OutputFile)
	if err != nil {
		log.Fatalf("Fehler beim CSV-Export: %v", err)
	}

	fmt.Printf("Export abgeschlossen: %s\n", config.OutputFile)
}

// Korrigierte Version ohne oauth2 (einfacher mit Header)
func NewGitLabExporter(config Config) *GitLabExporter {
	httpClient := &http.Client{
		Transport: &authTransport{
			token: config.Token,
			base:  http.DefaultTransport,
		},
	}

	client := graphql.NewClient(config.GitLabURL+"/api/graphql", httpClient)

	return &GitLabExporter{
		client: client,
		config: config,
	}
}

type authTransport struct {
	token string
	base  http.RoundTripper
}
