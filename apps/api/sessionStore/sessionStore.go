// Originally from https://github.com/gofiber/storage/blob/main/postgres/postgres.go
// https://docs.gofiber.io/storage/postgres_v3.x.x/postgres/
package sessionStore

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config defines the config for storage.
type Config struct {
	// DB pgxpool.Pool object will override connection uri and other connection fields
	//
	// Optional. Default is nil
	DB *pgxpool.Pool

	// Connection string to use for DB. Will override all other authentication values if used
	//
	// Optional. Default is ""
	ConnectionURI string

	// Host name where the DB is hosted
	//
	// Optional. Default is "127.0.0.1"
	Host string

	// Port where the DB is listening on
	//
	// Optional. Default is 5432
	Port int

	// Server username
	//
	// Optional. Default is ""
	Username string

	// Server password
	//
	// Optional. Default is ""
	Password string

	// Database name
	//
	// Optional. Default is "fiber"
	Database string

	// Table name
	//
	// Optional. Default is "fiber_storage"
	Table string

	// The SSL mode for the connection
	//
	// Optional. Default is "disable"
	SSLMode string

	// Reset clears any existing keys in existing Table
	//
	// Optional. Default is false
	Reset bool

	// Time before deleting expired keys
	//
	// Optional. Default is 10 * time.Second
	GCInterval time.Duration
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	ConnectionURI: "",
	Host:          "127.0.0.1",
	Port:          5432,
	Database:      "fiber",
	Table:         "fiber_storage",
	SSLMode:       "disable",
	Reset:         false,
	GCInterval:    10 * time.Second,
}

func (c *Config) getDSN() string {
	// Just return ConnectionURI if it's already exists
	if c.ConnectionURI != "" {
		return c.ConnectionURI
	}

	// Generate DSN
	dsn := "postgresql://"
	if c.Username != "" {
		dsn += url.QueryEscape(c.Username)
	}
	if c.Password != "" {
		dsn += ":" + url.QueryEscape(c.Password)
	}
	if c.Username != "" || c.Password != "" {
		dsn += "@"
	}

	// unix socket host path
	if strings.HasPrefix(c.Host, "/") {
		dsn += fmt.Sprintf("%s:%d", c.Host, c.Port)
	} else {
		dsn += fmt.Sprintf("%s:%d", url.QueryEscape(c.Host), c.Port)
	}

	dsn += fmt.Sprintf("/%s?sslmode=%s",
		url.QueryEscape(c.Database),
		c.SSLMode)

	return dsn
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}
	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Host == "" {
		cfg.Host = ConfigDefault.Host
	}
	if cfg.Port <= 0 {
		cfg.Port = ConfigDefault.Port
	}
	if cfg.Database == "" {
		cfg.Database = ConfigDefault.Database
	}
	if cfg.Table == "" {
		cfg.Table = ConfigDefault.Table
	}
	if cfg.Table == "" {
		cfg.Table = ConfigDefault.Table
	}
	if int(cfg.GCInterval.Seconds()) <= 0 {
		cfg.GCInterval = ConfigDefault.GCInterval
	}
	return cfg
}

// Storage interface that is implemented by storage providers
type Storage struct {
	db         *pgxpool.Pool
	gcInterval time.Duration
	done       chan struct{}

	sqlSelect string
	sqlInsert string
	sqlDelete string
	sqlReset  string
	sqlGC     string
}

var (
	checkSchemaMsg = "The `data` row has an incorrect data type. " +
		"It should be BYTEA but is instead %s. This will cause encoding-related panics if the DB is not migrated (see https://github.com/gofiber/storage/blob/main/MIGRATE.md)."
	dropQuery = `DROP TABLE IF EXISTS %s;`
	initQuery = []string{
		`CREATE TABLE IF NOT EXISTS %s (
			id  VARCHAR(64) PRIMARY KEY NOT NULL DEFAULT '',
			data  BYTEA NOT NULL,
			expires  BIGINT NOT NULL DEFAULT '0'
		);`,
		`CREATE INDEX IF NOT EXISTS expires ON %s (expires);`,
	}
	checkSchemaQuery = `SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS
		WHERE table_name = '%s' AND COLUMN_NAME = 'data';`
)

