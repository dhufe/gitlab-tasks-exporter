package main

import "time"

type Config struct {
	Structured     bool
	GitLabURL      string
	Token          string
	ProjectPath    string
	MilestoneTitle *string
	OutputFile     string
}

type Label struct {
	Title       string `graphql:"title"`
	Color       string `graphql:"color"`
	Description string `graphql:"description"`
}

type User struct {
	Name     string `graphql:"name"`
	Username string `graphql:"username"`
	Email    string `graphql:"email"`
}

type Milestone struct {
	Title       string     `graphql:"title"`
	Description string     `graphql:"description"`
	DueDate     *time.Time `graphql:"dueDate"`
	State       string     `graphql:"state"`
}

type TimeStats struct {
	TimeEstimate   int `graphql:"timeEstimate"`
	TotalTimeSpent int `graphql:"totalTimeSpent"`
}

type TaskCompletionStatus struct {
	CompletedCount int `graphql:"completedCount"`
	Count          int `graphql:"count"`
}

type Issue struct {
	IID          int        `graphql:"iid"`
	Title        string     `graphql:"title"`
	Description  string     `graphql:"description"`
	State        string     `graphql:"state"`
	DueDate      *time.Time `graphql:"dueDate"`
	CreatedAt    time.Time  `graphql:"createdAt"`
	UpdatedAt    time.Time  `graphql:"updatedAt"`
	ClosedAt     *time.Time `graphql:"closedAt"`
	Confidential bool       `graphql:"confidential"`
	WebURL       string     `graphql:"webUrl"`
	Weight       *int       `graphql:"weight"`
	TimeStats    TimeStats  `graphql:"timeStats"`
	Labels       struct {
		Nodes []Label `graphql:"nodes"`
	} `graphql:"labels"`
	Assignees struct {
		Nodes []User `graphql:"nodes"`
	} `graphql:"assignees"`
	Author               User                 `graphql:"author"`
	Milestone            *Milestone           `graphql:"milestone"`
	TaskCompletionStatus TaskCompletionStatus `graphql:"taskCompletionStatus"`
}

type PageInfo struct {
	HasNextPage bool   `graphql:"hasNextPage"`
	EndCursor   string `graphql:"endCursor"`
}

type IssuesQuery struct {
	Project struct {
		Issues struct {
			Nodes    []Issue  `graphql:"nodes"`
			PageInfo PageInfo `graphql:"pageInfo"`
		} `graphql:"issues(first: 100, milestoneTitle: $milestoneTitle, after: $after)"`
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
