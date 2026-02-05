package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InternalAccount struct {
	StandardTable
	InternalUserID    int    `json:"internal_user_id"`
	Type              string `json:"type"`
	Provider          string `json:"provider"`
	ProviderAccountID string `json:"provider_account_id"`
}

type InternalAccountModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func internalAccountFromGen(genAcc model.InternalAccounts) *InternalAccount {
	account := InternalAccount{
		StandardTable: StandardTable{
			ID:        int(genAcc.ID),
			UUID:      genAcc.UUID.String(),
			CreatedAt: genAcc.CreatedAt,
			UpdatedAt: genAcc.UpdatedAt,
			Version:   int(genAcc.Version),
		},
		InternalUserID:    int(genAcc.InternalUserID),
		Type:              genAcc.Type,
		Provider:          genAcc.Provider,
		ProviderAccountID: genAcc.ProviderAccountID,
	}

	return &account
}

func internalAccountToGen(a *InternalAccount) (*model.InternalAccounts, error) {
	var accountUUID uuid.UUID
	var err error

	if a.UUID != "" {
		accountUUID, err = uuid.Parse(a.UUID)
		if err != nil {
			return nil, err
		}
	}

	genAcc := model.InternalAccounts{
		ID:                int32(a.ID),
		UUID:              accountUUID,
		InternalUserID:    int32(a.InternalUserID),
		Type:              a.Type,
		Provider:          a.Provider,
		ProviderAccountID: a.ProviderAccountID,
		UpdatedAt:         a.UpdatedAt,
		CreatedAt:         a.CreatedAt,
		Version:           int32(a.Version),
	}

	return &genAcc, nil
}

func (m InternalAccountModel) Insert(account *InternalAccount) error {
	genAcc, err := internalAccountToGen(account)
	if err != nil {
		return err
	}

	query := table.InternalAccounts.INSERT(
		table.InternalAccounts.InternalUserID,
		table.InternalAccounts.Type,
		table.InternalAccounts.Provider,
		table.InternalAccounts.ProviderAccountID,
	).MODEL(
		genAcc,
	).RETURNING(
		table.InternalAccounts.ID,
		table.InternalAccounts.UUID,
		table.InternalAccounts.UpdatedAt,
		table.InternalAccounts.CreatedAt,
		table.InternalAccounts.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalAccounts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	account.ID = int(dest.ID)
	account.UUID = dest.UUID.String()
	account.UpdatedAt = dest.UpdatedAt
	account.CreatedAt = dest.CreatedAt
	account.Version = int(dest.Version)

	return nil
}

func (m InternalAccountModel) GetByID(id int) (*InternalAccount, bool, error) {
	query := postgres.SELECT(
		table.InternalAccounts.AllColumns,
	).FROM(
		table.InternalAccounts,
	).WHERE(
		table.InternalAccounts.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalAccounts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalAccountFromGen(dest), true, nil
}

func (m InternalAccountModel) GetByUUID(uuidStr string) (*InternalAccount, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InternalAccounts.AllColumns,
	).FROM(
		table.InternalAccounts,
	).WHERE(
		table.InternalAccounts.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalAccounts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalAccountFromGen(dest), true, nil
}

func (m InternalAccountModel) GetByProvider(provider string, providerAccountID string) (*InternalAccount, bool, error) {
	query := postgres.SELECT(
		table.InternalAccounts.AllColumns,
	).FROM(
		table.InternalAccounts,
	).WHERE(
		postgres.AND(
			table.InternalAccounts.Provider.EQ(postgres.String(provider)),
			table.InternalAccounts.ProviderAccountID.EQ(postgres.String(providerAccountID)),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalAccounts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalAccountFromGen(dest), true, nil
}

func (m InternalAccountModel) Update(account *InternalAccount) error {
	genAcc, err := internalAccountToGen(account)
	if err != nil {
		return err
	}

	query := table.InternalAccounts.UPDATE(
		table.InternalAccounts.Type,
		table.InternalAccounts.Provider,
		table.InternalAccounts.ProviderAccountID,
		table.InternalAccounts.Version,
	).MODEL(
		genAcc,
	).WHERE(
		postgres.AND(
			table.InternalAccounts.ID.EQ(postgres.Int(int64(account.ID))),
			table.InternalAccounts.Version.EQ(postgres.Int(int64(account.Version))),
		),
	).RETURNING(
		table.InternalAccounts.UpdatedAt,
		table.InternalAccounts.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalAccounts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	account.UpdatedAt = dest.UpdatedAt
	account.Version = int(dest.Version)

	return nil
}

func (m InternalAccountModel) Delete(id int) error {
	query := table.InternalAccounts.DELETE().WHERE(
		table.InternalAccounts.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
