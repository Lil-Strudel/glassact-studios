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
	DealershipScopeLogin   = "login"
	DealershipScopeAccess  = "access"
	DealershipScopeRefresh = "refresh"
)

type DealershipToken struct {
	Plaintext        string
	Hash             []byte
	DealershipUserID int
	Expiry           time.Time
	Scope            string
}

func generateDealershipToken(dealershipUserID int, ttl time.Duration, scope string) *DealershipToken {
	token := &DealershipToken{
		Plaintext:        rand.Text(),
		DealershipUserID: dealershipUserID,
		Expiry:           time.Now().Add(ttl),
		Scope:            scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

type DealershipTokenModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func (m DealershipTokenModel) New(dealershipUserID int, ttl time.Duration, scope string) (*DealershipToken, error) {
	token := generateDealershipToken(dealershipUserID, ttl, scope)

	err := m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (m DealershipTokenModel) Insert(token *DealershipToken) error {
	query := table.DealershipTokens.INSERT(
		table.DealershipTokens.Hash,
		table.DealershipTokens.DealershipUserID,
		table.DealershipTokens.Expiry,
		table.DealershipTokens.Scope,
	).VALUES(
		token.Hash,
		int32(token.DealershipUserID),
		token.Expiry,
		token.Scope,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m DealershipTokenModel) DeleteAllForUser(scope string, dealershipUserID int) error {
	query := table.DealershipTokens.DELETE().WHERE(
		postgres.AND(
			table.DealershipTokens.Scope.EQ(postgres.String(scope)),
			table.DealershipTokens.DealershipUserID.EQ(postgres.Int(int64(dealershipUserID))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m DealershipTokenModel) DeleteByPlaintext(scope string, plaintext string) error {
	hash := sha256.Sum256([]byte(plaintext))

	query := table.DealershipTokens.DELETE().WHERE(
		postgres.AND(
			table.DealershipTokens.Scope.EQ(postgres.String(scope)),
			table.DealershipTokens.Hash.EQ(postgres.Bytea(hash[:])),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}
