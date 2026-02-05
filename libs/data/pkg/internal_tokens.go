package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	InternalScopeLogin   = "login"
	InternalScopeAccess  = "access"
	InternalScopeRefresh = "refresh"
)

type InternalToken struct {
	Plaintext      string
	Hash           []byte
	InternalUserID int
	Expiry         time.Time
	Scope          string
}

func generateInternalToken(internalUserID int, ttl time.Duration, scope string) *InternalToken {
	token := &InternalToken{
		Plaintext:      rand.Text(),
		InternalUserID: internalUserID,
		Expiry:         time.Now().Add(ttl),
		Scope:          scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

type InternalTokenModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func (m InternalTokenModel) New(internalUserID int, ttl time.Duration, scope string) (*InternalToken, error) {
	token := generateInternalToken(internalUserID, ttl, scope)

	err := m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (m InternalTokenModel) Insert(token *InternalToken) error {
	query := table.InternalTokens.INSERT(
		table.InternalTokens.Hash,
		table.InternalTokens.InternalUserID,
		table.InternalTokens.Expiry,
		table.InternalTokens.Scope,
	).VALUES(
		token.Hash,
		int32(token.InternalUserID),
		token.Expiry,
		token.Scope,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m InternalTokenModel) DeleteAllForUser(scope string, internalUserID int) error {
	query := table.InternalTokens.DELETE().WHERE(
		postgres.AND(
			table.InternalTokens.Scope.EQ(postgres.String(scope)),
			table.InternalTokens.InternalUserID.EQ(postgres.Int(int64(internalUserID))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m InternalTokenModel) DeleteByPlaintext(scope string, plaintext string) error {
	hash := sha256.Sum256([]byte(plaintext))

	query := table.InternalTokens.DELETE().WHERE(
		postgres.AND(
			table.InternalTokens.Scope.EQ(postgres.String(scope)),
			table.InternalTokens.Hash.EQ(postgres.Bytea(hash[:])),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}
