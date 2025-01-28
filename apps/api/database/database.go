package database

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Db *pgxpool.Pool

func Connect() {
	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connected to db!")

	Db = conn
}

func Migrate(direction string) {
	m, err := migrate.New("file://database/migrations", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to migrate the database: %v\n", err)
		os.Exit(1)
	}

	if direction == "up" {
		m.Up()
	}

	if direction == "down" {
		m.Down()
	}

	fmt.Println("Database migrated!")
}
