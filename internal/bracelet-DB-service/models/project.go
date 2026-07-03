package models

type Project struct {
	ProjectId string `json:"project_id"`
	UserId    string `json:"user_id"`
	RepoUrl   string `json:"repo_url"`
	Name      string `json:"name"`
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
