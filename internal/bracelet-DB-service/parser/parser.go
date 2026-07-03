package parser

import (
	dbactions "bracelet-cicd/internal/bracelet-DB-service/db-actions"
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type EntityData interface {
	models.Job | models.Project | models.User
}

type RawEvent struct {
	Method     string          `json:"method"`
	EntityName string          `json:"entity_name"`
	Operation  string          `json:"operation,omitempty"`
	EntityData json.RawMessage `json:"entity_data"`
}

type Parser func(RawEvent) (dbactions.Event, error)

var registry = map[string]Parser{
	"job":     parseEvent[models.Job],
	"project": parseEvent[models.Project],
	"user":    parseEvent[models.User],
}

func parseEvent[T EntityData](payload RawEvent) (dbactions.Event, error) {
	var entityData T
	if err := json.Unmarshal(payload.EntityData, &entityData); err != nil {
		return dbactions.Event{}, err
	}

	return dbactions.Event{
		Method:     payload.Method,
		EntityName: payload.EntityName,
		Operation:  payload.Operation,
		EntityData: entityData,
	}, nil
}

func ParseEvent(payload RawEvent) (dbactions.Event, error) {
	entityName := strings.ToLower(strings.TrimSpace(payload.EntityName))
	parse, ok := registry[entityName]
	if !ok {
		return dbactions.Event{}, fmt.Errorf("unsupported entity name %q", payload.EntityName)
	}

	method := strings.ToLower(strings.TrimSpace(payload.Method))
	switch method {
	case "create", "read", "update", "delete":
		if !hasEntityData(payload.EntityData) {
			return dbactions.Event{}, fmt.Errorf("entity_data is required for %s events", method)
		}
	case "query":
		payload.Operation = strings.ToLower(strings.TrimSpace(payload.Operation))
		if payload.Operation == "" {
			return dbactions.Event{}, fmt.Errorf("operation is required for query events")
		}
		if !hasEntityData(payload.EntityData) {
			return dbactions.Event{}, fmt.Errorf("entity_data is required for query events")
		}
	default:
		return dbactions.Event{}, fmt.Errorf("unsupported method %q", payload.Method)
	}

	payload.Method = method
	payload.EntityName = entityName
	return parse(payload)
}

func hasEntityData(data json.RawMessage) bool {
	data = bytes.TrimSpace(data)
	return len(data) > 0 && !bytes.Equal(data, []byte("null"))
}
