package joblogactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/db"
	dbactions "bracelet-cicd/internal/bracelet-DB-service/db-actions"
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"time"
)

type JobLogEditor struct {
	event  dbactions.Event
	dbConn *db.DBInstance
}

func NewJobLogEditor(event dbactions.Event, db *db.DBInstance) (JobLogEditor, error) {
	if event.EntityName == "" || event.Method == "" {
		return JobLogEditor{}, fmt.Errorf("empty method or entity")
	}
	return JobLogEditor{event: event, dbConn: db}, nil
}

func (e *JobLogEditor) Create() (any, error) {
	log, ok := e.event.EntityData.(models.JobLog)
	if !ok {
		return nil, fmt.Errorf("[JobLog Create Error] invalid job_log data")
	}
	if log.LogId == "" || log.JobId == "" || log.LogData == nil || log.CreatedAt == nil {
		return nil, fmt.Errorf("[JobLog Create Error] log_id, job_id, log_data, and created_at are required")
	}

	if err := e.dbConn.ExecuteQuery(
		`INSERT INTO job_logs(id, job_id, log_data, created_at) VALUES($1, $2, $3, $4)`,
		log.LogId,
		log.JobId,
		*log.LogData,
		*log.CreatedAt,
	); err != nil {
		return nil, err
	}
	return nil, nil
}

func (e *JobLogEditor) Read() (any, error) {
	log, ok := e.event.EntityData.(models.JobLog)
	if !ok {
		return nil, fmt.Errorf("[JobLog Read Error] invalid job_log data")
	}
	if log.LogId == "" {
		return nil, fmt.Errorf("[JobLog Read Error] log_id is required")
	}

	rows, err := e.dbConn.FetchRecords(
		`SELECT job_id, log_data, created_at FROM job_logs WHERE id = $1`,
		log.LogId,
	)
	if err != nil {
		return nil, fmt.Errorf("[JobLog Read Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[JobLog Read Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[JobLog Read Error] log not found")
	}

	var jobId, logData string
	var createdAt time.Time
	if err := rows.Scan(&jobId, &logData, &createdAt); err != nil {
		return nil, fmt.Errorf("[JobLog Read Error] %w", err)
	}

	return models.JobLog{
		LogId:     log.LogId,
		JobId:     jobId,
		LogData:   &logData,
		CreatedAt: &createdAt,
	}, nil
}

func (e *JobLogEditor) Update() (any, error) {
	log, ok := e.event.EntityData.(models.JobLog)
	if !ok {
		return nil, fmt.Errorf("[JobLog Update Error] invalid job_log data")
	}
	if log.LogId == "" {
		return nil, fmt.Errorf("[JobLog Update Error] log_id is required")
	}
	if log.LogData == nil {
		return nil, fmt.Errorf("[JobLog Update Error] no fields provided to update")
	}

	if err := e.dbConn.ExecuteQuery(
		`UPDATE job_logs SET log_data = $1 WHERE id = $2`,
		*log.LogData,
		log.LogId,
	); err != nil {
		return nil, err
	}
	return nil, nil
}

func (e *JobLogEditor) Delete() (any, error) {
	log, ok := e.event.EntityData.(models.JobLog)
	if !ok {
		return nil, fmt.Errorf("[JobLog Delete Error] invalid job_log data")
	}
	if log.LogId == "" {
		return nil, fmt.Errorf("[JobLog Delete Error] log_id is required")
	}

	if err := e.dbConn.ExecuteQuery(
		`DELETE FROM job_logs WHERE id = $1`, log.LogId,
	); err != nil {
		return nil, err
	}
	return nil, nil
}
