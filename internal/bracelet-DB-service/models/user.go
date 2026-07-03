package models

type User struct {
	UserId   string  `json:"user_id,omitempty"`
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type UserDetails struct {
	UserId       string `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	ProjectCount int64  `json:"project_count"`
	JobCount     int64  `json:"job_count"`
}
