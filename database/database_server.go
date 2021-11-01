package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DatabaseServer struct {
	ConnectionPool *pgxpool.Pool
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
