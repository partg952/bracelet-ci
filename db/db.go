package DB

import (
	"context"
	"fmt"
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
		fmt.Errorf("[Database Connectivity Error] Failed to create db connection pool : %v\n", err)
		return DBInstance{}, err
	}

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		fmt.Errorf("[Database Connectivity Error] Failed to connect to the db : %v\n", err)
		return DBInstance{}, err
	}

	return DBInstance{
		Conn: db,
	}, nil

}

func (db *DBInstance) FetchRecords(query string, args ...any) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.Conn.Query(ctx, query, args...)
	if err != nil {
		fmt.Errorf("[Database Query Error] An error occurred while querying the DB : %v\n", err)
		return rows, err
	}
	return rows, nil

}



func (db *DBInstance) ExecuteQuery(query string, args ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Conn.Exec(ctx, query, args...)
	if err != nil {
		fmt.Errorf("[Database Query Error] An error occurred while executing the query : %v\n", err)
		return err
	}

	return nil
}
