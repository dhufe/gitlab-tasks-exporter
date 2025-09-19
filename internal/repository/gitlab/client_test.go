package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	domain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/models"
)

func newGitLabRepoWithServer(t *testing.T, handler http.HandlerFunc) (*Repository, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)

	// Build config pointing to our fake server
	cfg := &config.Config{
		GitLabToken: "test-token",
		GitLabURL:   srv.URL, // NewRepository derives baseURL from this
	}

	repo := NewRepository(cfg)
	return repo, srv
}

func TestGitLab_ValidateConnection_OK(t *testing.T) {
	repo, srv := newGitLabRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/user" {
			if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
				t.Fatalf("missing Authorization header, got %q", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "username": "tester"}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	if err := repo.ValidateConnection(); err != nil {
		t.Fatalf("ValidateConnection() error = %v", err)
	}
}

func TestGitLab_ValidateConnection_Unauthorized(t *testing.T) {
	repo, srv := newGitLabRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/user" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"401 Unauthorized"}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	err := repo.ValidateConnection()
	if err == nil || !strings.Contains(err.Error(), "invalid GitLab token") {
		t.Fatalf("expected unauthorized error, got: %v", err)
	}
}

func TestGitLab_GetProjectIssues_Success(t *testing.T) {
	repo, srv := newGitLabRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v4/projects/") && strings.HasSuffix(r.URL.Path, "/issues") {
			if r.URL.Query().Get("state") != "opened" || r.URL.Query().Get("per_page") != "100" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer test-token") {
				t.Fatalf("missing/invalid Authorization header: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Minimal payload compatible with REST tags in domain.Issue
			issues := []domain.Issue{{
				IID: "1", Title: "Issue A", State: "opened", WebURL: "https://example/1",
				CreatedAt: time.Unix(0, 0), UpdatedAt: time.Unix(0, 0), Labels: domain.Labels{}, Assignees: domain.Assignees{},
			}, {
				IID: "2", Title: "Issue B", State: "opened", WebURL: "https://example/2",
				CreatedAt: time.Unix(0, 0), UpdatedAt: time.Unix(0, 0), Labels: domain.Labels{}, Assignees: domain.Assignees{},
			}}
			_ = json.NewEncoder(w).Encode(issues)
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	got, err := repo.GetProjectIssues("group/project")
	if err != nil {
		t.Fatalf("GetProjectIssues() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(got))
	}
}

func TestGitLab_GetProjectIssues_ErrorStatus(t *testing.T) {
	repo, srv := newGitLabRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v4/projects/") && strings.HasSuffix(r.URL.Path, "/issues") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	})
	defer srv.Close()

	_, err := repo.GetProjectIssues("group/project")
	if err == nil || !strings.Contains(err.Error(), "GitLab API error: 500") {
		t.Fatalf("expected API error, got: %v", err)
	}
}

func TestGitLab_GetMilestoneIssues_Success(t *testing.T) {
	repo, srv := newGitLabRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/graphql" && r.Method == http.MethodPost {
			// Return a minimal GraphQL envelope
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(domain.GraphQLResponse{
				Data: struct {
					Project struct {
						Issues struct {
							Nodes []domain.Issue `json:"nodes"`
						} `json:"issues"`
					} `json:"project"`
				}{Project: struct {
					Issues struct {
						Nodes []domain.Issue `json:"nodes"`
					} `json:"issues"`
				}{
					Issues: struct {
						Nodes []domain.Issue `json:"nodes"`
					}(struct {
						Nodes []domain.Issue `json:"nodes"`
					}{Nodes: []domain.Issue{{IID: "7"}}}),
				}},
			})
			return
		}
		t.Fatalf("unexpected path/method: %s %s", r.Method, r.URL.Path)
	})
	defer srv.Close()

	// Use repo but ensure GraphQL hits the server; executeGraphQLQuery uses cfg.GetGitLabBaseURL()
	// which reads from cfg.GitLabURL we already set in newGitLabRepoWithServer.
	res, err := repo.GetMilestoneIssues("group/project", nil)
	if err != nil {
		t.Fatalf("GetMilestoneIssues() error = %v", err)
	}
	if len(res) != 1 || res[0].IID != "7" {
		t.Fatalf("unexpected GraphQL issues: %+v", res)
	}
}

func TestGitLab_GetMilestoneIssues_GraphQLErrors(t *testing.T) {
	repo, srv := newGitLabRepoWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/graphql" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(domain.GraphQLResponse{
				Errors: []struct {
					Message string `json:"message"`
				}{{Message: "bad query"}},
			})
			return
		}
		t.Fatalf("unexpected path/method: %s %s", r.Method, r.URL.Path)
	})
	defer srv.Close()

	_, err := repo.GetMilestoneIssues("group/project", nil)
	if err == nil || !strings.Contains(err.Error(), "GraphQL errors:") {
		t.Fatalf("expected graphQL errors, got %v", err)
	}
}
