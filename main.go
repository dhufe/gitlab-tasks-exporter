// main.go
package main

import (
	"fmt"
	"log"
)

func main() {

	config := parseFlags()
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

	if config.ExportMarkdown {
		err = exporter.ExportMarkdown(issues, config.MarkdownFile)
		if err != nil {
			log.Fatalf("Markdown-Export fehlgeschlagen: %v", err)
		}
		fmt.Printf("Markdown-Export erstellt: %s\n", config.MarkdownFile)
	}

	fmt.Printf("Export abgeschlossen: %s\n", config.OutputFile)
}
