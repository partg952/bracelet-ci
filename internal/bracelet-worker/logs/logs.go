package logs

import (
	"bracelet-cicd/internal/bracelet-worker/worker"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Logs struct {
	mu sync.Mutex
	buffer []Log
	jobId  string
}

type Log struct {
	timestamp time.Time
	contents  string
}

func New(jobId string) *Logs {
	return &Logs{
		buffer: []Log{},
		jobId:  jobId,
	}
}

func (l *Logs) Flush(dbWorker *worker.Worker) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, log := range l.buffer {
		content := log.contents
		createdAt := log.timestamp
		if err := dbWorker.SendDBEvent(ctx, worker.DBEvent{
			Method:     "create",
			EntityName: "job_log",
			EntityData: map[string]any{
				"log_id":     uuid.NewString(),
				"job_id":     l.jobId,
				"log_data":   content,
				"created_at": createdAt,
			},
		}, nil); err != nil {
			return err
		}
	}

	l.buffer = l.buffer[:0]
	return nil
}

func (l *Logs) Push(log Log) {
	l.mu.Lock()
	l.buffer = append(l.buffer, log)
	l.mu.Unlock()
}