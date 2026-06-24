package main

import (
	dbpkg "bracelet-cicd/internal/db"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type RepoInfo struct {
	CloneUrl string `json:"clone_url"`
}

type PushEvent struct {
	Repository RepoInfo `json:"repository"`
	After      string   `json:"after"`
}

type Job struct {
	JobId      string `json:"job_id"`
	RepoUrl    string `json:"repo_url"`
	CommitSha  string `json:"commit_sha"`
	CreatedAt  time.Time `json:"created_at"`
	FinishedAt time.Time `json:"finished_at"`
}

func webhookHandler(dbInst *dbpkg.DBInstance, client *redis.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		job_id := uuid.New()
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		var event PushEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON unmarshalling failed"})
			return
		}
		jobInstance := Job{
			JobId:      job_id.String(),
			RepoUrl:    event.Repository.CloneUrl,
			CommitSha:  event.After,
			CreatedAt:  time.Now(),
		}
		err = dbInst.ExecuteQuery(`INSERT INTO jobs(id, repo_url, commit_sha, created_at, finished_at) VALUES($1, $2, $3, $4, $5)`, jobInstance.JobId, jobInstance.RepoUrl, jobInstance.CommitSha, jobInstance.CreatedAt, jobInstance.FinishedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		if err := client.LPush(ctx, "job_queue", jobInstance.JobId).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	}
}

func init() {
	godotenv.Load()
}
func main() {
	dbInst, err := dbpkg.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbInst.Conn.Close()

	opts, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Redis URL parse failed: %v", err)
	}
	client := redis.NewClient(opts)
	defer client.Close()

	r := gin.Default()

	r.POST("/webhook", webhookHandler(&dbInst, client))

	r.Run()
}



