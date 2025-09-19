package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	gitlabDomain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/models"
)

type Repository struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
}

func NewRepository(cfg *config.Config) *Repository {
	return &Repository{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    cfg.GetGitLabBaseURL() + "/api/v4",
	}
}

// GetMilestoneIssues holt alle Issues eines Milestones via GraphQL
func (r *Repository) GetMilestoneIssues(projectPath string, milestoneTitle *string) ([]gitlabDomain.Issue, error) {
	query := r.buildMilestoneQuery(projectPath, milestoneTitle)

	response, err := r.executeGraphQLQuery(query)
	if err != nil {
		return nil, fmt.Errorf("GraphQL query failed: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", response.Errors[0].Message)
	}

	return response.Data.Project.Issues.Nodes, nil
}

// GetProjectIssues holt alle Issues eines Projekts via REST API
func (r *Repository) GetProjectIssues(projectPath string) ([]gitlabDomain.Issue, error) {
	url := fmt.Sprintf("%s/projects/%s/issues", r.baseURL, projectPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.GitLabToken)
	req.Header.Set("Content-Type", "application/json")

	// Query parameters
	q := req.URL.Query()
	q.Add("state", "opened")
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Printf("fehler beim Abschliessen des Response bodies.")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	var issues []gitlabDomain.Issue
	err = json.NewDecoder(resp.Body).Decode(&issues)
	return issues, err
}

// ValidateConnection prüft ob die GitLab-Verbindung funktioniert
func (r *Repository) ValidateConnection() error {
	url := fmt.Sprintf("%s/user", r.baseURL)
	fmt.Printf("   Config GitLab URL: %s\n", r.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.GitLabToken)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Printf("fehler beim Abschliessen des Response bodies.")
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid GitLab token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	return nil
}

// Private helper methods

func (r *Repository) buildMilestoneQuery(projectPath string, milestoneTitle *string) string {
	milestoneFilter := ""
	if milestoneTitle != nil && *milestoneTitle != "" && *milestoneTitle != "*" {
		milestoneFilter = fmt.Sprintf(`, milestoneTitle: "%s"`, *milestoneTitle)
	}

	return fmt.Sprintf(`{
        project(fullPath: "%s") {
            issues(first: 100%s) {
                nodes {
                    iid
                    title
                    description
                    state
                    webUrl
                    dueDate
                    createdAt
                    updatedAt
                    labels {
                        nodes {
                            title
                        }
                    }
                    assignees {
                        nodes {
                            name
                        }
                    }
                }
            }
        }
    }`, projectPath, milestoneFilter)
}

func (r *Repository) executeGraphQLQuery(query string) (*gitlabDomain.GraphQLResponse, error) {
	url := r.config.GetGitLabBaseURL() + "/api/graphql"

	requestBody := map[string]string{"query": query}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.GitLabToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Printf("fehler beim Abschliessen des Response bodies.")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var response gitlabDomain.GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	return &response, err
}
