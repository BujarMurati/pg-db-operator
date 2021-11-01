package database

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	sanitizedDatabaseName := strings.Trim(pgx.Identifier{databaseName}.Sanitize(), "\"") // for reference https://github.com/jackc/pgx/issues/498
	query := fmt.Sprintf("CREATE DATABASE %v;", sanitizedDatabaseName)
	_, err = d.ConnectionPool.Exec(context.Background(), query)
	return err
}

func (d DatabaseServer) CheckUserExists(userName string) (exists bool, err error) {
	query := "SELECT COUNT(usename) FROM pg_user WHERE usename = $1;"
	var count int
	err = d.ConnectionPool.QueryRow(context.Background(), query, userName).Scan(&count)
	return count > 0, err
}

func (d DatabaseServer) CreateUserOrUpdatePassword(userName string, password string) (err error) {
	var exists bool
	exists, err = d.CheckUserExists(userName)
	verb := "CREATE"
	if err != nil {
		return err
	}
	if exists {
		verb = "ALTER"
	}
	sanitizedUserName := strings.Trim(pgx.Identifier{userName}.Sanitize(), "\"")
	sanitizedPassword := "'" + strings.Trim(pgx.Identifier{password}.Sanitize(), "\"") + "'"
	query := fmt.Sprintf("%v USER %v WITH ENCRYPTED PASSWORD %v;", verb, sanitizedUserName, sanitizedPassword)
	_, err = d.ConnectionPool.Exec(context.Background(), query)
	return err
}

func (d DatabaseServer) CheckUserHasAllPrivileges(userName string, databaseName string) (hasPrivileges bool, err error) {
	// https://www.postgresql.org/docs/current/functions-info.html#FUNCTIONS-INFO-ACCESS-TABLE
	// Here we test for CREATE permissions, there does not seem to be an elegant way to
	// verify that "ALL" privileges on a database were granted
	query := "SELECT has_database_privilege($1, $2, 'CREATE');"
	err = d.ConnectionPool.QueryRow(context.Background(), query, userName, databaseName).Scan(&hasPrivileges)
	return hasPrivileges, err
}

func (d DatabaseServer) EnsureUserHasAllPrivileges(userName string, databaseName string) (err error) {
	var exists bool
	exists, err = d.CheckUserHasAllPrivileges(userName, databaseName)
	if err != nil || exists {
		return err
	}
	sanitizedUserName := strings.Trim(pgx.Identifier{userName}.Sanitize(), "\"")
	santizedDatabaseName := strings.Trim(pgx.Identifier{databaseName}.Sanitize(), "\"")
	query := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %v TO %v;", santizedDatabaseName, sanitizedUserName)
	_, err = d.ConnectionPool.Exec(context.Background(), query)
	return err
}

func (d DatabaseServer) ReconcileDatabaseState(userName string, databaseName string, password string) (err error) {
	err = d.CreateDatabaseIfNotExists(databaseName)
	if err != nil {
		return err
	}
	err = d.CreateUserOrUpdatePassword(userName, password)
	if err != nil {
		return err
	}
	err = d.EnsureUserHasAllPrivileges(userName, databaseName)
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
