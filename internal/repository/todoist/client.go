package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/config"
	todoistDomain "hufschlaeger.net/gitlab-tasks-exporter/internal/domain/todoist"
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
		baseURL:    "https://api.todoist.com/rest/v2",
	}
}

// Project operations

func (r *Repository) GetProjects() ([]todoistDomain.Project, error) {
	url := fmt.Sprintf("%s/projects", r.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)

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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get projects failed %d: %s", resp.StatusCode, string(body))
	}

	var projects []todoistDomain.Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	return projects, err
}

func (r *Repository) CreateProject(name string) (*todoistDomain.Project, error) {
	url := fmt.Sprintf("%s/projects", r.baseURL)

	projectData := map[string]interface{}{
		"name":  name,
		"color": "blue",
	}

	jsonData, err := json.Marshal(projectData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create project failed %d: %s", resp.StatusCode, string(body))
	}

	var project todoistDomain.Project
	err = json.NewDecoder(resp.Body).Decode(&project)
	return &project, err
}

// Section operations

func (r *Repository) GetProjectSections(projectID string) ([]todoistDomain.Section, error) {
	url := fmt.Sprintf("%s/sections?project_id=%s", r.baseURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)

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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get sections failed %d: %s", resp.StatusCode, string(body))
	}

	var sections []todoistDomain.Section
	err = json.NewDecoder(resp.Body).Decode(&sections)
	return sections, err
}

func (r *Repository) CreateSection(projectID string, name string, order int) (*todoistDomain.Section, error) {
	url := fmt.Sprintf("%s/sections", r.baseURL)

	sectionData := todoistDomain.Section{
		ProjectID: projectID,
		Name:      name,
		Order:     order,
	}

	jsonData, err := json.Marshal(sectionData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create section failed %d: %s", resp.StatusCode, string(body))
	}

	var section todoistDomain.Section
	err = json.NewDecoder(resp.Body).Decode(&section)
	return &section, err
}

// Task operations

func (r *Repository) GetProjectTasks(projectID string) ([]todoistDomain.Task, error) {
	url := fmt.Sprintf("%s/tasks?project_id=%s", r.baseURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)

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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get tasks failed %d: %s", resp.StatusCode, string(body))
	}

	var tasks []todoistDomain.Task
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	return tasks, err
}

func (r *Repository) CreateTask(taskRequest todoistDomain.CreateTaskRequest) (*todoistDomain.Task, error) {
	url := fmt.Sprintf("%s/tasks", r.baseURL)

	jsonData, err := json.Marshal(taskRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create task failed %d: %s", resp.StatusCode, string(body))
	}

	var task todoistDomain.Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	return &task, err
}

func (r *Repository) UpdateTask(taskID string, updates map[string]interface{}) (*todoistDomain.Task, error) {
	url := fmt.Sprintf("%s/tasks/%s", r.baseURL, taskID)

	jsonData, err := json.Marshal(updates)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.config.TodoistToken)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update task failed %d: %s", resp.StatusCode, string(body))
	}

	var task todoistDomain.Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	return &task, err
}

// ValidateConnection prüft ob die Todoist-Verbindung funktioniert
func (r *Repository) ValidateConnection() error {
	_, err := r.GetProjects()
	if err != nil {
		return fmt.Errorf("todoist connection failed: %w", err)
	}
	return nil
}

// FindProjectByName sucht Projekt nach Namen
func (r *Repository) FindProjectByName(name string) (*todoistDomain.Project, error) {
	projects, err := r.GetProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, nil // Nicht gefunden
}

// FindSectionByName sucht Section nach Namen in einem Projekt
func (r *Repository) FindSectionByName(projectID, name string) (*todoistDomain.Section, error) {
	sections, err := r.GetProjectSections(projectID)
	if err != nil {
		return nil, err
	}

	for _, section := range sections {
		if section.Name == name {
			return &section, nil
		}
	}

	return nil, nil // Nicht gefunden
}

// FindTaskByTitle sucht Task nach Content in einem Projekt
func (r *Repository) FindTaskByTitle(projectID, title string) (*todoistDomain.Task, error) {
	tasks, err := r.GetProjectTasks(projectID)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.Content == title {
			return &task, nil
		}
	}

	return nil, nil // Nicht gefunden
}
