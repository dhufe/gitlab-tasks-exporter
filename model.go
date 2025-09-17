package main

import "net/http"

type Config struct {
	Structured     bool
	GitLabURL      string
	Token          string
	ProjectPath    string
	MilestoneTitle *string
	OutputFile     string
	AssignedUser   *string
	ExportMarkdown bool
	MarkdownFile   string
}

type authTransport struct {
	token string
	base  http.RoundTripper
}

type User struct {
	Name     string `graphql:"name"`
	Username string `graphql:"username"`
}

type Label struct {
	Title string `graphql:"title"`
}

type Milestone struct {
	ID          string `graphql:"id"`
	Title       string `graphql:"title"`
	Description string `graphql:"description"`
	DueDate     string `graphql:"dueDate"`
	State       string `graphql:"state"`
}

type Issue struct {
	IID         string     `graphql:"iid"`
	Title       string     `graphql:"title"`
	Description string     `graphql:"description"`
	State       string     `graphql:"state"`
	DueDate     *string    `graphql:"dueDate"`
	CreatedAt   string     `graphql:"createdAt"`
	UpdatedAt   string     `graphql:"updatedAt"`
	WebURL      string     `graphql:"webUrl"`
	Milestone   *Milestone `graphql:"milestone"`
	Assignees   struct {
		Nodes []User `graphql:"nodes"`
	} `graphql:"assignees"`
	Labels struct {
		Nodes []Label `graphql:"nodes"`
	} `graphql:"labels"`
}

type PageInfo struct {
	HasNextPage bool    `graphql:"hasNextPage"`
	EndCursor   *string `graphql:"endCursor"`
}

type ProjectQuery struct {
	Project struct {
		Milestones struct {
			Nodes []Milestone `graphql:"nodes"`
		} `graphql:"milestones(searchTitle: $milestoneSearch, first: 1)"`
		Issues struct {
			Nodes    []Issue  `graphql:"nodes"`
			PageInfo PageInfo `graphql:"pageInfo"`
		} `graphql:"issues(first: $first, milestoneTitle: $milestoneTitle, assigneeUsername: $assigneeUsername, after: $after)"`
	} `graphql:"project(fullPath: $projectPath)"`
}

type TodoistRecord struct {
	Type        string
	Content     string
	Description string
	Priority    string
	Indent      string
	Author      string
	Responsible string
	Date        string
	DateLang    string
	Timezone    string
	Labels      string
}
