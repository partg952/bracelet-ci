package main

import (
	"context"
	"log"
	"os"

	"bracelet-cicd/internal/bracelet-worker/worker"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func init() {
	godotenv.Load("../../.env")
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	dbServiceURL := os.Getenv("DB_SERVICE_URL")
	if dbServiceURL == "" {
		dbServiceURL = "http://localhost:8081"
	}
	var redisOptions *redis.Options
	var err error
	if redisURL != "" {
		redisOptions, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatal("invalid REDIS_URL:", err)
		}
	}
	client := redis.NewClient(redisOptions)
	log.Printf("Redis Client created\n")
	worker := worker.New(client, dbServiceURL, "job_queue", 0)
	worker.Start(context.Background())
}
