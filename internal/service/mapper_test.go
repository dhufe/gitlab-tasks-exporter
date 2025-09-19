package service

import (
	"strings"
	"testing"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	domain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/models"
)

func testIssue(baseTitle string) domain.Issue {
	due := "2024-02-15T10:00:00Z" // ISO → ConvertToTodoistDate => 2024-02-15; display => 15.02.2024
	return domain.Issue{
		IID:         "123",
		Title:       baseTitle,
		State:       "opened",
		WebURL:      "https://gitlab.com/group/repo/-/issues/123",
		Description: strings.Repeat("A", 350), // will be truncated to 300 with ellipsis in description builder
		DueDate:     &due,
		Assignees:   domain.Assignees{Nodes: []domain.Assignee{{Name: "Alice"}, {Name: "Bob"}}},
		Labels:      domain.Labels{Nodes: []domain.Label{{Title: "High"}, {Title: "Bug Fix"}}},
	}
}

func TestGitLabToTodoistTask_BasicMapping(t *testing.T) {
	m := NewMapper(&config.Config{})
	issue := testIssue("Improve feature")

	req := m.GitLabToTodoistTask(issue, "p1", "s1")

	// Title mapping
	if req.Content != "#123 - Improve feature" {
		t.Fatalf("unexpected content: %q", req.Content)
	}
	// Project & section
	if req.ProjectID != "p1" || req.SectionID != "s1" {
		t.Fatalf("project/section mismatch: %q/%q", req.ProjectID, req.SectionID)
	}
	// Due date converted to YYYY-MM-DD
	if req.DueDate != "2024-02-15" {
		t.Fatalf("due date not converted: %q", req.DueDate)
	}
	// Labels normalized (lowercase, spaces→underscores) and includes state label "open"
	// From issue labels: "High" -> "high", "Bug Fix" -> "bug_fix"
	joined := strings.Join(req.Labels, ",")
	if !strings.Contains(joined, "high") || !strings.Contains(joined, "bug_fix") || !strings.Contains(joined, "open") {
		t.Fatalf("expected normalized labels incl. state, got: %v", req.Labels)
	}
	// Priority based on label ("High" → 3)
	if req.Priority != 3 {
		t.Fatalf("expected priority 3, got %d", req.Priority)
	}
	// Description should include GitLab link line
	if !strings.Contains(req.Description, "[GitLab Issue #123](https://gitlab.com/group/repo/-/issues/123)") {
		t.Fatalf("description missing GitLab link: %s", req.Description)
	}
}

func TestBuildTaskDescription_ContainsExpectedBlocks(t *testing.T) {
	m := NewMapper(&config.Config{})
	issue := testIssue("Title")

	desc := m.buildTaskDescription(issue)

	// Link
	if !strings.Contains(desc, "[GitLab Issue #123](https://gitlab.com/group/repo/-/issues/123)") {
		t.Errorf("missing link: %s", desc)
	}
	// Assignees
	if !strings.Contains(desc, "Assignees:") || !strings.Contains(desc, "Alice, Bob") {
		t.Errorf("missing assignees: %s", desc)
	}
	// Labels appear as backticked original titles (not normalized here)
	if !strings.Contains(desc, "`High`") || !strings.Contains(desc, "`Bug Fix`") {
		t.Errorf("labels not formatted: %s", desc)
	}
	// Due date displayed as dd.mm.yyyy after conversion
	if !strings.Contains(desc, "15.02.2024") {
		t.Errorf("due date not formatted for display: %s", desc)
	}
	// Description section exists and is truncated (should end with ...)
	if !strings.Contains(desc, "**Beschreibung:**") {
		t.Errorf("missing description header: %s", desc)
	}
	// Ensure we see ellipsis; exact length may vary, but should include "..."
	if !strings.Contains(desc, "...") {
		t.Errorf("expected truncated description with ellipsis: %s", desc)
	}
}

func TestExtractLabels_NormalizationAndState(t *testing.T) {
	m := NewMapper(&config.Config{})
	issue := domain.Issue{
		State:  "closed",
		Labels: domain.Labels{Nodes: []domain.Label{{Title: "Very Important"}, {Title: "needs QA"}}},
	}

	labels := m.extractLabels(issue)
	joined := strings.Join(labels, ",")

	// Normalized
	if !strings.Contains(joined, "very_important") || !strings.Contains(joined, "needs_qa") {
		t.Fatalf("labels not normalized: %v", labels)
	}
	// Contains state label "closed"
	if !strings.Contains(joined, "closed") {
		t.Fatalf("missing state label: %v", labels)
	}
}

func TestDeterminePriority_LabelMapping(t *testing.T) {
	m := NewMapper(&config.Config{})

	cases := []struct {
		labels []string
		want   int
	}{
		{[]string{"critical"}, 4},
		{[]string{"URGENT BUG"}, 4},
		{[]string{"high"}, 3},
		{[]string{"Important task"}, 3},
		{[]string{"medium"}, 2},
		{[]string{"low"}, 1},
		{[]string{"other"}, 1}, // default
	}

	for i, c := range cases {
		issue := domain.Issue{Labels: domain.Labels{Nodes: func() []domain.Label {
			out := make([]domain.Label, len(c.labels))
			for i := range c.labels {
				out[i] = domain.Label{Title: c.labels[i]}
			}
			return out
		}()}}
		if got := m.determinePriority(issue); got != c.want {
			t.Fatalf("case %d: expected %d, got %d", i, c.want, got)
		}
	}
}

func TestBuildProjectName_PreferencesAndMilestone(t *testing.T) {
	// If TodoistProject is set in config, it wins
	m1 := NewMapper(&config.Config{TodoistProject: "Custom Name"})
	if got := m1.BuildProjectName("group/repo", nil); got != "Custom Name" {
		t.Fatalf("expected custom name, got %q", got)
	}

	// Default: project path
	m2 := NewMapper(&config.Config{})
	if got := m2.BuildProjectName("group/repo", nil); got != "group/repo" {
		t.Fatalf("expected project path, got %q", got)
	}

	// With milestone (not empty, not *)
	ms := "v1.0"
	if got := m2.BuildProjectName("group/repo", &ms); got != "group/repo - v1.0" {
		t.Fatalf("expected path with milestone, got %q", got)
	}

	// With wildcard milestone "*" → ignored
	star := "*"
	if got := m2.BuildProjectName("group/repo", &star); got != "group/repo" {
		t.Fatalf("expected path without wildcard milestone, got %q", got)
	}
}

func TestDetermineSectionID_PicksClosedElseOpenFallback(t *testing.T) {
	m := NewMapper(&config.Config{})
	sections := map[string]string{"open": "sec-open", "closed": "sec-closed"}

	// Closed issue → closed section
	closedIssue := domain.Issue{State: "closed"}
	if got := m.DetermineSectionID(closedIssue, sections); got != "sec-closed" {
		t.Fatalf("expected sec-closed, got %q", got)
	}

	// Opened issue → open section
	openedIssue := domain.Issue{State: "opened"}
	if got := m.DetermineSectionID(openedIssue, sections); got != "sec-open" {
		t.Fatalf("expected sec-open, got %q", got)
	}

	// Missing sections → empty string
	empty := map[string]string{}
	if got := m.DetermineSectionID(openedIssue, empty); got != "" {
		t.Fatalf("expected empty section id, got %q", got)
	}
}
