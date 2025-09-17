package todoist

type Task struct {
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	ProjectID   string   `json:"project_id"`
	SectionID   string   `json:"section_id,omitempty"`
	Completed   bool     `json:"is_completed"`
	Labels      []string `json:"labels"`
	Priority    int      `json:"priority"`
	DueDate     string   `json:"due_date,omitempty"`
	URL         string   `json:"url,omitempty"`
}

type CreateTaskRequest struct {
	Content     string   `json:"content"`
	Description string   `json:"description,omitempty"`
	ProjectID   string   `json:"project_id"`
	SectionID   string   `json:"section_id,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	DueString   string   `json:"due_string,omitempty"`
}

type Project struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Section struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Order     int    `json:"order"`
}
