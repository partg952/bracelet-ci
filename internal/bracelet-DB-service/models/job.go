package models

import "time"

type Job struct {
	JobId      string     `json:"job_id,omitempty"`
	ProjectId  *string    `json:"project_id,omitempty"`
	RepoUrl    *string    `json:"repo_url,omitempty"`
	CommitSHA  *string    `json:"commit_sha,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Status     *string    `json:"status,omitempty"`
}
