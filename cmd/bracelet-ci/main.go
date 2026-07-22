package main

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bracelet-cicd/internal/bracelet-ci/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

//go:embed docker-compose.yml
var embeddedCompose []byte

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


func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}


var rootCmd = &cobra.Command{
	Use:   "bracelet-ci",
	Short: "braceletCI — self-hosted CI orchestrator",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate secrets and write ~/.bracelet-ci/.env and docker-compose.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		envPath, _ := config.EnvFilePath()
		composePath, _ := config.ComposeFilePath()
		cfgPath, _ := config.Path()

		if _, err := os.Stat(envPath); err == nil {
			if _, err := os.Stat(composePath); err == nil {
				if _, err := os.Stat(cfgPath); err == nil {
					fmt.Println("braceletCI already initialised.")
					fmt.Printf("  Config:        %s\n", cfgPath)
					fmt.Printf("  Env file:      %s\n", envPath)
					fmt.Printf("  Compose file:  %s\n", composePath)
					return nil
				}
			}
		}

		workerToken, err := randomHex(32)
		if err != nil {
			return fmt.Errorf("failed to generate worker token: %w", err)
		}
		redisPassword, err := randomHex(24)
		if err != nil {
			return fmt.Errorf("failed to generate redis password: %w", err)
		}
		postgresPassword, err := randomHex(16)
		if err != nil {
			return fmt.Errorf("failed to generate postgres password: %w", err)
		}

		postgresDSN := fmt.Sprintf(
			"postgres://bracelet:%s@localhost:5432/bracelet?sslmode=disable",
			postgresPassword,
		)
		redisURL := fmt.Sprintf("redis://:%s@localhost:6379", redisPassword)

		dir, err := config.Dir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create config dir: %w", err)
		}

		envContent := fmt.Sprintf(
			"WORKER_TOKEN=%s\nREDIS_PASSWORD=%s\nREDIS_URL=%s\nPOSTGRES_PASSWORD=%s\nDATABASE_URL=%s\n",
			workerToken, redisPassword, redisURL, postgresPassword, postgresDSN,
		)
		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			return fmt.Errorf("failed to write .env: %w", err)
		}

		if err := os.WriteFile(composePath, embeddedCompose, 0644); err != nil {
			return fmt.Errorf("failed to write docker-compose.yml: %w", err)
		}

		cfg := config.Config{
			WorkerToken:      workerToken,
			RedisPassword:    redisPassword,
			RedisURL:         redisURL,
			PostgresPassword: postgresPassword,
			PostgresDSN:      postgresDSN,
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("braceletCI initialised.")
		fmt.Printf("  Config:        %s\n", cfgPath)
		fmt.Printf("  Env file:      %s\n", envPath)
		fmt.Printf("  Compose file:  %s\n", composePath)
		fmt.Println()
		fmt.Println("Next: run `bracelet-ci up` to start Redis, Postgres, and the DB service.")
		return nil
	},
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start Redis, Postgres, and bracelet-db-service, then run the webhook server",
	RunE: func(cmd *cobra.Command, args []string) error {
		envPath, err := config.EnvFilePath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			return fmt.Errorf("no env file found — run `bracelet-ci init` first")
		}

		composePath, err := config.ComposeFilePath()
		if err != nil {
			return err
		}

		fmt.Println("Starting Redis and Postgres...")
		compose := exec.Command("docker", "compose",
			"--env-file", envPath,
			"-f", composePath,
			"up", "-d",
		)
		compose.Stdout = os.Stdout
		compose.Stderr = os.Stderr
		if err := compose.Run(); err != nil {
			return fmt.Errorf("docker compose up failed: %w", err)
		}

		dir, _ := config.Dir()
		logFile, err := os.OpenFile(
			filepath.Join(dir, "db-service.log"),
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644,
		)
		if err != nil {
			return fmt.Errorf("failed to open db-service log: %w", err)
		}

		dbSvc := exec.Command("/home/parthsharma/bracelet-cicd/cmd/bracelet-DB-service/bracelet-DB-service", "start")
		dbSvc.Env = append(os.Environ(), "GODOTENV_PATH="+envPath)
		dbSvc.Stdout = logFile
		dbSvc.Stderr = logFile
		if err := dbSvc.Start(); err != nil {
			return fmt.Errorf("failed to start bracelet-db-service: %w", err)
		}

		pidPath := filepath.Join(dir, "db-service.pid")
		if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", dbSvc.Process.Pid)), 0644); err != nil {
			return fmt.Errorf("failed to write pid file: %w", err)
		}

		fmt.Printf("bracelet-db-service started (pid %d) — logs at %s\n", dbSvc.Process.Pid, filepath.Join(dir, "db-service.log"))
		fmt.Println("Starting webhook server on :8080...")

		runServer()
		return nil
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop bracelet-db-service and shut down Redis and Postgres",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := config.Dir()
		if err != nil {
			return err
		}

		pidPath := filepath.Join(dir, "db-service.pid")
		pidBytes, err := os.ReadFile(pidPath)
		if err == nil {
			var pid int
			if _, err := fmt.Sscanf(string(pidBytes), "%d", &pid); err == nil {
				if proc, err := os.FindProcess(pid); err == nil {
					proc.Signal(os.Interrupt)
					fmt.Printf("Stopped bracelet-db-service (pid %d)\n", pid)
				}
			}
			os.Remove(pidPath)
		}

		envPath, _ := config.EnvFilePath()
		composePath, _ := config.ComposeFilePath()
		compose := exec.Command("docker", "compose",
			"--env-file", envPath,
			"-f", composePath,
			"down",
		)
		compose.Stdout = os.Stdout
		compose.Stderr = os.Stderr
		if err := compose.Run(); err != nil {
			return fmt.Errorf("docker compose down failed: %w", err)
		}

		fmt.Println("Done.")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of braceletCI services",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := config.Dir()
		pidPath := filepath.Join(dir, "db-service.pid")
		pidBytes, err := os.ReadFile(pidPath)
		if err == nil {
			fmt.Printf("bracelet-db-service pid: %s\n", strings.TrimSpace(string(pidBytes)))
		} else {
			fmt.Println("bracelet-db-service: not running")
		}

		envPath, _ := config.EnvFilePath()
		composePath, _ := config.ComposeFilePath()
		compose := exec.Command("docker", "compose",
			"--env-file", envPath,
			"-f", composePath,
			"ps",
		)
		compose.Stdout = os.Stdout
		compose.Stderr = os.Stderr
		return compose.Run()
	},
}

func init() {
	rootCmd.AddCommand(initCmd, upCmd, downCmd, statusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}


func runServer() {
	godotenv.Load(envFilePath())

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

func envFilePath() string {
	p, err := config.EnvFilePath()
	if err != nil {
		return ""
	}
	return p
}
