// main.go
package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Lade .env wenn vorhanden
	godotenv.Load()

	config := parseFlags()
	if milestone := getEnv("MILESTONE_TITLE", ""); milestone != "" {
		config.MilestoneTitle = &milestone
	}
	if assignedUser := getEnv("ASSIGNED_USER", ""); assignedUser != "" {
		config.AssignedUser = &assignedUser
	}

	if config.Token == "" || config.ProjectPath == "" {
		log.Fatal("GITLAB_TOKEN und PROJECT_PATH müssen gesetzt sein")
	}

	exporter := NewGitLabExporter(config)

	fmt.Printf("Exportiere Issues aus %s...\n", config.ProjectPath)
	if config.MilestoneTitle != nil {
		if *config.MilestoneTitle == "*" {
			fmt.Println("Alle Milestones")
		} else {
			fmt.Printf("Milestone: %s\n", *config.MilestoneTitle)
		}
	}
	if config.AssignedUser != nil {
		fmt.Printf("Assigned to: %s\n", *config.AssignedUser)
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
