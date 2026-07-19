package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

const heartbeatPrefix = "heartbeat:"

func watchExpiredHeartbeats(ctx context.Context, client *redis.Client, queue string) {
	pubsub := client.PSubscribe(ctx, "__keyevent@*__:expired")
	defer pubsub.Close()

	log.Printf("[heartbeat] watching for expired job heartbeats")

	for msg := range pubsub.Channel() {
		key := msg.Payload
		if !strings.HasPrefix(key, heartbeatPrefix) {
			continue
		}
		jobId := strings.TrimPrefix(key, heartbeatPrefix)
		log.Printf("[heartbeat] job %s expired — re-queuing", jobId)
		if err := client.LPush(ctx, queue, jobId).Err(); err != nil {
			log.Printf("[heartbeat] failed to re-queue %s: %v", jobId, err)
		}
	}
}

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

func parsePushEvent(body []byte, contentType string) (PushEvent, error) {
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return PushEvent{}, fmt.Errorf("failed to parse form payload: %w", err)
		}

		payload := values.Get("payload")
		if payload == "" {
			return PushEvent{}, fmt.Errorf("form payload is missing")
		}
		body = []byte(payload)
	}

	var event PushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return PushEvent{}, fmt.Errorf("failed to unmarshal push event: %w", err)
	}

	return event, nil
}

func QueryProjectIdByRepoUrl(ctx *gin.Context, dbServiceURL string, repoUrl string) (string, error) {
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

func SendDBEvent(ctx *gin.Context, dbServiceURL string, event DBEvent) error {
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
		if githubEvent := c.GetHeader("X-GitHub-Event"); githubEvent != "" && githubEvent != "push" {
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "event": githubEvent})
			return
		}

		job_id := uuid.New()
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		event, err := parsePushEvent(body, c.GetHeader("Content-Type"))
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if event.Repository.CloneUrl == "" || event.After == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "push event is missing repository.clone_url or after"})
			return
		}
		jobInstance := Job{
			JobId:     job_id.String(),
			RepoUrl:   event.Repository.CloneUrl,
			CommitSha: event.After,
			CreatedAt: time.Now(),
		}

		// Look up the project that owns this repository so the job is linked.
		projectId, err := QueryProjectIdByRepoUrl(c, dbServiceURL, event.Repository.CloneUrl)
		if err != nil {
			log.Printf("[webhook] could not resolve project for repo %s: %v", event.Repository.CloneUrl, err)
			// Non-fatal: continue without a project_id rather than dropping the job.
		} else {
			jobInstance.ProjectId = projectId
		}
		if err := SendDBEvent(c, dbServiceURL, DBEvent{
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

	go watchExpiredHeartbeats(context.Background(), client, "job_queue")

	r := gin.Default()

	r.POST("/webhook", webhookHandler(dbServiceURL, client))

	r.Run()
}
