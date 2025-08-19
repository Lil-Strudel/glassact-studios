package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID                int       `json:"id"`
	UUID              string    `json:"uuid"`
	UserID            int       `json:"user_id"`
	Type              string    `json:"type"`
	Provider          string    `json:"provider"`
	ProviderAccountID string    `json:"provider_account_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Version           int       `json:"version"`
}

type AccountModel struct {
	DB *pgxpool.Pool
}

func (m AccountModel) Insert(account *Account) error {
	query := `
        INSERT INTO accounts (user_id, type, provider, provider_account_id) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		account.UserID,
		account.Type,
		account.Provider,
		account.ProviderAccountID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&account.ID, &account.UUID, &account.CreatedAt, &account.UpdatedAt, &account.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m AccountModel) GetByID(id int) (*Account, bool, error) {
	query := `
        SELECT id, uuid, user_id, type, provider, provider_account_id, created_at, updated_at, version
        FROM accounts
        WHERE id = $1`

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(&account.ID, &account.UUID, &account.UserID, &account.Type, &account.Provider, &account.ProviderAccountID, &account.CreatedAt, &account.UpdatedAt, &account.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &account, true, nil
}

func (m AccountModel) GetByUUID(uuid string) (*Account, bool, error) {
	query := `
        SELECT id, uuid, user_id, type, provider, provider_account_id, created_at, updated_at, version
        FROM accounts
        WHERE uuid = $1`

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(&account.ID, &account.UUID, &account.UserID, &account.Type, &account.Provider, &account.ProviderAccountID, &account.CreatedAt, &account.UpdatedAt, &account.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &account, true, nil
}

func (m AccountModel) GetByProvider(provider string, providerAccountId string) (*Account, bool, error) {
	query := `
        SELECT id, uuid, user_id, type, provider, provider_account_id, created_at, updated_at, version
        FROM accounts
        WHERE provider = $1 AND provider_account_id = $2`

	args := []any{
		provider,
		providerAccountId,
	}

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&account.ID, &account.UUID, &account.UserID, &account.Type, &account.Provider, &account.ProviderAccountID, &account.CreatedAt, &account.UpdatedAt, &account.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &account, true, nil
}

func (m AccountModel) Update(account *Account) error {
	query := `
        UPDATE accounts 
        SET type = $1, provider = $2, provider_account_id = $3, version = version + 1
        WHERE id = $4 AND version = $5
        RETURNING version`

	args := []any{
		account.Type,
		account.Provider,
		account.ProviderAccountID,
		account.ID,
		account.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&account.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m AccountModel) Delete(id int) error {
	query := `
        DELETE FROM accounts
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
