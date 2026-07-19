package logs

import (
	"bracelet-cicd/internal/bracelet-worker/dbclient"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const flushSize = 10

type Logs struct {
	mu     sync.Mutex
	buffer []Log
	jobId  string
	client *dbclient.Client
}

type Log struct {
	Timestamp time.Time
	Contents  string
}

func New(jobId string, client *dbclient.Client) *Logs {
	return &Logs{
		buffer: make([]Log, 0, flushSize),
		jobId:  jobId,
		client: client,
	}
}

// Push appends a log line. Triggers a flush if the buffer hits flushSize.
func (l *Logs) Push(entry Log) {
	l.mu.Lock()
	l.buffer = append(l.buffer, entry)
	ready := len(l.buffer) >= flushSize
	l.mu.Unlock()

	if ready {
		_ = l.Flush()
	}
}

// Flush sends all buffered lines as a single chunk to the DB service.
func (l *Logs) Flush() error {
	l.mu.Lock()
	if len(l.buffer) == 0 {
		l.mu.Unlock()
		return nil
	}
	pending := append([]Log(nil), l.buffer...)
	l.buffer = l.buffer[:0]
	l.mu.Unlock()

	lines := make([]string, len(pending))
	for i, entry := range pending {
		ts := entry.Timestamp
		if ts.IsZero() {
			ts = time.Now().UTC()
		}
		lines[i] = fmt.Sprintf("[%s] %s", ts.UTC().Format(time.RFC3339), entry.Contents)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return l.client.Send(ctx, dbclient.DBEvent{
		Method:     "create",
		EntityName: "job_log",
		EntityData: map[string]any{
			"log_id":     uuid.NewString(),
			"job_id":     l.jobId,
			"log_data":   strings.Join(lines, "\n"),
			"created_at": pending[0].Timestamp.UTC(),
		},
	}, nil)
}

// StartFlusher flushes every 10 seconds and performs a final flush on ctx cancel.
func (l *Logs) StartFlusher(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				_ = l.Flush()
				return
			case <-ticker.C:
				_ = l.Flush()
			}
		}
	}()
}
