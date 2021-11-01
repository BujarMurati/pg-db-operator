package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DatabaseServer struct {
	ConnectionPool *pgxpool.Pool
}

func (d DatabaseServer) CheckDatabaseExists(databaseName string) (exists bool, err error) {
	query := "SELECT COUNT(datname) FROM pg_database WHERE datname = $1;"
	var count int
	err = d.ConnectionPool.QueryRow(context.Background(), query, databaseName).Scan(&count)
	return count > 0, err
}

func (d DatabaseServer) CreateDatabaseIfNotExists(databaseName string) (err error) {
	var exists bool
	exists, err = d.CheckDatabaseExists(databaseName)
	if err != nil || exists {
		return err
	}
	sanitizedDatabaseName := pgx.Identifier{databaseName}.Sanitize() // for reference https://github.com/jackc/pgx/issues/498
	query := fmt.Sprintf("CREATE DATABASE %s;", sanitizedDatabaseName)
	_, err = d.ConnectionPool.Exec(context.Background(), query)
	return err
}

//Accepts libpq environment variables https://www.postgresql.org/docs/9.4/libpq-envars.html
func NewDatabaseServerFromEnvironment() (db DatabaseServer, err error) {
	config, err := pgxpool.ParseConfig(os.Getenv("PG_DB_OPERATOR_CONNECTION_STRING"))
	if err != nil {
		return DatabaseServer{}, err
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return DatabaseServer{}, err
	}
	return DatabaseServer{ConnectionPool: pool}, nil
}
