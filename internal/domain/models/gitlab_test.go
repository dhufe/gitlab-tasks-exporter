package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGitLabIssue_JSON_RoundTrip(t *testing.T) {
	// Simulate a REST-style JSON payload coming from GitLab
	jsonIn := `{
        "iid": "42",
        "title": "Fix bug",
        "description": "Detailed description",
        "state": "opened",
        "web_url": "https://gitlab.com/g/r/p/-/issues/42",
        "due_date": "2025-09-01",
        "created_at": "2025-09-01T12:00:00Z",
        "updated_at": "2025-09-02T13:30:00Z",
        "labels": { "nodes": [ {"title": "Bug"}, {"title": "High"} ] },
        "assignees": { "nodes": [ {"name": "Alice"}, {"name": "Bob"} ] }
    }`

	var iss Issue
	if err := json.Unmarshal([]byte(jsonIn), &iss); err != nil {
		t.Fatalf("unmarshal Issue: %v", err)
	}

	if iss.IID != "42" || iss.Title != "Fix bug" || iss.State != "opened" {
		t.Fatalf("basic fields mismatch: %+v", iss)
	}
	if iss.WebURL != "https://gitlab.com/g/r/p/-/issues/42" {
		t.Fatalf("web url mismatch: %q", iss.WebURL)
	}
	if iss.DueDate == nil || *iss.DueDate != "2025-09-01" {
		t.Fatalf("due date mismatch: %#v", iss.DueDate)
	}

	// time.Time tags use created_at/updated_at; ensure parsed (UTC) without checking exact zone
	if iss.CreatedAt.IsZero() || iss.UpdatedAt.IsZero() {
		t.Fatalf("timestamps not parsed: created=%v updated=%v", iss.CreatedAt, iss.UpdatedAt)
	}

	if len(iss.Labels.Nodes) != 2 || iss.Labels.Nodes[0].Title != "Bug" {
		t.Fatalf("labels mismatch: %+v", iss.Labels)
	}
	if len(iss.Assignees.Nodes) != 2 || iss.Assignees.Nodes[1].Name != "Bob" {
		t.Fatalf("assignees mismatch: %+v", iss.Assignees)
	}

	// Marshal back to JSON and spot-check key names
	b, err := json.Marshal(iss)
	if err != nil {
		t.Fatalf("marshal Issue: %v", err)
	}
	s := string(b)
	// Ensure snake_case keys according to struct tags exist
	for _, key := range []string{"\"iid\"", "\"title\"", "\"description\"", "\"state\"", "\"web_url\"", "\"due_date\"", "\"created_at\"", "\"updated_at\""} {
		if !strings.Contains(s, key) {
			t.Fatalf("expected marshaled json to contain key %s: %s", key, s)
		}
	}
}

func TestGraphQLResponse_Unmarshal(t *testing.T) {
	// Minimal GraphQL envelope that should map to GraphQLResponse
	jsonIn := `{
        "data": {
            "project": {
                "issues": {
                    "nodes": [
                        {"iid":"1","title":"A","description":"","state":"opened","web_url":"u1","labels":{"nodes":[]},"assignees":{"nodes":[]},"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-02T00:00:00Z"},
                        {"iid":"2","title":"B","description":"","state":"closed","web_url":"u2","labels":{"nodes":[]},"assignees":{"nodes":[]},"created_at":"2025-02-01T00:00:00Z","updated_at":"2025-02-02T00:00:00Z"}
                    ]
                }
            }
        },
        "errors": []
    }`

	var resp GraphQLResponse
	if err := json.Unmarshal([]byte(jsonIn), &resp); err != nil {
		t.Fatalf("unmarshal GraphQLResponse: %v", err)
	}

	nodes := resp.Data.Project.Issues.Nodes
	if len(nodes) != 2 || nodes[0].IID != "1" || nodes[1].State != "closed" {
		t.Fatalf("nodes mismatch: %+v", nodes)
	}

	// Sanity: created/updated parsed into time.Time
	if nodes[0].CreatedAt.Equal(time.Time{}) || nodes[1].UpdatedAt.Equal(time.Time{}) {
		t.Fatalf("timestamps not parsed: %+v", nodes)
	}
}
