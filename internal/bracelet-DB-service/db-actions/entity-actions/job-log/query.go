package joblogactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"time"
)

func (e *JobLogEditor) Query(operation string) (any, error) {
	switch operation {
	case "get_by_job_id":
		return e.getByJobId()
	default:
		return nil, fmt.Errorf("[JobLog Query Error] unsupported operation %q", operation)
	}
}

func (e *JobLogEditor) getByJobId() (any, error) {
	log, ok := e.event.EntityData.(models.JobLog)
	if !ok {
		return nil, fmt.Errorf("[JobLog Query Error] invalid job_log data")
	}
	if log.JobId == "" {
		return nil, fmt.Errorf("[JobLog Query Error] job_id is required")
	}

	rows, err := e.dbConn.FetchRecords(
		`SELECT id, log_data, created_at FROM job_logs WHERE job_id = $1 ORDER BY created_at ASC`,
		log.JobId,
	)
	if err != nil {
		return nil, fmt.Errorf("[JobLog Query Error] %w", err)
	}
	defer rows.Close()

	logs := make([]models.JobLog, 0)
	for rows.Next() {
		var logId, logData string
		var createdAt time.Time

		if err := rows.Scan(&logId, &logData, &createdAt); err != nil {
			return nil, fmt.Errorf("[JobLog Query Error] %w", err)
		}

		logs = append(logs, models.JobLog{
			LogId:     logId,
			JobId:     log.JobId,
			LogData:   &logData,
			CreatedAt: &createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[JobLog Query Error] %w", err)
	}

	return logs, nil
}
