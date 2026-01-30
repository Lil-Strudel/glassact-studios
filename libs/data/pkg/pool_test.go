package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPool_InvalidDSN(t *testing.T) {
	t.Run("invalid DSN format", func(t *testing.T) {
		pool, sqlDB, err := NewPool("not-a-valid-dsn")

		assert.Error(t, err)

		assert.Nil(t, pool)
		assert.Nil(t, sqlDB)
	})
}

func TestNewPool_ConnectionFailure(t *testing.T) {
	t.Run("connection to non-existent host", func(t *testing.T) {
		dsn := "postgres://user:password@localhost:9999/nonexistent?sslmode=disable"
		pool, sqlDB, err := NewPool(dsn)

		assert.Error(t, err)

		if pool != nil {
			pool.Close()
		}
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
}

func TestNewPool_MissingDSN(t *testing.T) {
	t.Run("empty DSN", func(t *testing.T) {
		pool, sqlDB, err := NewPool("")

		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Nil(t, sqlDB)
	})
}

func TestNewPool_ValidDSNFormat(t *testing.T) {
	t.Run("valid DSN but unreachable host", func(t *testing.T) {
		dsn := "postgresql://user:password@nonexistent-host.invalid:5432/testdb?sslmode=disable"
		pool, sqlDB, err := NewPool(dsn)

		assert.Error(t, err)

		if pool != nil {
			pool.Close()
		}
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
}

func TestNewPool_Integration(t *testing.T) {
	t.Run("valid connection with test database", func(t *testing.T) {
		pool, sqlDB, err := NewPool(testDB.DSN)

		assert.NoError(t, err)
		assert.NotNil(t, pool)
		assert.NotNil(t, sqlDB)

		if pool != nil {
			pool.Close()
		}
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
}
