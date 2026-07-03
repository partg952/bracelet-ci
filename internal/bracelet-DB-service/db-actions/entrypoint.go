package dbactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/db"
	"fmt"
	"strings"
)

type Event struct {
	Method     string
	EntityName string
	Operation  string
	EntityData any
}

type EntityMethods interface {
	Create() (any, error)
	Update() (any, error)
	Delete() (any, error)
	Read() (any, error)
}

type EntityQueries interface {
	Query(operation string) (any, error)
}

type EditorFactory func(Event, *db.DBInstance) (EntityMethods, error)

var editorRegistry = map[string]EditorFactory{}

func RegisterEditor(entityName string, factory EditorFactory) {
	editorRegistry[strings.ToLower(entityName)] = factory
}

func (event *Event) Execute(database *db.DBInstance) (any, error) {
	newEditor, ok := editorRegistry[strings.ToLower(event.EntityName)]
	if !ok {
		return nil, fmt.Errorf("unsupported entity name %q", event.EntityName)
	}

	editor, err := newEditor(*event, database)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(event.Method) {
	case "create":
		return editor.Create()
	case "read":
		return editor.Read()
	case "update":
		return editor.Update()
	case "delete":
		return editor.Delete()
	case "query":
		queryEditor, ok := editor.(EntityQueries)
		if !ok {
			return nil, fmt.Errorf("%s does not support query operations", event.EntityName)
		}
		return queryEditor.Query(strings.ToLower(event.Operation))
	default:
		return nil, fmt.Errorf("unsupported method %q", event.Method)
	}
}
