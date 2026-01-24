package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func NewPool(dsn string) (*pgxpool.Pool, *sql.DB, error) {
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, nil, err
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, err
	}

	sqlDB := stdlib.OpenDB(*config.ConnConfig)

	return db, sqlDB, nil
}
