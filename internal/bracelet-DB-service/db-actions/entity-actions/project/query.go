package projectactions

import (
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
)

func (pEditor *ProjectEditor) Query(operation string) (any, error) {
	switch operation {
	case "get_details":
		return pEditor.getDetails()
	case "get_by_user_id":
		return pEditor.getByUserId()
	case "get_by_repo_url":
		return pEditor.getByRepoUrl()
	default:
		return nil, fmt.Errorf("[Project Query Error] unsupported operation %q", operation)
	}
}

func (pEditor *ProjectEditor) getByUserId() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Query Error] invalid project data")
	}
	if project.UserId == nil || *project.UserId == "" {
		return nil, fmt.Errorf("[Project Query Error] user_id is required")
	}

	rows, err := pEditor.dbConn.FetchRecords(
		`SELECT id, user_id, repo_url, name
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		*project.UserId,
	)
	if err != nil {
		return nil, fmt.Errorf("[Project Query Error] %w", err)
	}
	defer rows.Close()

	projects := make([]models.Project, 0)
	for rows.Next() {
		var result models.Project
		var userId string
		var repoUrl string
		var name string
		if err := rows.Scan(
			&result.ProjectId,
			&userId,
			&repoUrl,
			&name,
		); err != nil {
			return nil, fmt.Errorf("[Project Query Error] %w", err)
		}
		result.UserId = &userId
		result.RepoUrl = &repoUrl
		result.Name = &name
		projects = append(projects, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[Project Query Error] %w", err)
	}

	return projects, nil
}

func (pEditor *ProjectEditor) getDetails() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Query Error] invalid project data")
	}
	if project.ProjectId == "" {
		return nil, fmt.Errorf("[Project Query Error] project_id is required")
	}

	rows, err := pEditor.dbConn.FetchRecords(
		`SELECT
			p.id,
			p.user_id,
			p.repo_url,
			p.name,
			u.username,
			u.email,
			COUNT(j.id)
		FROM projects p
		JOIN "user" u ON u.id = p.user_id
		LEFT JOIN jobs j ON j.project_id = p.id
		WHERE p.id = $1
		GROUP BY p.id, p.user_id, p.repo_url, p.name, u.username, u.email`,
		project.ProjectId,
	)
	if err != nil {
		return nil, fmt.Errorf("[Project Query Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[Project Query Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[Project Query Error] project not found")
	}

	var details models.ProjectDetails
	if err := rows.Scan(
		&details.ProjectId,
		&details.UserId,
		&details.RepoUrl,
		&details.Name,
		&details.OwnerUsername,
		&details.OwnerEmail,
		&details.JobCount,
	); err != nil {
		return nil, fmt.Errorf("[Project Query Error] %w", err)
	}

	return details, nil
}

func (pEditor *ProjectEditor) getByRepoUrl() (any, error) {
	project, ok := pEditor.event.EntityData.(models.Project)
	if !ok {
		return nil, fmt.Errorf("[Project Query Error] invalid project data")
	}
	if project.RepoUrl == nil || *project.RepoUrl == "" {
		return nil, fmt.Errorf("[Project Query Error] repo_url is required")
	}

	rows, err := pEditor.dbConn.FetchRecords(
		`SELECT id FROM projects WHERE repo_url = $1 LIMIT 1`,
		*project.RepoUrl,
	)
	if err != nil {
		return nil, fmt.Errorf("[Project Query Error] %w", err)
	}
	defer rows.Close()
	
	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[Project Query Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[Project Query Error] no project found for repo_url %q", *project.RepoUrl)
	}
	
	var result models.Project
	if err := rows.Scan(&result.ProjectId); err != nil {
		return nil, fmt.Errorf("[Project Query Error] %w", err)
	}

	return result, nil
}
