package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type GitLabExporter struct {
	client *graphql.Client
	config Config
}

func (e *GitLabExporter) GetAllIssues() ([]Issue, error) {
	var allIssues []Issue
	var after *string

	for {
		var query IssuesQuery
		variables := map[string]interface{}{
			"projectPath": graphql.String(e.config.ProjectPath),
			"after":       (*graphql.String)(after),
		}

		if e.config.MilestoneTitle != nil {
			variables["milestoneTitle"] = []graphql.String{graphql.String(*e.config.MilestoneTitle)}
		} else {
			variables["milestoneTitle"] = ([]graphql.String)(nil)
		}

		err := e.client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, fmt.Errorf("GraphQL query fehler: %w", err)
		}

		issues := query.Project.Issues.Nodes
		allIssues = append(allIssues, issues...)

		if !query.Project.Issues.PageInfo.HasNextPage {
			break
		}

		after = &query.Project.Issues.PageInfo.EndCursor
	}

	return allIssues, nil
}

func (e *GitLabExporter) ExportToTodoistCSV(issues []Issue, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen der Datei: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header schreiben
	header := []string{
		"TYPE", "CONTENT", "DESCRIPTION", "PRIORITY", "INDENT",
		"AUTHOR", "RESPONSIBLE", "DATE", "DATE_LANG", "TIMEZONE", "LABELS",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("fehler beim Schreiben des Headers: %w", err)
	}

	// Issues schreiben
	for _, issue := range issues {
		record := e.convertIssueToTodoistRecord(issue)
		row := []string{
			record.Type, record.Content, record.Description, record.Priority,
			record.Indent, record.Author, record.Responsible, record.Date,
			record.DateLang, record.Timezone, record.Labels,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("fehler beim Schreiben des Issues %d: %w", issue.IID, err)
		}
	}

	return nil
}

func (e *GitLabExporter) convertIssueToTodoistRecord(issue Issue) TodoistRecord {
	// Labels extrahieren
	var labels []string
	for _, label := range issue.Labels.Nodes {
		labels = append(labels, label.Title)
	}

	// Assignee extrahieren
	responsible := ""
	if len(issue.Assignees.Nodes) > 0 {
		responsible = issue.Assignees.Nodes[0].Name
	}

	// Priorität bestimmen
	priority := getTodoistPriority(issue)

	// Erweiterte Beschreibung erstellen
	description := createEnhancedDescription(issue)

	// Datum formatieren
	date := ""
	if issue.DueDate != nil {
		date = issue.DueDate.Format("2006-01-02")
	}

	return TodoistRecord{
		Type:        "task",
		Content:     issue.Title,
		Description: description,
		Priority:    strconv.Itoa(priority),
		Indent:      "1",
		Author:      "",
		Responsible: responsible,
		Date:        date,
		DateLang:    "de",
		Timezone:    "Europe/Berlin",
		Labels:      strings.Join(labels, ","),
	}
}

func getTodoistPriority(issue Issue) int {
	// Todoist Prioritäten: 1=niedrig, 2=normal, 3=hoch, 4=urgent
	labelTitles := make([]string, len(issue.Labels.Nodes))
	for i, label := range issue.Labels.Nodes {
		labelTitles[i] = strings.ToLower(label.Title)
	}

	for _, label := range labelTitles {
		switch {
		case contains([]string{"urgent", "critical", "blocker"}, label):
			return 4
		case contains([]string{"high", "important"}, label):
			return 3
		case contains([]string{"low", "minor"}, label):
			return 1
		}
	}

	return 2 // normal
}

func createEnhancedDescription(issue Issue) string {
	var parts []string

	if issue.Description != "" {
		parts = append(parts, issue.Description)
	}

	// GitLab Link hinzufügen
	parts = append(parts, fmt.Sprintf("\n🔗 GitLab: %s", issue.WebURL))

	// Milestone info
	if issue.Milestone != nil {
		parts = append(parts, fmt.Sprintf("📋 Milestone: %s", issue.Milestone.Title))
	}

	// Weight (falls vorhanden)
	if issue.Weight != nil {
		parts = append(parts, fmt.Sprintf("⚖️ Weight: %d", *issue.Weight))
	}

	// Time tracking
	if issue.TimeStats.TimeEstimate > 0 || issue.TimeStats.TotalTimeSpent > 0 {
		parts = append(parts, fmt.Sprintf("⏱️ Time: %dh estimated, %dh spent",
			issue.TimeStats.TimeEstimate/3600, issue.TimeStats.TotalTimeSpent/3600))
	}

	return strings.Join(parts, " | ")
}

// structured_export.go
func (e *GitLabExporter) ExportToTodoistWithStructure(issues []Issue, filename string) error {
	// Gruppiere nach Milestone
	milestoneGroups := make(map[string][]Issue)

	for _, issue := range issues {
		milestoneTitle := "Ohne Milestone"
		if issue.Milestone != nil {
			milestoneTitle = issue.Milestone.Title
		}
		milestoneGroups[milestoneTitle] = append(milestoneGroups[milestoneTitle], issue)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen der Datei: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header schreiben
	header := []string{
		"TYPE", "CONTENT", "DESCRIPTION", "PRIORITY", "INDENT",
		"AUTHOR", "RESPONSIBLE", "DATE", "DATE_LANG", "TIMEZONE", "LABELS",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Für jeden Milestone
	for milestoneTitle, milestoneIssues := range milestoneGroups {
		// Projekt erstellen
		projectRow := []string{
			"project", milestoneTitle, "", "1", "1", "", "", "", "de", "Europe/Berlin", "",
		}
		if err := writer.Write(projectRow); err != nil {
			return err
		}

		// Issues als Tasks hinzufügen
		for _, issue := range milestoneIssues {
			record := e.convertIssueToTodoistRecord(issue)
			record.Indent = "2" // Unter Projekt eingerückt

			row := []string{
				record.Type, record.Content, record.Description, record.Priority,
				record.Indent, record.Author, record.Responsible, record.Date,
				record.DateLang, record.Timezone, record.Labels,
			}

			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	return nil
}
