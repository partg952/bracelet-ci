package models

type Project struct {
	ProjectId string  `json:"project_id,omitempty"`
	UserId    *string `json:"user_id,omitempty"`
	RepoUrl   *string `json:"repo_url,omitempty"`
	Name      *string `json:"name,omitempty"`
}

type ProjectDetails struct {
	ProjectId     string `json:"project_id"`
	UserId        string `json:"user_id"`
	RepoUrl       string `json:"repo_url"`
	Name          string `json:"name"`
	OwnerUsername string `json:"owner_username"`
	OwnerEmail    string `json:"owner_email"`
	JobCount      int64  `json:"job_count"`
}
