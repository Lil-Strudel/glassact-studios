package data

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m UserModel) Insert(user *User) error {
	query := `
        INSERT INTO users (name, email, avatar) 
        VALUES ($1, $2, $3)
        RETURNING id, uuid, created_at, version`

	args := []any{user.Name, user.Email, user.Avatar}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&user.ID, &user.UUID, &user.CreatedAt, &user.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) GetByID(id int) (*User, bool, error) {
	query := `
        SELECT id, uuid, name, email, avatar, created_at, version
        FROM users
        WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.Avatar, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &user, true, nil
}

func (m UserModel) GetByUUID(uuid string) (*User, bool, error) {
	query := `
        SELECT id, uuid, name, email, avatar, created_at, version
        FROM users
        WHERE uuid = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.Avatar, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &user, true, nil
}

func (m UserModel) GetByEmail(email string) (*User, bool, error) {
	query := `
        SELECT id, uuid, name, email, avatar, created_at, version
        FROM users
        WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, email).Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.Avatar, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &user, true, nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, bool, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT users.id, users.uuid, users.name, users.email, users.avatar, users.created_at, users.version
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Avatar,
		&user.CreatedAt,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &user, true, nil
}

func (m UserModel) Update(user *User) error {
	query := `
        UPDATE users 
        SET name = $1, email = $2, avatar = $3, version = version + 1
        WHERE id = $4 AND version = $5
        RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Avatar,
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
