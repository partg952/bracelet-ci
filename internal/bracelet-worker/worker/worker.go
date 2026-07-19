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

	"github.com/redis/go-redis/v9"
)

type Worker struct {
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
		redisClient: redisClient,
		queue:       queue,
		timeout:     timeout,
		db:          dbclient.New(dbService, httpClient),
	}
}

func (w *Worker) SendDBEvent(ctx context.Context, event dbclient.DBEvent, result any) error {
	return w.db.Send(ctx, event, result)
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

		job_id := values[1]
		log.Printf("received queue message: %s", job_id)

		jobCtx, cancelJob := context.WithCancel(ctx)
		jobLog := worklogs.New(job_id, w.db)
		jobLog.StartFlusher(jobCtx)

		runningCtx, cancelRunning := context.WithTimeout(ctx, 5*time.Second)
		_ = w.db.Send(runningCtx, dbclient.DBEvent{
			Method:     "update",
			EntityName: "job",
			EntityData: map[string]any{"job_id": job_id, "status": Running},
		}, nil)
		cancelRunning()

		fetchCtx, cancelFetch := context.WithTimeout(ctx, 5*time.Second)
		var data JobDetails
		err = w.db.Send(fetchCtx, dbclient.DBEvent{
			Method:     "read",
			EntityName: "job",
			EntityData: map[string]any{"job_id": job_id},
		}, &data)
		cancelFetch()
		if err != nil {
			log.Printf("job %s: database service read failed: %v", job_id, err)
			cancelJob()
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
			log.Printf("job %s: repository clone failed: %v", job_id, err)
			jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("clone failed: %v", err)})
			cancelJob()
			continue
		}
		log.Printf("successfully cloned the repo: %v", pathname)

		// Parse YAML
		parsedYaml, err := parser.ParseYaml(pathname)
		if err != nil {
			log.Printf("job %s: yaml parse failed: %v", job_id, err)
			jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("yaml parse failed: %v", err)})
			cancelJob()
			continue
		}

		// Run tests
		jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: "starting test runner"})
		dockerInst := dockerexecutors.New(job_id, pathname)
		testResults, err := testrunner.Run(&dockerInst, parsedYaml)
		for _, line := range strings.Split(strings.TrimSpace(testResults.Output), "\n") {
			if strings.TrimSpace(line) != "" {
				jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: line})
			}
		}
		if err != nil {
			log.Printf("job %s: test runner failed: %v", job_id, err)
			jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("test runner error: %v", err)})
			cancelJob()
			continue
		}

		status := Failed
		if testResults.Passed {
			status = Passed
		}
		log.Printf("job status: %v", status)
		jobLog.Push(worklogs.Log{Timestamp: time.Now(), Contents: fmt.Sprintf("job finished: %s", status)})

		// Final flush
		cancelJob()

		// Update job status
		updateCtx, cancelUpdate := context.WithTimeout(ctx, 5*time.Second)
		err = w.db.Send(updateCtx, dbclient.DBEvent{
			Method:     "update",
			EntityName: "job",
			EntityData: map[string]any{
				"job_id":      job_id,
				"status":      status,
				"finished_at": time.Now().UTC(),
			},
		}, nil)
		cancelUpdate()
		if err != nil {
			log.Printf("job %s: database service update failed: %v", job_id, err)
			continue
		}
		log.Printf("job %s: status updated successfully", job_id)
	}
}
