package useractions

import (
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (uEditor *UserEditor) Query(operation string) (any, error) {
	switch operation {
	case "get_details":
		return uEditor.getDetails()
	case "get_by_email":
		return uEditor.getByEmail()
	case "sign_in":
		return uEditor.signIn()
	case "exists_by_email":
		return uEditor.existsByEmail()
	default:
		return nil, fmt.Errorf("[User Query Error] unsupported operation %q", operation)
	}
}

func (uEditor *UserEditor) existsByEmail() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Query Error] invalid user data")
	}
	if user.Email == nil || strings.TrimSpace(*user.Email) == "" {
		return nil, fmt.Errorf("[User Query Error] email is required")
	}

	rows, err := uEditor.dbConn.FetchRecords(
		`SELECT EXISTS(
			SELECT 1
			FROM "user"
			WHERE email = $1
		)`,
		strings.TrimSpace(*user.Email),
	)
	if err != nil {
		return nil, fmt.Errorf("[User Query Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[User Query Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[User Query Error] failed to check user")
	}

	var result models.UserExists
	if err := rows.Scan(&result.Exists); err != nil {
		return nil, fmt.Errorf("[User Query Error] %w", err)
	}

	return result, nil
}

func (uEditor *UserEditor) signIn() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Sign In Error] invalid user data")
	}
	if user.Email == nil || strings.TrimSpace(*user.Email) == "" ||
		user.Password == nil || *user.Password == "" {
		return nil, fmt.Errorf("[User Sign In Error] email and password are required")
	}

	rows, err := uEditor.dbConn.FetchRecords(
		`SELECT id, username, email, password
		FROM "user"
		WHERE email = $1`,
		strings.TrimSpace(*user.Email),
	)
	if err != nil {
		return nil, fmt.Errorf("[User Sign In Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[User Sign In Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[User Sign In Error] invalid email or password")
	}

	var result models.User
	var username string
	var email string
	var passwordHash string
	if err := rows.Scan(
		&result.UserId,
		&username,
		&email,
		&passwordHash,
	); err != nil {
		return nil, fmt.Errorf("[User Sign In Error] %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(*user.Password),
	); err != nil {
		return nil, fmt.Errorf("[User Sign In Error] invalid email or password")
	}

	result.Username = &username
	result.Email = &email
	return result, nil
}

func (uEditor *UserEditor) getDetails() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Query Error] invalid user data")
	}
	if user.UserId == "" {
		return nil, fmt.Errorf("[User Query Error] user_id is required")
	}

	rows, err := uEditor.dbConn.FetchRecords(
		`SELECT
			u.id,
			u.username,
			u.email,
			COUNT(DISTINCT p.id),
			COUNT(j.id)
		FROM "user" u
		LEFT JOIN projects p ON p.user_id = u.id
		LEFT JOIN jobs j ON j.project_id = p.id
		WHERE u.id = $1
		GROUP BY u.id, u.username, u.email`,
		user.UserId,
	)
	if err != nil {
		return nil, fmt.Errorf("[User Query Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[User Query Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[User Query Error] user not found")
	}

	var details models.UserDetails
	if err := rows.Scan(
		&details.UserId,
		&details.Username,
		&details.Email,
		&details.ProjectCount,
		&details.JobCount,
	); err != nil {
		return nil, fmt.Errorf("[User Query Error] %w", err)
	}

	return details, nil
}

func (uEditor *UserEditor) getByEmail() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Query Error] invalid user data")
	}
	if user.Email == nil || strings.TrimSpace(*user.Email) == "" {
		return nil, fmt.Errorf("[User Query Error] email is required")
	}

	rows, err := uEditor.dbConn.FetchRecords(
		`SELECT id, username, email
		FROM "user"
		WHERE email = $1`,
		*user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("[User Query Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[User Query Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[User Query Error] user not found")
	}

	var result models.User
	var username string
	var email string
	if err := rows.Scan(&result.UserId, &username, &email); err != nil {
		return nil, fmt.Errorf("[User Query Error] %w", err)
	}
	result.Username = &username
	result.Email = &email

	return result, nil
}
