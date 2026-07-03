package useractions

import (
	"bracelet-cicd/internal/bracelet-DB-service/db"
	dbactions "bracelet-cicd/internal/bracelet-DB-service/db-actions"
	"bracelet-cicd/internal/bracelet-DB-service/models"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserEditor struct {
	event  dbactions.Event
	dbConn *db.DBInstance
}

func NewUserEditor(event dbactions.Event, db *db.DBInstance) (UserEditor, error) {
	if event.EntityName == "" || event.Method == "" {
		return UserEditor{}, fmt.Errorf("Empty method or entity")
	}

	return UserEditor{
		event,
		db,
	}, nil
}

func (uEditor *UserEditor) Create() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Create Error] invalid user data")
	}
	if user.Password == nil || *user.Password == "" {
		return nil, fmt.Errorf("[User Create Error] password is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(*user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("[User Create Error] failed to hash password: %w", err)
	}

	if err := uEditor.dbConn.ExecuteQuery(
		`INSERT INTO "user"(id, username, email, password) VALUES($1, $2, $3, $4)`,
		user.UserId,
		user.Username,
		user.Email,
		string(hashedPassword),
	); err != nil {
		return nil, err
	}

	return nil, nil
}

func (uEditor *UserEditor) Update() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Update Error] invalid user data")
	}
	if user.UserId == "" {
		return nil, fmt.Errorf("[User Update Error] user_id is required")
	}

	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if user.Username != nil {
		args = append(args, *user.Username)
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", len(args)))
	}
	if user.Email != nil {
		args = append(args, *user.Email)
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", len(args)))
	}
	if user.Password != nil {
		args = append(args, *user.Password)
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("[User Update Error] no fields provided to update")
	}

	args = append(args, user.UserId)
	query := fmt.Sprintf(
		`UPDATE "user" SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "),
		len(args),
	)

	if err := uEditor.dbConn.ExecuteQuery(query, args...); err != nil {
		return nil, err
	}

	return nil, nil
}

func (uEditor *UserEditor) Read() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Read Error] invalid user data")
	}
	if user.UserId == "" {
		return nil, fmt.Errorf("[User Read Error] user_id is required")
	}

	rows, err := uEditor.dbConn.FetchRecords(
		`SELECT username, email FROM "user" WHERE id = $1`,
		user.UserId,
	)
	if err != nil {
		return nil, fmt.Errorf("[User Read Error] %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("[User Read Error] %w", rows.Err())
		}
		return nil, fmt.Errorf("[User Read Error] user not found")
	}

	var username string
	var email string
	if err := rows.Scan(&username, &email); err != nil {
		return nil, fmt.Errorf("[User Read Error] %w", err)
	}

	return models.User{
		UserId:   user.UserId,
		Username: &username,
		Email:    &email,
	}, nil
}

func (uEditor *UserEditor) Delete() (any, error) {
	user, ok := uEditor.event.EntityData.(models.User)
	if !ok {
		return nil, fmt.Errorf("[User Delete Error] invalid user data")
	}
	if user.UserId == "" {
		return nil, fmt.Errorf("[User Delete Error] No user id provided")
	}
	if err := uEditor.dbConn.ExecuteQuery("DELETE user WHERE id=$1", user.UserId); err != nil {
		return nil, err
	}
	return nil, nil
}
