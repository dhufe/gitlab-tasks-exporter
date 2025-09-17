package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type GitLabExporter struct {
	client *graphql.Client
	config Config
}

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

func (e *GitLabExporter) GetAllIssues() ([]Issue, error) {
	var allIssues []Issue
	var after *string
	first := 100

	for {
		var query ProjectQuery
		variables := map[string]interface{}{
			"projectPath": graphql.ID(e.config.ProjectPath),
			"first":       graphql.Int(first),
			"after":       (*graphql.String)(after),
		}

		// Milestone Filter
		if e.config.MilestoneTitle != nil && *e.config.MilestoneTitle != "" && *e.config.MilestoneTitle != "*" {
			variables["milestoneSearch"] = graphql.String(*e.config.MilestoneTitle)
			variables["milestoneTitle"] = []graphql.String{graphql.String(*e.config.MilestoneTitle)}
		} else {
			variables["milestoneSearch"] = (*graphql.String)(nil)
			variables["milestoneTitle"] = []graphql.String{}
		}

		// Assigned User Filter - als einzelner String, nicht Array
		if e.config.AssignedUser != nil && *e.config.AssignedUser != "" {
			variables["assigneeUsername"] = graphql.String(*e.config.AssignedUser)
		} else {
			variables["assigneeUsername"] = (*graphql.String)(nil)
		}

		fmt.Printf("Führe GraphQL Query aus... (after: %v)\n", after)

		err := e.client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, fmt.Errorf("GraphQL query fehler: %w", err)
		}

		issues := query.Project.Issues.Nodes
		fmt.Printf("Batch gefunden: %d Issues\n", len(issues))
		allIssues = append(allIssues, issues...)

		if !query.Project.Issues.PageInfo.HasNextPage {
			break
		}

		after = query.Project.Issues.PageInfo.EndCursor
		if after == nil {
			break
		}
	}

	return allIssues, nil
}

func (e *GitLabExporter) writeStructuredCSV(writer *csv.Writer, issues []Issue) error {
	milestoneTitle := "GitLab Issues"
	if e.config.MilestoneTitle != nil {
		milestoneTitle = *e.config.MilestoneTitle
	}

	projectRow := []string{
		"project", milestoneTitle, "", "1", "1", "", "", "", "de", "Europe/Berlin", "",
	}
	if err := writer.Write(projectRow); err != nil {
		return err
	}

	openIssues := []Issue{}
	closedIssues := []Issue{}

	for _, issue := range issues {
		if issue.State == "opened" {
			openIssues = append(openIssues, issue)
		} else {
			closedIssues = append(closedIssues, issue)
		}
	}

	if len(openIssues) > 0 {
		sectionRow := []string{
			"section", "🔓 Offen", "", "1", "2", "", "", "", "de", "Europe/Berlin", "",
		}
		if err := writer.Write(sectionRow); err != nil {
			return err
		}

		for _, issue := range openIssues {
			record := e.convertIssueToTodoistRecord(issue)
			record.Indent = "3"

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

	if len(closedIssues) > 0 {
		sectionRow := []string{
			"section", "✅ Geschlossen", "", "1", "2", "", "", "", "de", "Europe/Berlin", "",
		}
		if err := writer.Write(sectionRow); err != nil {
			return err
		}

		for _, issue := range closedIssues {
			record := e.convertIssueToTodoistRecord(issue)
			record.Indent = "3"

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

func (e *GitLabExporter) ExportToTodoistCSV(issues []Issue, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen der Datei: %w", err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Print("fehler beim Schliessen der Datei.")
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"TYPE", "CONTENT", "DESCRIPTION", "PRIORITY", "INDENT",
		"AUTHOR", "RESPONSIBLE", "DATE", "DATE_LANG", "TIMEZONE", "LABELS",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("fehler beim Schreiben des Headers: %w", err)
	}

	if e.config.Structured {
		return e.writeStructuredCSV(writer, issues)
	}

	for _, issue := range issues {
		record := e.convertIssueToTodoistRecord(issue)
		row := []string{
			record.Type, record.Content, record.Description, record.Priority,
			record.Indent, record.Author, record.Responsible, record.Date,
			record.DateLang, record.Timezone, record.Labels,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("fehler beim Schreiben des Issues %s: %w", issue.IID, err)
		}
	}

	return nil
}

func (e *GitLabExporter) convertIssueToTodoistRecord(issue Issue) TodoistRecord {
	var labels []string
	for _, label := range issue.Labels.Nodes {
		labels = append(labels, label.Title)
	}

	responsible := ""
	if len(issue.Assignees.Nodes) > 0 {
		responsible = issue.Assignees.Nodes[0].Name
	}

	priority := getTodoistPriority(labels)

	var parts []string
	parts = append(parts, fmt.Sprintf("🔗 GitLab: %s", issue.WebURL))
	parts = append(parts, fmt.Sprintf("IID: %s", issue.IID))
	parts = append(parts, fmt.Sprintf("Status: %s", issue.State))

	// DueDate einfach als String hinzufügen wenn vorhanden
	if issue.DueDate != nil && *issue.DueDate != "" {
		parts = append(parts, fmt.Sprintf("📅 Due: %s", *issue.DueDate))
	}

	description := strings.Join(parts, " | ")

	return TodoistRecord{
		Type:        "task",
		Content:     issue.Title,
		Description: description,
		Priority:    strconv.Itoa(priority),
		Indent:      "1",
		Author:      "",
		Responsible: responsible,
		Date:        "",
		DateLang:    "de",
		Timezone:    "Europe/Berlin",
		Labels:      strings.Join(labels, ","),
	}
}

func getTodoistPriority(labels []string) int {
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "urgent", "critical", "blocker", "p1":
			return 4
		case "high", "important", "p2":
			return 3
		case "low", "minor", "p4":
			return 1
		}
	}
	return 2
}

func (e *GitLabExporter) ExportToTodoistWithStructure(issues []Issue, filename string) error {
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
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Printf("fehler beim Schliessen der Datei.")
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"TYPE", "CONTENT", "DESCRIPTION", "PRIORITY", "INDENT",
		"AUTHOR", "RESPONSIBLE", "DATE", "DATE_LANG", "TIMEZONE", "LABELS",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for milestoneTitle, milestoneIssues := range milestoneGroups {
		projectRow := []string{
			"project", milestoneTitle, "", "1", "1", "", "", "", "de", "Europe/Berlin", "",
		}
		if err := writer.Write(projectRow); err != nil {
			return err
		}

		for _, issue := range milestoneIssues {
			record := e.convertIssueToTodoistRecord(issue)
			record.Indent = "2"

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
