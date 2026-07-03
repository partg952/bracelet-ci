package models

type Job struct {
	JobId      string  `json:"job_id"`
	ProjectId  string  `json:"project_id"`
	CommitSHA  string  `json:"commit_sha"`
	FinishedAt *string `json:"finished_at"`
}
