package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	gitlabDomain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/gitlab"
	todoistDomain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/todoist"
	gitlabRepo "hufschlaeger.net/gitlab-tasks-exporter/internal/repository/gitlab"
	todoistRepo "hufschlaeger.net/gitlab-tasks-exporter/internal/repository/todoist"
	"hufschlaeger.net/gitlab-tasks-exporter/pkg/utils"
)

type Exporter struct {
	config      *config.Config
	gitlabRepo  *gitlabRepo.Repository
	todoistRepo *todoistRepo.Repository
	mapper      *Mapper
}

func NewExporter(cfg *config.Config) *Exporter {
	return &Exporter{
		config:      cfg,
		gitlabRepo:  gitlabRepo.NewRepository(cfg),
		todoistRepo: todoistRepo.NewRepository(cfg),
		mapper:      NewMapper(cfg),
	}
}

// Export startet den Hauptexport-Prozess
func (e *Exporter) Export() error {
	// 1. Konfiguration validieren
	if err := e.config.Validate(); err != nil {
		return fmt.Errorf("konfiguration ungültig: %w", err)
	}

	fmt.Printf("🔍 Lade Issues aus GitLab: %s\n", e.config.ProjectPath)

	// 2. Issues von GitLab laden
	issues, err := e.loadGitLabIssues()
	if err != nil {
		return fmt.Errorf("fehler beim Laden der GitLab Issues: %w", err)
	}

	fmt.Printf("📊 Gefunden: %d Issues\n", len(issues))

	if len(issues) == 0 {
		fmt.Println("ℹ️  Keine Issues gefunden")
		return nil
	}

	// 3. Export-Modus bestimmen
	if e.config.TodoistAPI {
		return e.exportToTodoist(issues)
	}

	return e.exportToFile(issues)
}

func (e *Exporter) loadGitLabIssues() ([]gitlabDomain.Issue, error) {
	// Verbindung testen
	if err := e.gitlabRepo.ValidateConnection(); err != nil {
		return nil, fmt.Errorf("GitLab-Verbindung fehlgeschlagen: %w", err)
	}

	// Issues laden (je nach Milestone-Filter)
	if e.config.MilestoneTitle != nil && *e.config.MilestoneTitle != "*" {
		fmt.Printf("🎯 Filter nach Milestone: %s\n", *e.config.MilestoneTitle)
		return e.gitlabRepo.GetMilestoneIssues(e.config.ProjectPath, e.config.MilestoneTitle)
	}

	fmt.Println("📋 Lade alle Issues...")
	return e.gitlabRepo.GetMilestoneIssues(e.config.ProjectPath, nil)
}

// exportToTodoist exportiert Issues zu Todoist
func (e *Exporter) exportToTodoist(issues []gitlabDomain.Issue) error {
	fmt.Println("🚀 Exportiere zu Todoist...")

	// 1. Todoist-Verbindung testen
	if err := e.todoistRepo.ValidateConnection(); err != nil {
		return fmt.Errorf("Todoist-Verbindung fehlgeschlagen: %w", err)
	}

	// 2. Projekt einrichten
	projectID, err := e.setupTodoistProject()
	if err != nil {
		return fmt.Errorf("projekt-Setup fehlgeschlagen: %w", err)
	}

	// 3. Sections einrichten
	sections, err := e.setupTodoistSections(projectID)
	if err != nil {
		return fmt.Errorf("section-Setup fehlgeschlagen: %w", err)
	}

	// 4. Bestehende Tasks laden
	existingTasks, err := e.loadExistingTasks(projectID)
	if err != nil {
		return fmt.Errorf("fehler beim Laden bestehender Tasks: %w", err)
	}

	fmt.Printf("🔍 Gefunden: %d bestehende Tasks\n", len(existingTasks))

	// 5. Issues zu Tasks konvertieren und erstellen/aktualisieren
	return e.syncIssuesToTasks(issues, projectID, sections, existingTasks)
}

// setupTodoistProject richtet das Todoist-Projekt ein
func (e *Exporter) setupTodoistProject() (string, error) {
	projectName := e.mapper.BuildProjectName(e.config.ProjectPath, e.config.MilestoneTitle)

	// Projekt suchen
	existingProject, err := e.todoistRepo.FindProjectByName(projectName)
	if err != nil {
		return "", err
	}

	if existingProject != nil {
		fmt.Printf("📋 Verwende bestehendes Projekt: %s (ID: %s)\n",
			existingProject.Name, existingProject.ID)
		return existingProject.ID, nil
	}

	// Neues Projekt erstellen
	fmt.Printf("📋 Erstelle neues Projekt: %s\n", projectName)
	newProject, err := e.todoistRepo.CreateProject(projectName)
	if err != nil {
		return "", err
	}

	return newProject.ID, nil
}

