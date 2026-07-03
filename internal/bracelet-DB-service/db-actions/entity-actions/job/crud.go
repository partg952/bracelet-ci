package jobactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/db"
	dbactions "bracelet-cicd/internal/bracelet-DB-service/db-actions"
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"strings"
)

type JobEditor struct {
	event  dbactions.Event
	dbConn *db.DBInstance
}

func NewJobEditor(event dbactions.Event, db *db.DBInstance) (JobEditor, error) {
	if event.EntityName == "" || event.Method == "" {
		return JobEditor{}, fmt.Errorf("Empty method or entity")
	}

	return JobEditor{
		event:  event,
		dbConn: db,
	}, nil
}

func (jEditor *JobEditor) Create() (any, error) {
	job, ok := jEditor.event.EntityData.(models.Job)
	if !ok {
		return nil, fmt.Errorf("[Job Create Error] invalid job data")
	}
	if job.JobId == "" || job.ProjectId == "" || job.CommitSHA == "" {
		return nil, fmt.Errorf("[Job Create Error] job_id, project_id, and commit_sha are required")
	}

	if err := jEditor.dbConn.ExecuteQuery(
		`INSERT INTO "job"(id, project_id, commit_sha, finished_at) VALUES($1, $2, $3, $4)`,
		job.JobId,
		job.ProjectId,
		job.CommitSHA,
		job.FinishedAt,
	); err != nil {
		return nil, err
	}

	return nil, nil
}

func (jEditor *JobEditor) Read() (any, error) {
	job, ok := jEditor.event.EntityData.(models.Job)
	if !ok {
		return nil, fmt.Errorf("[Job Read Error] invalid job data")
	}
	if job.JobId == "" {
		return nil, fmt.Errorf("[Job Read Error] job_id is required")
	}

	rows, err := jEditor.dbConn.FetchRecords(
		`SELECT project_id, commit_sha, finished_at FROM "job" WHERE id = $1`,
		job.JobId,
	)
	if err != nil {
		return nil, fmt.Errorf("[Job Read Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[Job Read Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[Job Read Error] job not found")
	}

	var projectId string
	var commitSha string
	var finishedAt *string
	if err := rows.Scan(&projectId, &commitSha, &finishedAt); err != nil {
		return nil, fmt.Errorf("[Job Read Error] %w", err)
	}

	return models.Job{
		JobId:      job.JobId,
		ProjectId:  projectId,
		CommitSHA:  commitSha,
		FinishedAt: finishedAt,
	}, nil
}

func (jEditor *JobEditor) Update() (any, error) {
	job, ok := jEditor.event.EntityData.(models.Job)
	if !ok {
		return nil, fmt.Errorf("[Job Update Error] invalid job data")
	}
	if job.JobId == "" {
		return nil, fmt.Errorf("[Job Update Error] job_id is required")
	}

	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if job.ProjectId != "" {
		args = append(args, job.ProjectId)
		setClauses = append(setClauses, fmt.Sprintf("project_id = $%d", len(args)))
	}
	if job.CommitSHA != "" {
		args = append(args, job.CommitSHA)
		setClauses = append(setClauses, fmt.Sprintf("commit_sha = $%d", len(args)))
	}
	if job.FinishedAt != nil {
		args = append(args, *job.FinishedAt)
		setClauses = append(setClauses, fmt.Sprintf("finished_at = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("[Job Update Error] no fields provided to update")
	}

	args = append(args, job.JobId)
	query := fmt.Sprintf(
		`UPDATE "job" SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "),
		len(args),
	)

	if err := jEditor.dbConn.ExecuteQuery(query, args...); err != nil {
		return nil, err
	}

	return nil, nil
}

func (jEditor *JobEditor) Delete() (any, error) {
	job, ok := jEditor.event.EntityData.(models.Job)
	if !ok {
		return nil, fmt.Errorf("[Job Delete Error] invalid job data")
	}
	if job.JobId == "" {
		return nil, fmt.Errorf("[Job Delete Error] job_id is required")
	}

	if err := jEditor.dbConn.ExecuteQuery(`DELETE FROM "job" WHERE id = $1`, job.JobId); err != nil {
		return nil, err
	}

	return nil, nil
}
