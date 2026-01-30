package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDB struct {
	Pool      *pgxpool.Pool
	STDB      *sql.DB
	Container testcontainers.Container
	DSN       string
}

var testDB *TestDB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgis/postgis:16-3.4",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		os.Exit(1)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to get connection string: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("Failed to create pool: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}

	for i := 0; i < 30; i++ {
		err = pool.Ping(ctx)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Printf("Failed to parse config: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}
	stdb := stdlib.OpenDB(*config.ConnConfig)

	testDB = &TestDB{
		Pool:      pool,
		STDB:      stdb,
		Container: container,
		DSN:       dsn,
	}

	if err := runMigrations(stdb); err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}

	code := m.Run()

	pool.Close()
	stdb.Close()
	container.Terminate(ctx)

	os.Exit(code)
}

func runMigrations(db *sql.DB) error {
	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationPath, err := filepath.Abs("../migrations")
	if err != nil {
		return fmt.Errorf("failed to resolve migrations directory: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func getTestModels(t *testing.T) Models {
	t.Helper()
	return NewModels(testDB.Pool, testDB.STDB)
}

func cleanupTables(t *testing.T) {
	t.Helper()
	_, err := testDB.STDB.Exec("TRUNCATE TABLE inlay_proofs, inlay_chats, inlay_custom_infos, inlay_catalog_infos, inlays, projects, catalog_items, tokens, accounts, users, dealerships CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}
}

func createTestDealership(t *testing.T, models Models) *Dealership {
	t.Helper()

	dealership := &Dealership{
		Name: "Test Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "Suite 100",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}

	err := models.Dealerships.Insert(dealership)
	if err != nil {
		t.Fatalf("Failed to create test dealership: %v", err)
	}

	return dealership
}

func createTestUser(t *testing.T, models Models, dealershipID int) *User {
	t.Helper()

	user := &User{
		Name:         "Test User",
		Email:        fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealershipID,
		Role:         UserRoles.User,
	}

	err := models.Users.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func createTestProject(t *testing.T, models Models, dealershipID int) *Project {
	t.Helper()

	project := &Project{
		Name:         "Test Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealershipID,
	}

	err := models.Projects.Insert(project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return project
}

func createTestCatalogItem(t *testing.T, models Models) int {
	t.Helper()

	var id int
	err := testDB.STDB.QueryRow("INSERT INTO catalog_items DEFAULT VALUES RETURNING id").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to create test catalog item: %v", err)
	}

	return id
}
