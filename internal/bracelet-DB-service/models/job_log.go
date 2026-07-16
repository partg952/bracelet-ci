package models

import "time"

type JobLog struct {
	LogId     string     `json:"log_id,omitempty"`
	JobId     string     `json:"job_id,omitempty"`
	LogData   *string    `json:"log_data,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
