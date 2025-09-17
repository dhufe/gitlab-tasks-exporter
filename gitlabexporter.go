package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

func (e *GitLabExporter) ExportMarkdown(issues []Issue, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("datei konnte nicht erstellt werden: %w", err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Print("fehler beim Schliessen der Datei.")
		}
	}()

	// Header
	timestamp := time.Now().Format("2006-01-02 15:04")
	_, err = fmt.Fprintf(file, "# Meeting Notes - %s\n", e.config.ProjectPath)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "**Exportiert am:** %s  \n", timestamp)
	if err != nil {
		return err
	}

	if e.config.MilestoneTitle != nil && *e.config.MilestoneTitle != "*" && *e.config.MilestoneTitle != "" {
		_, err = fmt.Fprintf(file, "**Milestone:** %s  \n", *e.config.MilestoneTitle)
		if err != nil {
			return err
		}
	}

	if e.config.AssignedUser != nil {
		_, err = fmt.Fprintf(file, "**Assigned to:** %s  \n", *e.config.AssignedUser)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "**Issues gesamt:** %d\n\n", len(issues))
	if err != nil {
		return err
	}

	// Inhaltsverzeichnis
	_, err = fmt.Fprintf(file, "## 📋 Agenda\n\n")
	if err != nil {
		return err
	}

	for _, issue := range issues {
		priority := e.getPriorityEmoji(issue)
		state := e.getStateEmoji(issue.State)

		_, err = fmt.Fprintf(file, "- [%s #%s: %s](#issue-%s) %s %s\n",
			state, issue.IID, issue.Title, issue.IID, priority, e.getLabelsString(issue))
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "\n---\n\n")
	if err != nil {
		return err
	}

	// Issues detailliert
	for i, issue := range issues {
		err = e.writeIssueSection(file, issue, i+1, len(issues))
		if err != nil {
			return err
		}
	}

	// Footer mit Aktionsbereich
	err = e.writeFooterSection(file)
	if err != nil {
		return err
	}

	return nil
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

	// DueDate nur hinzufügen wenn vorhanden
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

func (e *GitLabExporter) writeIssueSection(file *os.File, issue Issue, index, total int) error {
	priority := e.getPriorityEmoji(issue)
	state := e.getStateEmoji(issue.State)

	// Issue Header
	_, err := fmt.Fprintf(file, "## Issue #%s: %s {#issue-%s}\n\n",
		issue.IID, issue.Title, issue.IID)
	if err != nil {
		return err
	}

	// Status und Metadaten
	_, err = fmt.Fprintf(file, "**Status:** %s %s | **Priorität:** %s | **Progress:** %d/%d\n\n",
		state, e.getStateDisplayName(issue.State), priority, index, total)
	if err != nil {
		return err
	}

	// Issue-Infos in Box
	_, err = fmt.Fprintf(file, "> 📎 **GitLab:** [%s](%s)  \n", issue.WebURL, issue.WebURL)
	if err != nil {
		return err
	}

	// Assignee nur wenn vorhanden
	if len(issue.Assignees.Nodes) > 0 && issue.Assignees.Nodes[0].Name != "" {
		_, err = fmt.Fprintf(file, "> 👤 **Assignee:** %s  \n", issue.Assignees.Nodes[0].Name)
		if err != nil {
			return err
		}
	}

	// DueDate nur wenn vorhanden
	if issue.DueDate != nil && *issue.DueDate != "" {
		_, err = fmt.Fprintf(file, "> 📅 **Due Date:** %s  \n", *issue.DueDate)
		if err != nil {
			return err
		}
	}

	labels := e.getLabelsString(issue)
	if labels != "" {
		_, err = fmt.Fprintf(file, "> 🏷️ **Labels:** %s  \n", labels)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "\n")
	if err != nil {
		return err
	}

	// Rest der Funktion bleibt gleich...
	// Beschreibung falls vorhanden
	if issue.Description != "" {
		_, err = fmt.Fprintf(file, "### 📝 Beschreibung\n\n")
		if err != nil {
			return err
		}

		// Beschreibung einrücken
		description := strings.ReplaceAll(issue.Description, "\n", "\n> ")
		_, err = fmt.Fprintf(file, "> %s\n\n", description)
		if err != nil {
			return err
		}
	}

	// Meeting-Notizen Bereich
	_, err = fmt.Fprintf(file, "### 💬 Meeting Notes\n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "**Diskussion:**\n")
	if err != nil {
		return err
	}

	for i := 0; i < 5; i++ {
		_, err = fmt.Fprintf(file, "- \n")
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "\n**Entscheidungen:**\n")
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		_, err = fmt.Fprintf(file, "- [ ] \n")
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "\n**Nächste Schritte:**\n")
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		_, err = fmt.Fprintf(file, "- [ ] \n")
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "\n**Release Notes:**\n```\n\n\n```\n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "---\n\n")
	return err
}

func (e *GitLabExporter) writeFooterSection(file *os.File) error {
	_, err := fmt.Fprintf(file, "## 📋 Meeting Summary\n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "### Gesammelte Action Items\n")
	if err != nil {
		return err
	}

	for i := 0; i < 8; i++ {
		_, err = fmt.Fprintf(file, "- [ ] \n")
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(file, "\n### Follow-up Meeting\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "**Datum:** _________________  \n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "**Teilnehmer:** _________________  \n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "**Themen:** _________________  \n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "### Changelog/Release Notes\n```markdown\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "## Version X.X.X\n\n### Features\n- \n\n### Fixes\n- \n\n### Changes\n- \n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "```\n\n")
	if err != nil {
		return err
	}

	// Timestamp Footer
	_, err = fmt.Fprintf(file, "---\n*Generiert am %s*\n",
		time.Now().Format("2006-01-02 15:04:05"))
	return err
}

// Helper functions
func (e *GitLabExporter) getPriorityEmoji(issue Issue) string {
	// Basierend auf Labels oder anderen Kriterien
	labels := strings.ToLower(e.getLabelsString(issue))
	if strings.Contains(labels, "high") || strings.Contains(labels, "urgent") {
		return "🔴"
	}
	if strings.Contains(labels, "medium") {
		return "🟡"
	}
	if strings.Contains(labels, "low") {
		return "🟢"
	}
	return "⚪"
}

func (e *GitLabExporter) getStateEmoji(state string) string {
	switch strings.ToLower(state) {
	case "opened":
		return "🔵"
	case "closed":
		return "✅"
	case "merged":
		return "🟣"
	default:
		return "⚪"
	}
}

func (e *GitLabExporter) getLabelsString(issue Issue) string {
	if len(issue.Labels.Nodes) == 0 {
		return ""
	}

	var labels []string
	for _, label := range issue.Labels.Nodes {
		labels = append(labels, "`"+label.Title+"`")
	}
	return strings.Join(labels, " ")
}

func (e *GitLabExporter) getStateDisplayName(state string) string {
	switch strings.ToLower(state) {
	case "opened":
		return "Offen"
	case "closed":
		return "Geschlossen"
	case "merged":
		return "Gemergt"
	default:
		return strings.ToUpper(state[:1]) + strings.ToLower(state[1:])
	}
}