// setupTodoistSections richtet die Sections ein
func (e *Exporter) setupTodoistSections(projectID string) (map[string]string, error) {
	sections := make(map[string]string)

	requiredSections := []struct {
		name  string
		key   string
		order int
	}{
		{"Offen", "open", 1},
		{"Geschlossen", "closed", 2},
	}

	for _, reqSection := range requiredSections {
		// Section suchen
		existingSection, err := e.todoistRepo.FindSectionByName(projectID, reqSection.name)
		if err != nil {
			return nil, err
		}

		if existingSection != nil {
			sections[reqSection.key] = existingSection.ID
			continue
		}

		// Section erstellen
		newSection, err := e.todoistRepo.CreateSection(projectID, reqSection.name, reqSection.order)
		if err != nil {
			return nil, fmt.Errorf("fehler beim Erstellen der Section '%s': %w", reqSection.name, err)
		}

		sections[reqSection.key] = newSection.ID
	}

	fmt.Printf("📂 Sections eingerichtet: Offen (%s), Geschlossen (%s)\n",
		sections["open"], sections["closed"])

	return sections, nil
}

// loadExistingTasks lädt alle bestehenden Tasks des Projekts
func (e *Exporter) loadExistingTasks(projectID string) (map[string]*todoistDomain.Task, error) {
	tasks, err := e.todoistRepo.GetProjectTasks(projectID)
	if err != nil {
		return nil, err
	}

	// Tasks in Map für schnellen Lookup (Key: Issue-IID aus Content)
	taskMap := make(map[string]*todoistDomain.Task)

	for i := range tasks {
		task := &tasks[i]

		// Issue-IID aus Task-Content extrahieren (Format: "#123 - Title")
		if issueIID := extractIssueIIDFromContent(task.Content); issueIID != "" {
			taskMap[issueIID] = task
		}
	}

	return taskMap, nil
}

// syncIssuesToTasks synchronisiert GitLab Issues mit Todoist Tasks
func (e *Exporter) syncIssuesToTasks(issues []gitlabDomain.Issue, projectID string, sections map[string]string, existingTasks map[string]*todoistDomain.Task) error {
	stats := struct {
		created, updated, skipped int
	}{}

	for _, issue := range issues {
		if err := e.syncSingleIssue(issue, projectID, sections, existingTasks, &stats); err != nil {
			fmt.Printf("⚠️  Fehler bei Issue #%s: %v\n", issue.IID, err)
			continue
		}
	}

	// Statistiken ausgeben
	fmt.Printf("\n🎉 Synchronisation abgeschlossen:\n")
	fmt.Printf("  ✅  Erstellt: %d\n", stats.created)
	fmt.Printf("  🔄  Aktualisiert: %d\n", stats.updated)
	fmt.Printf("  ⏭️  Übersprungen: %d\n", stats.skipped)

	return nil
}

// syncSingleIssue synchronisiert ein einzelnes Issue
func (e *Exporter) syncSingleIssue(issue gitlabDomain.Issue, projectID string, sections map[string]string, existingTasks map[string]*todoistDomain.Task, stats *struct{ created, updated, skipped int }) error {
	existingTask := existingTasks[issue.IID]

	// Section für Issue bestimmen
	sectionID := e.mapper.DetermineSectionID(issue, sections)

	if existingTask == nil {
		// Neuen Task erstellen
		return e.createNewTask(issue, projectID, sectionID, stats)
	}

	// Bestehenden Task aktualisieren (falls nötig)
	return e.updateExistingTask(issue, existingTask, sectionID, stats)
}

// createNewTask erstellt einen neuen Todoist Task
func (e *Exporter) createNewTask(issue gitlabDomain.Issue, projectID string, sectionID string, stats *struct{ created, updated, skipped int }) error {
	taskRequest := e.mapper.GitLabToTodoistTask(issue, projectID, sectionID)

	createdTask, err := e.todoistRepo.CreateTask(taskRequest)
	if err != nil {
		return fmt.Errorf("task-Erstellung fehlgeschlagen: %w", err)
	}

	fmt.Printf("✅ Task erstellt: #%s - %s (ID: %s)\n",
		issue.IID, issue.Title, createdTask.ID)

	stats.created++
	return nil
}

// updateExistingTask aktualisiert einen bestehenden Task falls nötig
func (e *Exporter) updateExistingTask(issue gitlabDomain.Issue, existingTask *todoistDomain.Task, sectionID string, stats *struct{ created, updated, skipped int }) error {
	updates := make(map[string]interface{})
	needsUpdate := false

	// Title prüfen
	expectedTitle := fmt.Sprintf("#%s - %s", issue.IID, issue.Title)
	if existingTask.Content != expectedTitle {
		updates["content"] = expectedTitle
		needsUpdate = true
	}

	// Section prüfen (State-Änderung)
	if existingTask.SectionID != sectionID && sectionID != "" {
		updates["section_id"] = sectionID
		needsUpdate = true
	}

	// Description prüfen
	expectedDescription := e.mapper.buildTaskDescription(issue)
	if existingTask.Description != expectedDescription {
		updates["description"] = expectedDescription
		needsUpdate = true
	}

	if !needsUpdate {
		stats.skipped++
		return nil
	}

	// Task aktualisieren
	_, err := e.todoistRepo.UpdateTask(existingTask.ID, updates)
	if err != nil {
		return fmt.Errorf("task-Update fehlgeschlagen: %w", err)
	}

	fmt.Printf("🔄 Task aktualisiert: #%s - %s\n", issue.IID, issue.Title)
	stats.updated++

	return nil
}

