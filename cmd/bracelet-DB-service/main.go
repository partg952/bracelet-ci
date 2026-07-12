package main

import (
	dbpkg "bracelet-cicd/internal/bracelet-DB-service/db"
	_ "bracelet-cicd/internal/bracelet-DB-service/db-actions/entity-actions"
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"bracelet-cicd/internal/bracelet-DB-service/parser"
	"bracelet-cicd/internal/bracelet-DB-service/util"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func projectIdForJobEvent(database *dbpkg.DBInstance, job models.Job) string {
	if job.ProjectId != nil && *job.ProjectId != "" {
		return *job.ProjectId
	}
	if job.JobId == "" {
		return ""
	}

	rows, err := database.FetchRecords(`SELECT project_id FROM jobs WHERE id = $1`, job.JobId)
	if err != nil {
		log.Printf("[SSE Publish error] failed to fetch project_id for job %s: %v", job.JobId, err)
		return ""
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			log.Printf("[SSE Publish error] failed to read project_id for job %s: %v", job.JobId, err)
		}
		return ""
	}

	var projectId *string
	if err := rows.Scan(&projectId); err != nil {
		log.Printf("[SSE Publish error] failed to scan project_id for job %s: %v", job.JobId, err)
		return ""
	}
	if projectId == nil {
		return ""
	}

	return *projectId
}

func main() {
	dbInstance, err := dbpkg.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbInstance.Conn.Close()
	broker := util.NewHub()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:4173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Type"},
		AllowCredentials: false,
	}))
	r.GET("/stream", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		projectId := c.Query("project_id")
		if projectId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
			return
		}

		channel, unsubscribe := broker.Subscribe(projectId)
		defer unsubscribe()

		c.Stream(func(w io.Writer) bool {
			select {
			case <-c.Request.Context().Done():
				return false
			case t := <-channel:
				c.SSEvent("database-event", t)
				return true
			}
		})
	})
	r.POST("/event", func(ctx *gin.Context) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}
		var rawEventInstance parser.RawEvent
		err = json.Unmarshal(body, &rawEventInstance)
		if err != nil {
			log.Printf("Error occurred while unmarshalling for raw event : %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		parsedEvent, err := parser.ParseEvent(rawEventInstance)
		if err != nil {
			log.Printf("[Event Parsing error] An error occurred while parsing the event : %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		projectId := ""
		if parsedEvent.EntityName == "job" {
			job, ok := parsedEvent.EntityData.(models.Job)
			if !ok {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job event data"})
				return
			}
			// For create events the job doesn't exist in the DB yet, so read
			// project_id directly from the payload. For update/delete fall back
			// to a DB lookup (job may only carry job_id).
			if parsedEvent.Method == "create" && job.ProjectId != nil && *job.ProjectId != "" {
				projectId = *job.ProjectId
			} else {
				projectId = projectIdForJobEvent(&dbInstance, job)
			}
		}

		result, err := parsedEvent.Execute(&dbInstance)
		if err != nil {
			log.Printf("[Event Execution error] %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if projectId != "" {
			broker.Publish(projectId, gin.H{
				"method":      parsedEvent.Method,
				"entity_name": parsedEvent.EntityName,
				"operation":   parsedEvent.Operation,
				"entity_data": parsedEvent.EntityData,
				"result":      result,
			})
		}

		ctx.JSON(http.StatusOK, result)
	})
	r.Run(":8081")
}
