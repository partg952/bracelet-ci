package worker

import (
	"bracelet-cicd/internal/bracelet-worker/db"
	dockerexecutors "bracelet-cicd/internal/bracelet-worker/docker-executors"
	"bracelet-cicd/internal/bracelet-worker/repository"
	"bracelet-cicd/internal/bracelet-worker/testrunner"
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	redisClient *redis.Client
	queue       string
	timeout     time.Duration
	db          db.DBInstance
}
type Status string

const (
	Running Status = "Running"
	Failed  Status = "Failed"
	Passed  Status = "Passed"
)

type JobDetails struct {
	JobId     string `db:"id"`
	RepoUrl   string `db:"repo_url"`
	CommitSHA string `db:"commit_sha"`
}

func New(redisClient *redis.Client, dbInstance db.DBInstance, queue string, timeout time.Duration) *Worker {
	if redisClient == nil {
		log.Fatal("redis client is required")
	}
	if queue == "" {
		log.Fatal("queue name is required")
	}

	return &Worker{
		redisClient: redisClient,
		queue:       queue,
		timeout:     timeout,
		db:          dbInstance,
	}
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
		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		rows, err := w.db.FetchResults(fetchCtx, `SELECT id,repo_url,commit_sha FROM jobs WHERE id=$1`, job_id)
		if err != nil {
			cancel()
			log.Printf("job %s: database query failed: %v", job_id, err)
			continue
		}

		data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[JobDetails])
		cancel()
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Printf("job %s: no database record found", job_id)
			} else {
				log.Printf("job %s: could not read query result: %v", job_id, err)
			}
			continue
		}
		pathname, err := repository.Clone(data.RepoUrl, data.CommitSHA)
		if err != nil {
			log.Printf("job %s: repository clone failed: %v", job_id, err)
			continue
		}

		log.Printf("Successfully cloned the repo : %v", pathname)
		dockerInst := dockerexecutors.New(job_id , pathname)
		err = dockerInst.BuildImage()
		if err!=nil {
			continue;
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

		queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = w.db.ExecuteQuery(queryCtx, "UPDATE jobs SET status=$1 WHERE id=$2", status, job_id)
		if err != nil {
			log.Printf("job %s : update query failed : %v", err)
		}
		log.Printf("Update Query successful")
	}

}
