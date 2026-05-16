package model

import "time"

type GithubResponse struct {
	Arr []GithubUpdate
}

type GithubUpdate struct {
	ID            int          `json:"id"`
	RepositoryURL string       `json:"repository_url"`
	HTMLURL       string       `json:"html_url"`
	Type          string       `json:"-"`
	Title         string       `json:"title"`
	State         string       `json:"state"`
	User          User         `json:"user"`
	PullRequest   *PullRequest `json:"pull_request,omitempty"`
	Description   string       `json:"body"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type PullRequest struct {
	HTMLURL  string    `json:"html_url"`
	MergedAt time.Time `json:"merged_at"`
}

type User struct {
	Login   string `json:"login"`
	HTMLURL string `json:"html_url"`
}
