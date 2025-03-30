package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m UserModel) Insert(user *User) error {
	query := `
        INSERT INTO users () 
        VALUES ()
        RETURNING id, uuid, created_at, version`

	args := []any{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&user.ID, &user.UUID, &user.CreatedAt, &user.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) GetByID(id int) (*User, error) {
	query := `
        SELECT id, uuid, created_at, version
        FROM users
        WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(&user.ID, &user.UUID, &user.CreatedAt, &user.Version)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m UserModel) GetByUUID(uuid string) (*User, error) {
	query := `
        SELECT id, uuid, created_at, version
        FROM users
        WHERE uuid = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(&user.ID, &user.UUID, &user.CreatedAt, &user.Version)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
        UPDATE users 
        SET version = version + 1
        WHERE id = $1 AND version = $2
        RETURNING version`

	args := []any{
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) Delete(id int) error {
	query := `
        DELETE FROM users
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
