package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBInstance struct {
	Conn *pgxpool.Pool
}

func New(DBUrl string) (DBInstance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, DBUrl)
	if err != nil {
		log.Printf("[Database connectivity error] Couldn't create a connection pool : %v", err)
		return DBInstance{}, err
	}
	log.Printf("Connection Pool created\n")
	err = db.Ping(ctx)
	if err != nil {
		log.Printf("[Database connectivity error] Couldn't connect to the database : %v", err)
		return DBInstance{}, err
	}
	log.Printf("Connected to the database")
	return DBInstance{
		Conn: db,
	}, nil
}

func (db *DBInstance) FetchResults(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	value, err := db.Conn.Query(ctx, query, args...)
	if err != nil {
		log.Printf("[Database Query error] An error occured while executing the query : %v", err)
		return value, err
	}
	return value, nil
}

func (db *DBInstance) ExecuteQuery(ctx context.Context, query string, args ...any) error {
	_, err := db.Conn.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("[Query Execution error] An error occurred while executing query : %v", err)
		return err
	}
	return nil
}