// exportToFile exportiert Issues in eine Markdown-Datei
func (e *Exporter) exportToFile(issues []gitlabDomain.Issue) error {
	fmt.Println("📄 Exportiere zu Markdown-Datei...")

	filename := e.generateFilename()
	content := e.generateMarkdownContent(issues)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("datei-Export fehlgeschlagen: %w", err)
	}

	fmt.Printf("✅ Datei erstellt: %s (%d Issues)\n", filename, len(issues))
	return nil
}

// generateFilename erstellt einen Dateinamen
func (e *Exporter) generateFilename() string {
	if e.config.OutputFile != "" {
		return e.config.OutputFile
	}

	// Standard-Format: project-milestone-YYYY-MM-DD.md
	timestamp := time.Now().Format("2006-01-02")
	projectName := strings.ReplaceAll(e.config.ProjectPath, "/", "-")

	if e.config.MilestoneTitle != nil && *e.config.MilestoneTitle != "" {
		milestone := strings.ReplaceAll(*e.config.MilestoneTitle, " ", "-")
		return fmt.Sprintf("%s-%s-%s.md", projectName, milestone, timestamp)
	}

	return fmt.Sprintf("%s-%s.md", projectName, timestamp)
}

// generateMarkdownContent generiert Markdown-Content
func (e *Exporter) generateMarkdownContent(issues []gitlabDomain.Issue) string {
	var content strings.Builder

	// Header
	content.WriteString(fmt.Sprintf("# GitLab Issues Export - %s\n\n", e.config.ProjectPath))
	content.WriteString(fmt.Sprintf("**Export-Zeit:** %s  \n", time.Now().Format("02.01.2006 15:04:05")))
	content.WriteString(fmt.Sprintf("**Anzahl Issues:** %d  \n\n", len(issues)))

	if e.config.MilestoneTitle != nil && *e.config.MilestoneTitle != "" {
		content.WriteString(fmt.Sprintf("**Milestone:** %s  \n\n", *e.config.MilestoneTitle))
	}

	// Offene Issues
	openIssues := filterIssuesByState(issues, "opened")
	if len(openIssues) > 0 {
		content.WriteString("## 🟢 Offene Issues\n\n")
		for _, issue := range openIssues {
			content.WriteString(e.formatIssueAsMarkdown(issue))
		}
	}

	// Geschlossene Issues
	closedIssues := filterIssuesByState(issues, "closed")
	if len(closedIssues) > 0 {
		content.WriteString("## ✅ Geschlossene Issues\n\n")
		for _, issue := range closedIssues {
			content.WriteString(e.formatIssueAsMarkdown(issue))
		}
	}

	return content.String()
}

// formatIssueAsMarkdown formatiert ein Issue als Markdown
func (e *Exporter) formatIssueAsMarkdown(issue gitlabDomain.Issue) string {
	var content strings.Builder

	// Title mit Link
	content.WriteString(fmt.Sprintf("### [#%s - %s](%s)\n\n",
		issue.IID,
		utils.EscapeMarkdown(issue.Title),
		issue.WebURL))

	// Metadata-Tabelle
	content.WriteString("| Feld | Wert |\n")
	content.WriteString("|------|------|\n")
	content.WriteString(fmt.Sprintf("| **Status** | %s |\n", issue.State))

	// Due Date
	if issue.DueDate != nil && *issue.DueDate != "" {
		content.WriteString(fmt.Sprintf("| **Fällig** | %s |\n",
			utils.FormatDateForDisplay(*issue.DueDate)))
	}

	// Assignees
	if len(issue.Assignees.Nodes) > 0 {
		var assigneeNames []string
		for _, assignee := range issue.Assignees.Nodes {
			assigneeNames = append(assigneeNames, assignee.Name)
		}
		content.WriteString(fmt.Sprintf("| **Zugewiesen** | %s |\n",
			strings.Join(assigneeNames, ", ")))
	}

	// Labels
	if len(issue.Labels.Nodes) > 0 {
		var labelNames []string
		for _, label := range issue.Labels.Nodes {
			labelNames = append(labelNames, "`"+label.Title+"`")
		}
		content.WriteString(fmt.Sprintf("| **Labels** | %s |\n",
			strings.Join(labelNames, " ")))
	}

	content.WriteString("\n")

	// Description
	if issue.Description != "" {
		content.WriteString("**Beschreibung:**\n\n")
		content.WriteString(issue.Description)
		content.WriteString("\n\n")
	}

	content.WriteString("---\n\n")
	return content.String()
}

// Helper Functions

func extractIssueIIDFromContent(content string) string {
	// Format: "#123 - Title" -> "123"
	if strings.HasPrefix(content, "#") {
		parts := strings.Split(content, " - ")
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "#")
		}
	}
	return ""
}

func filterIssuesByState(issues []gitlabDomain.Issue, state string) []gitlabDomain.Issue {
	var filtered []gitlabDomain.Issue
	for _, issue := range issues {
		if issue.State == state {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}
