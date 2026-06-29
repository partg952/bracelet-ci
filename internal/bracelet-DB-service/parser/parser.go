package parser

import (
	"encoding/json"
	"fmt"
	"log"
)

// parsed struct from json from request body
// pure entity data structs
type Job struct {
	JobId      string  `json:"job_id"`
	ProjectId  string  `json:"project_id"`
	CommitSHA  string  `json:"commit_sha"`
	FinishedAt *string `json:"finished_at"`
}

type Project struct {
	ProjectId string `json:"project_id"`
	RepoUrl   string `json:"repo_url"`
	Name      string `json:"name"`
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// all pure data structs satisfy this interface
type EntityDataTypes interface {
	Job | Project | User
}

type RawEvent struct {
	Method     string          `json:"method"`
	EntityName string          `json:"entity_name"`
	EntityData json.RawMessage `json:"entity_data"`
}

type Event[T EntityDataTypes] struct {
	Method     string
	EntityName string
	EntityData T
}

type Parser func(RawEvent) (any, error)

var registry = map[string]Parser{
	"job":     parseEvent[Job],
	"project": parseEvent[Project],
	"user":    parseEvent[User],
}

func parseEvent[T EntityDataTypes](payload RawEvent) (any, error) {
	var entityData T
	if err := json.Unmarshal(payload.EntityData, &entityData); err != nil {
		return nil, err
	}

	log.Printf("Entity Data: %v", entityData)

	return Event[T]{
		Method:     payload.Method,
		EntityName: payload.EntityName,
		EntityData: entityData,
	}, nil
}

func ParseEvent(payload RawEvent) (any, error) {
	parse, ok := registry[payload.EntityName]
	if !ok {
		return nil, fmt.Errorf("unsupported entity name %q", payload.EntityName)
	}

	return parse(payload)
}
