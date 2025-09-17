package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TodoistAPI struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewTodoistAPI(token string) *TodoistAPI {
	return &TodoistAPI{
		token:      token,
		httpClient: &http.Client{},
		baseURL:    "https://api.todoist.com/rest/v2",
	}
}

func (api *TodoistAPI) GetProjectTasks(projectID string) ([]TodoistTask, error) {
	url := fmt.Sprintf("%s/tasks?project_id=%s", api.baseURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.token)

	resp, err := api.httpClient.Do(req)
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
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var tasks []TodoistTask
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	return tasks, err
}

func (api *TodoistAPI) TaskExists(projectID, content string) (bool, string, error) {
	tasks, err := api.GetProjectTasks(projectID)
	if err != nil {
		return false, "", err
	}

	for _, task := range tasks {
		if task.Content == content {
			return true, task.ID, nil
		}
	}

	return false, "", nil
}

func (api *TodoistAPI) GetProjects() ([]TodoistProject, error) {
	url := fmt.Sprintf("%s/projects", api.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.token)

	resp, err := api.httpClient.Do(req)
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
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var projects []TodoistProject
	err = json.NewDecoder(resp.Body).Decode(&projects)
	return projects, err
}

func (api *TodoistAPI) CreateProject(name string) (*TodoistProject, error) {
	url := fmt.Sprintf("%s/projects", api.baseURL)

	projectRequest := map[string]interface{}{
		"name":  name,
		"color": "blue",
	}

	jsonData, err := json.Marshal(projectRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
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
		return nil, fmt.Errorf("project creation failed %d: %s", resp.StatusCode, string(body))
	}

	var project TodoistProject
	err = json.NewDecoder(resp.Body).Decode(&project)
	return &project, err
}

func (api *TodoistAPI) CreateSection(projectID, name string, order int) (*TodoistSection, error) {
	url := fmt.Sprintf("%s/sections", api.baseURL)

	sectionRequest := TodoistSection{
		ProjectID: projectID,
		Name:      name,
		Order:     order,
	}

	jsonData, err := json.Marshal(sectionRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
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
		return nil, fmt.Errorf("section creation failed %d: %s", resp.StatusCode, string(body))
	}

	var section TodoistSection
	err = json.NewDecoder(resp.Body).Decode(&section)
	return &section, err
}

func (api *TodoistAPI) CreateTask(task TodoistCreateTaskRequest) (*TodoistTask, error) {
	url := fmt.Sprintf("%s/tasks", api.baseURL)

	jsonData, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
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
		return nil, fmt.Errorf("task creation failed %d: %s", resp.StatusCode, string(body))
	}

	var createdTask TodoistTask
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	return &createdTask, err
}
