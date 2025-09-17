package gitlab

import "time"

type Issue struct {
	IID         string    `json:"iid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	WebURL      string    `json:"web_url"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Labels      Labels    `json:"labels"`
	Assignees   Assignees `json:"assignees"`
}

type Labels struct {
	Nodes []Label `json:"nodes"`
}

type Label struct {
	Title string `json:"title"`
}

type Assignees struct {
	Nodes []Assignee `json:"nodes"`
}

type Assignee struct {
	Name string `json:"name"`
}

type GraphQLResponse struct {
	Data struct {
		Project struct {
			Issues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"issues"`
		} `json:"project"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}
