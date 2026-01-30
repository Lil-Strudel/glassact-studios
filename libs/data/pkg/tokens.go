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
	ScopeLogin   = "login"
	ScopeAccess  = "access"
	ScopeRefresh = "refresh"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int, ttl time.Duration, scope string) *Token {
	token := &Token{
		Plaintext: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

type TokenModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func (m TokenModel) New(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := generateToken(userID, ttl, scope)

	err := m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (m TokenModel) Insert(token *Token) error {
	query := table.Tokens.INSERT(
		table.Tokens.Hash,
		table.Tokens.UserID,
		table.Tokens.Expiry,
		table.Tokens.Scope,
	).VALUES(
		token.Hash,
		int32(token.UserID),
		token.Expiry,
		token.Scope,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, userID int) error {
	query := table.Tokens.DELETE().WHERE(
		postgres.AND(
			table.Tokens.Scope.EQ(postgres.String(scope)),
			table.Tokens.UserID.EQ(postgres.Int(int64(userID))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m TokenModel) DeleteByPlaintext(scope string, plaintext string) error {
	hash := sha256.Sum256([]byte(plaintext))

	query := table.Tokens.DELETE().WHERE(
		postgres.AND(
			table.Tokens.Scope.EQ(postgres.String(scope)),
			table.Tokens.Hash.EQ(postgres.Bytea(hash[:])),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}
