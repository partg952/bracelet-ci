package worker

import (
	"bracelet-cicd/internal/bracelet-worker/dbclient"
	dockerexecutors "bracelet-cicd/internal/bracelet-worker/docker-executors"
	worklogs "bracelet-cicd/internal/bracelet-worker/logs"
	"bracelet-cicd/internal/bracelet-worker/parser"
	"bracelet-cicd/internal/bracelet-worker/repository"
	"bracelet-cicd/internal/bracelet-worker/testrunner"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	heartbeatTTL      = 90 * time.Second
	heartbeatInterval = 30 * time.Second
	heartbeatPrefix   = "heartbeat:"
)

type Worker struct {
	id          string
	redisClient *redis.Client
	queue       string
	timeout     time.Duration
	db          *dbclient.Client
}

type Status string

const (
	Running Status = "Running"
	Failed  Status = "Failed"
	Passed  Status = "Passed"
)

type JobDetails struct {
	JobId     string `json:"job_id"`
	RepoUrl   string `json:"repo_url"`
	CommitSHA string `json:"commit_sha"`
}

func New(redisClient *redis.Client, dbService string, queue string, timeout time.Duration) *Worker {
	if redisClient == nil {
		log.Fatal("redis client is required")
	}
	if dbService == "" {
		log.Fatal("database service URL is required")
	}
	if queue == "" {
		log.Fatal("queue name is required")
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	return &Worker{
		id:          uuid.NewString(),
		redisClient: redisClient,
		queue:       queue,
		timeout:     timeout,
		db:          dbclient.New(dbService, httpClient),
	}
}

func (w *Worker) SendDBEvent(ctx context.Context, event dbclient.DBEvent, result any) error {
	return w.db.Send(ctx, event, result)
}

func heartbeatKey(jobId string) string {
	return heartbeatPrefix + jobId
}

func (w *Worker) startHeartbeat(ctx context.Context, jobId string) context.CancelFunc {
	hbCtx, cancel := context.WithCancel(ctx)

	set := func() {
		w.redisClient.Set(hbCtx, heartbeatKey(jobId), w.id, heartbeatTTL)
	}
	set()

	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				w.redisClient.Del(context.Background(), heartbeatKey(jobId))
				return
			case <-ticker.C:
				set()
			}
		}
	}()

	return cancel
}


func (w *Worker) Start(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		values, err := w.redisClient.BRPop(ctx, w.timeout, w.queue).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			log.Printf("redis BRPop failed: %v", err)
			continue
		}

		jobId := values[1]
		log.Printf("received queue message: %s", jobId)

		// Start heartbeat — cancel deletes the key when job finishes
		stopHeartbeat := w.startHeartbeat(ctx, jobId)

		jobCtx, cancelJob := context.WithCancel(ctx)
		jobLog := worklogs.New(jobId, w.db)
		jobLog.StartFlusher(jobCtx)

		runningCtx, cancelRunning := context.WithTimeout(ctx, 5*time.Second)
		_ = w.db.Send(runningCtx, dbclient.DBEvent{
			Method:     "update",
			EntityName: "job",
			EntityData: map[string]any{"job_id": jobId, "status": Running},
		}, nil)
		cancelRunning()

		fetchCtx, cancelFetch := context.WithTimeout(ctx, 5*time.Second)
		var data JobDetails
		err = w.db.Send(fetchCtx, dbclient.DBEvent{
			Method:     "read",
			EntityName: "job",
			EntityData: map[string]any{"job_id": jobId},
		}, &data)
		cancelFetch()
		if err != nil {
			log.Printf("job %s: database service read failed: %v", jobId, err)
			cancelJob()
			stopHeartbeat()
			continue
		}

		// Clone
		jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("cloning %s @ %s", data.RepoUrl, data.CommitSHA)})
		pathname, cloneOutput, err := repository.Clone(data.RepoUrl, data.CommitSHA)
		for _, line := range strings.Split(strings.TrimSpace(cloneOutput), "\n") {
			if strings.TrimSpace(line) != "" {
				jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: line})
			}
		}
		if err != nil {
			log.Printf("job %s: repository clone failed: %v", jobId, err)
			jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("clone failed: %v", err)})
			cancelJob()
			stopHeartbeat()
			continue
		}
		log.Printf("successfully cloned the repo: %v", pathname)

		// Parse YAML
		parsedYaml, err := parser.ParseYaml(pathname)
		if err != nil {
			log.Printf("job %s: yaml parse failed: %v", jobId, err)
			jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("yaml parse failed: %v", err)})
			cancelJob()
			stopHeartbeat()
			continue
		}

		// Run tests
		jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: "starting test runner"})
		dockerInst := dockerexecutors.New(jobId, pathname)
		testResults, err := testrunner.Run(&dockerInst, parsedYaml)
		for _, line := range strings.Split(strings.TrimSpace(testResults.Output), "\n") {
			if strings.TrimSpace(line) != "" {
				jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: line})
			}
		}
		if err != nil {
			log.Printf("job %s: test runner failed: %v", jobId, err)
			jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("test runner error: %v", err)})
			cancelJob()
			stopHeartbeat()
			continue
		}

		status := Failed
		if testResults.Passed {
			status = Passed
		}
		log.Printf("job status: %v", status)
		jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("job finished: %s", status)})

		// Flush logs then stop heartbeat
		cancelJob()
		stopHeartbeat()

		// Update job status
		updateCtx, cancelUpdate := context.WithTimeout(ctx, 5*time.Second)
		err = w.db.Send(updateCtx, dbclient.DBEvent{
			Method:     "update",
			EntityName: "job",
			EntityData: map[string]any{
				"job_id":      jobId,
				"status":      status,
				"finished_at": time.Now().UTC(),
			},
		}, nil)
		cancelUpdate()
		if err != nil {
			log.Printf("job %s: database service update failed: %v", jobId, err)
			continue
		}
		log.Printf("job %s: status updated successfully", jobId)
	}
}
