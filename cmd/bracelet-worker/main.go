package main

import (
	"context"
	"log"
	"os"

	"bracelet-cicd/internal/bracelet-worker/db"
	"bracelet-cicd/internal/bracelet-worker/worker"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)


func init() {
	godotenv.Load()
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	dbInstance,_ := db.New(os.Getenv("DATABASE_URL"))
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
	worker := worker.New(client , dbInstance, "job_queue", 0)
	worker.Start(context.Background())
}