// New creates a new storage
func New(config ...Config) *Storage {
	// Set default config
	cfg := configDefault(config...)

	// Select db connection
	var err error
	db := cfg.DB
	if db == nil {
		db, err = pgxpool.New(context.Background(), cfg.getDSN())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		}
	}

	// Ping database
	if err := db.Ping(context.Background()); err != nil {
		panic(err)
	}

	// Drop table if set to true
	if cfg.Reset {
		if _, err := db.Exec(context.Background(), fmt.Sprintf(dropQuery, cfg.Table)); err != nil {
			db.Close()
			panic(err)
		}
	}

	// Init database queries
	for _, query := range initQuery {
		if _, err := db.Exec(context.Background(), fmt.Sprintf(query, cfg.Table)); err != nil {
			db.Close()
			panic(err)
		}
	}

	// Create storage
	store := &Storage{
		db:         db,
		gcInterval: cfg.GCInterval,
		done:       make(chan struct{}),
		sqlSelect:  fmt.Sprintf(`SELECT data, expires FROM %s WHERE id=$1;`, cfg.Table),
		sqlInsert:  fmt.Sprintf("INSERT INTO %s (id, data, expires) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET data = $4, expires = $5", cfg.Table),
		sqlDelete:  fmt.Sprintf("DELETE FROM %s WHERE id=$1", cfg.Table),
		sqlReset:   fmt.Sprintf("TRUNCATE TABLE %s;", cfg.Table),
		sqlGC:      fmt.Sprintf("DELETE FROM %s WHERE expires <= $1 AND expires != 0", cfg.Table),
	}

	store.checkSchema(cfg.Table)

	// Start garbage collector
	go store.gcTicker()

	return store
}

// Get value by key
func (s *Storage) Get(key string) ([]byte, error) {
	if len(key) <= 0 {
		return nil, nil
	}
	row := s.db.QueryRow(context.Background(), s.sqlSelect, key)
	// Add db response to data
	var (
		data []byte
		exp  int64 = 0
	)
	if err := row.Scan(&data, &exp); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// If the expiration time has already passed, then return nil
	if exp != 0 && exp <= time.Now().Unix() {
		return nil, nil
	}

	return data, nil
}

// Set key with value
func (s *Storage) Set(key string, val []byte, exp time.Duration) error {
	// Ain't Nobody Got Time For That
	if len(key) <= 0 || len(val) <= 0 {
		return nil
	}
	var expSeconds int64
	if exp != 0 {
		expSeconds = time.Now().Add(exp).Unix()
	}
	_, err := s.db.Exec(context.Background(), s.sqlInsert, key, val, expSeconds, val, expSeconds)
	return err
}

// Delete entry by key
func (s *Storage) Delete(key string) error {
	// Ain't Nobody Got Time For That
	if len(key) <= 0 {
		return nil
	}
	_, err := s.db.Exec(context.Background(), s.sqlDelete, key)
	return err
}

// Reset all entries, including unexpired
func (s *Storage) Reset() error {
	_, err := s.db.Exec(context.Background(), s.sqlReset)
	return err
}

// Close the database
func (s *Storage) Close() error {
	s.done <- struct{}{}
	s.db.Stat()
	s.db.Close()
	return nil
}

// Return database client
func (s *Storage) Conn() *pgxpool.Pool {
	return s.db
}

// gcTicker starts the gc ticker
func (s *Storage) gcTicker() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case t := <-ticker.C:
			s.gc(t)
		}
	}
}

// gc deletes all expired entries
func (s *Storage) gc(t time.Time) {
	_, _ = s.db.Exec(context.Background(), s.sqlGC, t.Unix())
}

func (s *Storage) checkSchema(tableName string) {
	var data []byte

	row := s.db.QueryRow(context.Background(), fmt.Sprintf(checkSchemaQuery, tableName))
	if err := row.Scan(&data); err != nil {
		panic(err)
	}

	if strings.ToLower(string(data)) != "bytea" {
		fmt.Printf(checkSchemaMsg, string(data))
	}
}
