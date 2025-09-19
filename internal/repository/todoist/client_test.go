package todoist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	domain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/models"
)

func newTodoistRepoWithServer(t *testing.T, handler http.HandlerFunc) (*Repository, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	cfg := &config.Config{TodoistToken: "todo-token"}

	repo := NewRepository(cfg)
	// Redirect baseURL to our test server (field is package-private, and we’re in package todoist)
	repo.baseURL = srv.URL

	return repo, srv
}

func TestTodoist_GetProjects_And_FindProjectByName(t *testing.T) {
	repo, srv := newTodoistRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" && r.Method == http.MethodGet {
			if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer todo-token") {
				t.Fatalf("missing/invalid Authorization header: %q", got)
			}
			_ = json.NewEncoder(w).Encode([]domain.Project{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	ps, err := repo.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects() error = %v", err)
	}
	if len(ps) != 2 || ps[1].Name != "B" {
		t.Fatalf("unexpected projects: %+v", ps)
	}

	// Find existing
	p, err := repo.FindProjectByName("B")
	if err != nil || p == nil || p.ID != "p2" {
		t.Fatalf("FindProjectByName() got %v, err=%v", p, err)
	}
	// Not found
	p, err = repo.FindProjectByName("Z")
	if err != nil || p != nil {
		t.Fatalf("expected nil project, got %v, err=%v", p, err)
	}
}

func TestTodoist_CreateProject_ErrorStatus(t *testing.T) {
	repo, srv := newTodoistRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad"}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	_, err := repo.CreateProject("X")
	if err == nil || !strings.Contains(err.Error(), "create project failed 400") {
		t.Fatalf("expected create project error, got %v", err)
	}
}

func TestTodoist_Sections_CRUD_and_Find(t *testing.T) {
	// Minimal in-memory state
	sections := []domain.Section{{ID: "s1", ProjectID: "p1", Name: "Offen", Order: 1}}

	repo, srv := newTodoistRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/sections":
			if r.URL.Query().Get("project_id") != "p1" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(sections)
		case r.Method == http.MethodPost && r.URL.Path == "/sections":
			var in domain.Section
			_ = json.NewDecoder(r.Body).Decode(&in)
			// Echo with assigned ID
			in.ID = "s2"
			_ = json.NewEncoder(w).Encode(in)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	defer srv.Close()

	got, err := repo.GetProjectSections("p1")
	if err != nil || len(got) != 1 || got[0].Name != "Offen" {
		t.Fatalf("GetProjectSections() got %v err=%v", got, err)
	}

	created, err := repo.CreateSection("p1", "Geschlossen", 2)
	if err != nil || created == nil || created.ID != "s2" || created.Name != "Geschlossen" {
		t.Fatalf("CreateSection() got %v err=%v", created, err)
	}

	// FindSectionByName uses GetProjectSections under the hood
	sec, err := repo.FindSectionByName("p1", "Offen")
	if err != nil || sec == nil || sec.ID != "s1" {
		t.Fatalf("FindSectionByName() got %v err=%v", sec, err)
	}
}

func TestTodoist_Tasks_CRUD_and_Find(t *testing.T) {
	tasks := []domain.Task{{ID: "t1", ProjectID: "p1", Content: "#1 - A", Description: "d", SectionID: "s1"}}

	repo, srv := newTodoistRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/tasks":
			if r.URL.Query().Get("project_id") != "p1" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(tasks)
		case r.Method == http.MethodPost && r.URL.Path == "/tasks":
			var in domain.CreateTaskRequest
			_ = json.NewDecoder(r.Body).Decode(&in)
			_ = json.NewEncoder(w).Encode(domain.Task{ID: "t2", Content: in.Content, Description: in.Description, ProjectID: in.ProjectID, SectionID: in.SectionID})
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/tasks/"):
			var updates map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&updates)
			id := strings.TrimPrefix(r.URL.Path, "/tasks/")
			// Echo back a merged object
			_ = json.NewEncoder(w).Encode(domain.Task{ID: id, Content: updates["content"].(string)})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	defer srv.Close()

	// List existing
	got, err := repo.GetProjectTasks("p1")
	if err != nil || len(got) != 1 || got[0].ID != "t1" {
		t.Fatalf("GetProjectTasks() got %v err=%v", got, err)
	}

	// Create
	created, err := repo.CreateTask(domain.CreateTaskRequest{ProjectID: "p1", SectionID: "s1", Content: "#2 - B", Description: "x"})
	if err != nil || created == nil || created.ID != "t2" {
		t.Fatalf("CreateTask() got %v err=%v", created, err)
	}

	// Update
	updated, err := repo.UpdateTask("t1", map[string]interface{}{"content": "#1 - A (updated)"})
	if err != nil || updated == nil || updated.ID != "t1" || !strings.Contains(updated.Content, "updated") {
		t.Fatalf("UpdateTask() got %v err=%v", updated, err)
	}

	// Find by title uses GetProjectTasks
	task, err := repo.FindTaskByTitle("p1", "#1 - A")
	if err != nil || task == nil || task.ID != "t1" {
		t.Fatalf("FindTaskByTitle() got %v err=%v", task, err)
	}
}

func TestTodoist_ValidateConnection_ErrorWrapped(t *testing.T) {
	repo, srv := newTodoistRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"detail":"invalid token"}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	err := repo.ValidateConnection()
	if err == nil || !strings.Contains(err.Error(), "todoist connection failed:") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
