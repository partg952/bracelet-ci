package jobactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"time"
)

func (jEditor *JobEditor) Query(operation string) (any, error) {
	switch operation {
	case "get_by_project_id":
		return jEditor.getByProjectId()
	default:
		return nil, fmt.Errorf("[Job Query Error] unsupported operation %q", operation)
	}
}

func (jEditor *JobEditor) getByProjectId() (any, error) {
	job, ok := jEditor.event.EntityData.(models.Job)
	if !ok {
		return nil, fmt.Errorf("[Job Query Error] invalid job data")
	}
	if job.ProjectId == nil || *job.ProjectId == "" {
		return nil, fmt.Errorf("[Job Query Error] project_id is required")
	}

	rows, err := jEditor.dbConn.FetchRecords(
		`SELECT id, project_id, repo_url, commit_sha, created_at, finished_at, status
		FROM jobs
		WHERE project_id = $1
		ORDER BY created_at DESC`,
		*job.ProjectId,
	)
	if err != nil {
		return nil, fmt.Errorf("[Job Query Error] %w", err)
	}
	defer rows.Close()

	jobs := make([]models.Job, 0)
	for rows.Next() {
		var result models.Job
		var projectId *string
		var repoUrl *string
		var commitSHA string
		var createdAt time.Time
		var finishedAt *time.Time
		var status *string

		if err := rows.Scan(
			&result.JobId,
			&projectId,
			&repoUrl,
			&commitSHA,
			&createdAt,
			&finishedAt,
			&status,
		); err != nil {
			return nil, fmt.Errorf("[Job Query Error] %w", err)
		}

		result.ProjectId = projectId
		result.RepoUrl = repoUrl
		result.CommitSHA = &commitSHA
		result.CreatedAt = &createdAt
		result.FinishedAt = finishedAt
		result.Status = status
		jobs = append(jobs, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[Job Query Error] %w", err)
	}

	return jobs, nil
}
