package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTask_JSON_FieldNamesAndOmitEmpty(t *testing.T) {
	// A task with only required and some fields set; optional fields with omitempty should be omitted
	task := Task{
		ID:        "t1",
		Content:   "#1 - Title",
		ProjectID: "p1",
		// SectionID empty → should be omitted due to omitempty
		Completed: false,      // emits as is_completed: false
		Labels:    []string{}, // no omitempty → will serialize as []
		Priority:  1,
		DueDate:   "", // omitempty
		URL:       "", // omitempty
	}

	b, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal task: %v", err)
	}
	s := string(b)

	// Field name mappings
	checks := []string{`"id"`, `"content"`, `"project_id"`, `"is_completed"`, `"labels"`, `"priority"`}
	for _, key := range checks {
		if !strings.Contains(s, key) {
			t.Fatalf("expected key %s in json: %s", key, s)
		}
	}

	// Omitted fields
	for _, key := range []string{`"section_id"`, `"due_date"`, `"url"`} {
		if strings.Contains(s, key) {
			t.Fatalf("did not expect %s in json: %s", key, s)
		}
	}

	// Labels should appear as an empty array, not be omitted
	if !strings.Contains(s, `"labels":[]`) {
		t.Fatalf("expected empty labels array in json: %s", s)
	}

	// Round-trip
	var back Task
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal task: %v", err)
	}
	if back.ID != "t1" || back.ProjectID != "p1" || back.Content != "#1 - Title" || back.Priority != 1 {
		t.Fatalf("round-trip mismatch: %+v", back)
	}
}

func TestCreateTaskRequest_OmitEmptyAndIncludeWhenSet(t *testing.T) {
	// Only required fields set
	req := CreateTaskRequest{
		Content:   "#2 - Title",
		ProjectID: "p2",
		// others default/empty and should be omitted
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal req: %v", err)
	}
	s := string(b)

	if !strings.Contains(s, `"content"`) || !strings.Contains(s, `"project_id"`) {
		t.Fatalf("expected content and project_id in json: %s", s)
	}
	for _, key := range []string{`"description"`, `"section_id"`, `"labels"`, `"priority"`, `"due_date"`, `"due_string"`} {
		if strings.Contains(s, key) {
			t.Fatalf("did not expect %s in json: %s", key, s)
		}
	}

	// Now set optional fields; they must appear
	req.Description = "desc"
	req.SectionID = "s1"
	req.Labels = []string{"bug", "high"}
	req.Priority = 3
	req.DueDate = "2025-09-20"
	req.DueString = "tomorrow"

	b2, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal req with fields: %v", err)
	}
	s2 := string(b2)

	for _, key := range []string{`"description"`, `"section_id"`, `"labels"`, `"priority"`, `"due_date"`, `"due_string"`} {
		if !strings.Contains(s2, key) {
			t.Fatalf("expected %s in json: %s", key, s2)
		}
	}
}

func TestProjectAndSection_JSON(t *testing.T) {
	p := Project{ID: "p1", Name: "Project", Color: "blue"}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal project: %v", err)
	}
	if !strings.Contains(string(b), `"name"`) || !strings.Contains(string(b), `"color"`) {
		t.Fatalf("project keys missing: %s", string(b))
	}

	s := Section{ID: "s1", ProjectID: "p1", Name: "Offen", Order: 1}
	b2, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal section: %v", err)
	}
	for _, key := range []string{`"project_id"`, `"name"`, `"order"`} {
		if !strings.Contains(string(b2), key) {
			t.Fatalf("section key missing: %s in %s", key, string(b2))
		}
	}

	// Round-trip sanity
	var pBack Project
	if err := json.Unmarshal(b, &pBack); err != nil {
		t.Fatalf("unmarshal project: %v", err)
	}
	if pBack.ID != "p1" || pBack.Name != "Project" {
		t.Fatalf("project round-trip mismatch: %+v", pBack)
	}

	var sBack Section
	if err := json.Unmarshal(b2, &sBack); err != nil {
		t.Fatalf("unmarshal section: %v", err)
	}
	if sBack.ProjectID != "p1" || sBack.Order != 1 {
		t.Fatalf("section round-trip mismatch: %+v", sBack)
	}
}
