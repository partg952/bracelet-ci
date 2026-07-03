package projectactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/db"
	dbactions "bracelet-cicd/internal/bracelet-DB-service/db-actions"
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"strings"
)

type ProjectEditor struct {
	event  dbactions.Event
	dbConn *db.DBInstance
}

func NewProjectEditor(event dbactions.Event, db *db.DBInstance) (ProjectEditor, error) {
	if event.EntityName == "" || event.Method == "" {
		return ProjectEditor{}, fmt.Errorf("Empty method or entity")
	}

	return ProjectEditor{
		event:  event,
		dbConn: db,
	}, nil
}

func (pEditor *ProjectEditor) Create() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Create Error] invalid project data")
	}
	if project.ProjectId == "" ||
		project.UserId == nil || *project.UserId == "" ||
		project.RepoUrl == nil || *project.RepoUrl == "" ||
		project.Name == nil || *project.Name == "" {
		return nil, fmt.Errorf("[Project Create Error] project_id, user_id, repo_url, and name are required")
	}

	if err := pEditor.dbConn.ExecuteQuery(
		`INSERT INTO projects(id, user_id, repo_url, name) VALUES($1, $2, $3, $4)`,
		project.ProjectId,
		*project.UserId,
		*project.RepoUrl,
		*project.Name,
	); err != nil {
		return nil, err
	}

	return nil, nil
}

func (pEditor *ProjectEditor) Read() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Read Error] invalid project data")
	}
	if project.ProjectId == "" {
		return nil, fmt.Errorf("[Project Read Error] project_id is required")
	}

	rows, err := pEditor.dbConn.FetchRecords(
		`SELECT user_id, repo_url, name FROM projects WHERE id = $1`,
		project.ProjectId,
	)
	if err != nil {
		return nil, fmt.Errorf("[Project Read Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[Project Read Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[Project Read Error] project not found")
	}

	var userId string
	var repoUrl string
	var name string
	if err := rows.Scan(&userId, &repoUrl, &name); err != nil {
		return nil, fmt.Errorf("[Project Read Error] %w", err)
	}

	return models.Project{
		ProjectId: project.ProjectId,
		UserId:    &userId,
		RepoUrl:   &repoUrl,
		Name:      &name,
	}, nil
}

func (pEditor *ProjectEditor) Update() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Update Error] invalid project data")
	}
	if project.ProjectId == "" {
		return nil, fmt.Errorf("[Project Update Error] project_id is required")
	}

	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if project.UserId != nil {
		args = append(args, *project.UserId)
		setClauses = append(setClauses, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if project.RepoUrl != nil {
		args = append(args, *project.RepoUrl)
		setClauses = append(setClauses, fmt.Sprintf("repo_url = $%d", len(args)))
	}
	if project.Name != nil {
		args = append(args, *project.Name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("[Project Update Error] no fields provided to update")
	}

	args = append(args, project.ProjectId)
	query := fmt.Sprintf(
		`UPDATE projects SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "),
		len(args),
	)

	if err := pEditor.dbConn.ExecuteQuery(query, args...); err != nil {
		return nil, err
	}

	return nil, nil
}

func (pEditor *ProjectEditor) Delete() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Delete Error] invalid project data")
	}
	if project.ProjectId == "" {
		return nil, fmt.Errorf("[Project Delete Error] project_id is required")
	}

	if err := pEditor.dbConn.ExecuteQuery(`DELETE FROM projects WHERE id = $1`, project.ProjectId); err != nil {
		return nil, err
	}

	return nil, nil
}
