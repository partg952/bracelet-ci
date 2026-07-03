package main

import (
	dbpkg "bracelet-cicd/internal/bracelet-DB-service/db"
	_ "bracelet-cicd/internal/bracelet-DB-service/db-actions/entity-actions"
	"bracelet-cicd/internal/bracelet-DB-service/parser"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../../.env")

	dbInstance, err := dbpkg.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbInstance.Conn.Close()

	r := gin.Default()
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

		result, err := parsedEvent.Execute(&dbInstance)
		if err != nil {
			log.Printf("[Event Execution error] %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	})
	r.Run(":8081")
}
