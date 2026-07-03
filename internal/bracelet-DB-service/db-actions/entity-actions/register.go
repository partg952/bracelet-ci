package entityactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/db"
	dbactions "bracelet-cicd/internal/bracelet-DB-service/db-actions"
	jobactions "bracelet-cicd/internal/bracelet-DB-service/db-actions/entity-actions/job"
	projectactions "bracelet-cicd/internal/bracelet-DB-service/db-actions/entity-actions/project"
	useractions "bracelet-cicd/internal/bracelet-DB-service/db-actions/entity-actions/user"
)

func init() {
	dbactions.RegisterEditor("project", newProjectEditor)
	dbactions.RegisterEditor("job", newJobEditor)
	dbactions.RegisterEditor("user", newUserEditor)
}

func newProjectEditor(event dbactions.Event, database *db.DBInstance) (dbactions.EntityMethods, error) {
	editor, err := projectactions.NewProjectEditor(event, database)
	if err != nil {
		return nil, err
	}
	return &editor, nil
}

func newJobEditor(event dbactions.Event, database *db.DBInstance) (dbactions.EntityMethods, error) {
	editor, err := jobactions.NewJobEditor(event, database)
	if err != nil {
		return nil, err
	}
	return &editor, nil
}

func newUserEditor(event dbactions.Event, database *db.DBInstance) (dbactions.EntityMethods, error) {
	editor, err := useractions.NewUserEditor(event, database)
	if err != nil {
		return nil, err
	}
	return &editor, nil
}
