package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	JobId     string    `json:"job_id"`
	ProjectId string    `json:"project_id,omitempty"`
	RepoUrl   string    `json:"repo_url"`
	CommitSha string    `json:"commit_sha"`
	CreatedAt time.Time `json:"created_at"`
}

type DBEvent struct {
	Method     string `json:"method"`
	EntityName string `json:"entity_name"`
	Operation  string `json:"operation,omitempty"`
	EntityData any    `json:"entity_data"`
}

// queryProjectIdByRepoUrl asks the DB service for the project ID that owns the
// given repository URL. Returns an empty string if none is found.
func queryProjectIdByRepoUrl(ctx *gin.Context, dbServiceURL string, repoUrl string) (string, error) {
	event := DBEvent{
		Method:     "query",
		EntityName: "project",
		Operation:  "get_by_repo_url",
		EntityData: map[string]string{"repo_url": repoUrl},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(
		ctx.Request.Context(),
		http.MethodPost,
		strings.TrimRight(dbServiceURL, "/")+"/event",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("database service returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	// Response is a Project object: {"project_id":"..."}
	var project struct {
		ProjectId string `json:"project_id"`
	}
	if err := json.Unmarshal(body, &project); err != nil {
		return "", fmt.Errorf("failed to decode project response: %w", err)
	}

	return project.ProjectId, nil
}

func sendDBEvent(ctx *gin.Context, dbServiceURL string, event DBEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		ctx.Request.Context(),
		http.MethodPost,
		strings.TrimRight(dbServiceURL, "/")+"/event",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("database service returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func webhookHandler(dbServiceURL string, client *redis.Client) gin.HandlerFunc {
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
			JobId:     job_id.String(),
			RepoUrl:   event.Repository.CloneUrl,
			CommitSha: event.After,
			CreatedAt: time.Now(),
		}

		// Look up the project that owns this repository so the job is linked.
		projectId, err := queryProjectIdByRepoUrl(c, dbServiceURL, event.Repository.CloneUrl)
		if err != nil {
			log.Printf("[webhook] could not resolve project for repo %s: %v", event.Repository.CloneUrl, err)
			// Non-fatal: continue without a project_id rather than dropping the job.
		} else {
			jobInstance.ProjectId = projectId
		}
		if err := sendDBEvent(c, dbServiceURL, DBEvent{
			Method:     "create",
			EntityName: "job",
			EntityData: jobInstance,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := client.LPush(c.Request.Context(), "job_queue", jobInstance.JobId).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	}
}

func init() {
	godotenv.Load("../../.env")
}
func main() {
	dbServiceURL := os.Getenv("DB_SERVICE_URL")
	if dbServiceURL == "" {
		dbServiceURL = "http://localhost:8081"
	}

	opts, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Redis URL parse failed: %v", err)
	}
	client := redis.NewClient(opts)
	defer client.Close()

	r := gin.Default()

	r.POST("/webhook", webhookHandler(dbServiceURL, client))

	r.Run()
}
