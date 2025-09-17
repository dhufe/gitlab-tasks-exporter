package service

import (
	"fmt"
	"strings"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	gitlabDomain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/gitlab"
	todoistDomain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/todoist"
	"hufschlaeger.net/gitlab-tasks-exporter/pkg/utils"
)

type Mapper struct {
	config *config.Config
}

func NewMapper(cfg *config.Config) *Mapper {
	return &Mapper{config: cfg}
}

// GitLabToTodoistTask konvertiert GitLab Issue zu Todoist Task
func (m *Mapper) GitLabToTodoistTask(issue gitlabDomain.Issue, projectID string, sectionID string) todoistDomain.CreateTaskRequest {
	// Title mit Issue-Nummer
	title := fmt.Sprintf("#%s - %s", issue.IID, issue.Title)

	// Description mit Link
	description := m.buildTaskDescription(issue)

	// Labels extrahieren
	labels := m.extractLabels(issue)

	// Priority basierend auf Labels bestimmen
	priority := m.determinePriority(issue)

	// Due Date konvertieren
	dueDate := ""
	if issue.DueDate != nil && *issue.DueDate != "" {
		dueDate = utils.ConvertToTodoistDate(*issue.DueDate)
	}

	return todoistDomain.CreateTaskRequest{
		Content:     title,
		Description: description,
		ProjectID:   projectID,
		SectionID:   sectionID,
		Labels:      labels,
		Priority:    priority,
		DueDate:     dueDate,
	}
}

// buildTaskDescription erstellt eine strukturierte Task-Beschreibung
func (m *Mapper) buildTaskDescription(issue gitlabDomain.Issue) string {
	var parts []string

	// GitLab Link
	parts = append(parts, fmt.Sprintf("🔗 [GitLab Issue #%s](%s)", issue.IID, issue.WebURL))

	// Assignees
	if len(issue.Assignees.Nodes) > 0 {
		var assigneeNames []string
		for _, assignee := range issue.Assignees.Nodes {
			assigneeNames = append(assigneeNames, assignee.Name)
		}
		parts = append(parts, fmt.Sprintf("👤 **Assignees:** %s", strings.Join(assigneeNames, ", ")))
	}

	// Labels
	if len(issue.Labels.Nodes) > 0 {
		var labelNames []string
		for _, label := range issue.Labels.Nodes {
			labelNames = append(labelNames, "`"+label.Title+"`")
		}
		parts = append(parts, fmt.Sprintf("🏷️ **Labels:** %s", strings.Join(labelNames, " ")))
	}

	// Due Date
	if issue.DueDate != nil && *issue.DueDate != "" {
		formattedDate := utils.FormatDateForDisplay(utils.ConvertToTodoistDate(*issue.DueDate))
		parts = append(parts, fmt.Sprintf("📅 **Due Date:** %s", formattedDate))
	}

	// Original Description (gekürzt)
	if issue.Description != "" {
		truncatedDesc := utils.TruncateText(issue.Description, 300)
		parts = append(parts, "", "**Beschreibung:**", truncatedDesc)
	}

	return strings.Join(parts, "\n")
}

// extractLabels extrahiert Labels für Todoist
func (m *Mapper) extractLabels(issue gitlabDomain.Issue) []string {
	var labels []string

	for _, label := range issue.Labels.Nodes {
		// Label normalisieren (keine Leerzeichen, lowercase)
		normalizedLabel := strings.ToLower(strings.ReplaceAll(label.Title, " ", "_"))
		labels = append(labels, normalizedLabel)
	}

	// Issue State als Label hinzufügen
	if issue.State == "opened" {
		labels = append(labels, "open")
	} else if issue.State == "closed" {
		labels = append(labels, "closed")
	}

	return labels
}

// determinePriority bestimmt Todoist Priority basierend auf GitLab Labels
func (m *Mapper) determinePriority(issue gitlabDomain.Issue) int {
	for _, label := range issue.Labels.Nodes {
		labelLower := strings.ToLower(label.Title)

		// Priority Labels checken
		switch {
		case strings.Contains(labelLower, "critical") || strings.Contains(labelLower, "urgent"):
			return 4 // Höchste Priority
		case strings.Contains(labelLower, "high") || strings.Contains(labelLower, "important"):
			return 3
		case strings.Contains(labelLower, "medium"):
			return 2
		case strings.Contains(labelLower, "low"):
			return 1
		}
	}

	// Default Priority
	return 1
}

// BuildProjectName erstellt einen Todoist-Projektnamen
func (m *Mapper) BuildProjectName(projectPath string, milestoneTitle *string) string {
	if m.config.TodoistProject != "" {
		return m.config.TodoistProject
	}

	// Standard: GitLab Repository Name + Milestone
	projectName := projectPath
	if milestoneTitle != nil && *milestoneTitle != "" && *milestoneTitle != "*" {
		projectName = fmt.Sprintf("%s - %s", projectPath, *milestoneTitle)
	}

	return projectName
}

// DetermineSectionID bestimmt die richtige Section basierend auf Issue State
func (m *Mapper) DetermineSectionID(issue gitlabDomain.Issue, sections map[string]string) string {
	if issue.State == "closed" {
		if sectionID, exists := sections["closed"]; exists {
			return sectionID
		}
	}

	// Default: Open Section
	if sectionID, exists := sections["open"]; exists {
		return sectionID
	}

	return "" // Keine Section
}
