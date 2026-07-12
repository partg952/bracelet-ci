package worker

import (
	dockerexecutors "bracelet-cicd/internal/bracelet-worker/docker-executors"
	"bracelet-cicd/internal/bracelet-worker/repository"
	"bracelet-cicd/internal/bracelet-worker/testrunner"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	dbService   string
	httpClient  *http.Client
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
type dbEvent struct {
	Method     string `json:"method"`
	EntityName string `json:"entity_name"`
	EntityData any    `json:"entity_data"`
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

	return &Worker{
		redisClient: redisClient,
		queue:       queue,
		timeout:     timeout,
		dbService:   strings.TrimRight(dbService, "/"),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *Worker) sendDBEvent(ctx context.Context, event dbEvent, result any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		w.dbService+"/event",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := w.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("database service returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	if result == nil || len(bytes.TrimSpace(body)) == 0 || string(bytes.TrimSpace(body)) == "null" {
		return nil
	}

	return json.Unmarshal(body, result)
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

		// Mark job as running immediately so the dashboard reflects it
		runningCtx, cancelRunning := context.WithTimeout(ctx, 5*time.Second)
		_ = w.sendDBEvent(runningCtx, dbEvent{
			Method:     "update",
			EntityName: "job",
			EntityData: map[string]any{
				"job_id": job_id,
				"status": Running,
			},
		}, nil)
		cancelRunning()

		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		var data JobDetails
		err = w.sendDBEvent(fetchCtx, dbEvent{
			Method:     "read",
			EntityName: "job",
			EntityData: map[string]any{"job_id": job_id},
		}, &data)
		cancel()
		if err != nil {
			log.Printf("job %s: database service read failed: %v", job_id, err)
			continue
		}
		pathname, err := repository.Clone(data.RepoUrl, data.CommitSHA)
		if err != nil {
			log.Printf("job %s: repository clone failed: %v", job_id, err)
			continue
		}

		log.Printf("Successfully cloned the repo : %v", pathname)
		dockerInst := dockerexecutors.New(job_id, pathname)
		err = dockerInst.BuildImage()
		if err != nil {
			continue
		}

		testResults, err := testrunner.Run(&dockerInst)
		if err != nil {
			log.Printf("job %s: test runner failed: %v", job_id, err)
			continue
		}

		status := Failed

		if testResults.Passed {
			status = Passed
		}
		log.Printf("Job Status : %v", status)

		updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = w.sendDBEvent(updateCtx, dbEvent{
			Method:     "update",
			EntityName: "job",
			EntityData: map[string]any{
				"job_id":      job_id,
				"status":      status,
				"finished_at": time.Now().UTC(),
			},
		}, nil)
		cancel()
		if err != nil {
			log.Printf("job %s: database service update failed: %v", job_id, err)
			continue
		}
		log.Printf("job %s: status updated successfully", job_id)
	}
}
