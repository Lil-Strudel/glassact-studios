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

type DealershipAccount struct {
	StandardTable
	DealershipUserID  int    `json:"dealership_user_id"`
	Type              string `json:"type"`
	Provider          string `json:"provider"`
	ProviderAccountID string `json:"provider_account_id"`
}

type DealershipAccountModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func dealershipAccountFromGen(genAcc model.DealershipAccounts) *DealershipAccount {
	account := DealershipAccount{
		StandardTable: StandardTable{
			ID:        int(genAcc.ID),
			UUID:      genAcc.UUID.String(),
			CreatedAt: genAcc.CreatedAt,
			UpdatedAt: genAcc.UpdatedAt,
			Version:   int(genAcc.Version),
		},
		DealershipUserID:  int(genAcc.DealershipUserID),
		Type:              genAcc.Type,
		Provider:          genAcc.Provider,
		ProviderAccountID: genAcc.ProviderAccountID,
	}

	return &account
}

func dealershipAccountToGen(a *DealershipAccount) (*model.DealershipAccounts, error) {
	var accountUUID uuid.UUID
	var err error

	if a.UUID != "" {
		accountUUID, err = uuid.Parse(a.UUID)
		if err != nil {
			return nil, err
		}
	}

	genAcc := model.DealershipAccounts{
		ID:                int32(a.ID),
		UUID:              accountUUID,
		DealershipUserID:  int32(a.DealershipUserID),
		Type:              a.Type,
		Provider:          a.Provider,
		ProviderAccountID: a.ProviderAccountID,
		UpdatedAt:         a.UpdatedAt,
		CreatedAt:         a.CreatedAt,
		Version:           int32(a.Version),
	}

	return &genAcc, nil
}

func (m DealershipAccountModel) Insert(account *DealershipAccount) error {
	genAcc, err := dealershipAccountToGen(account)
	if err != nil {
		return err
	}

	query := table.DealershipAccounts.INSERT(
		table.DealershipAccounts.DealershipUserID,
		table.DealershipAccounts.Type,
		table.DealershipAccounts.Provider,
		table.DealershipAccounts.ProviderAccountID,
	).MODEL(
		genAcc,
	).RETURNING(
		table.DealershipAccounts.ID,
		table.DealershipAccounts.UUID,
		table.DealershipAccounts.UpdatedAt,
		table.DealershipAccounts.CreatedAt,
		table.DealershipAccounts.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipAccounts
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

func (m DealershipAccountModel) GetByID(id int) (*DealershipAccount, bool, error) {
	query := postgres.SELECT(
		table.DealershipAccounts.AllColumns,
	).FROM(
		table.DealershipAccounts,
	).WHERE(
		table.DealershipAccounts.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipAccounts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipAccountFromGen(dest), true, nil
}

func (m DealershipAccountModel) GetByUUID(uuidStr string) (*DealershipAccount, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.DealershipAccounts.AllColumns,
	).FROM(
		table.DealershipAccounts,
	).WHERE(
		table.DealershipAccounts.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipAccounts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipAccountFromGen(dest), true, nil
}

func (m DealershipAccountModel) GetByProvider(provider string, providerAccountID string) (*DealershipAccount, bool, error) {
	query := postgres.SELECT(
		table.DealershipAccounts.AllColumns,
	).FROM(
		table.DealershipAccounts,
	).WHERE(
		postgres.AND(
			table.DealershipAccounts.Provider.EQ(postgres.String(provider)),
			table.DealershipAccounts.ProviderAccountID.EQ(postgres.String(providerAccountID)),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipAccounts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipAccountFromGen(dest), true, nil
}

func (m DealershipAccountModel) Update(account *DealershipAccount) error {
	genAcc, err := dealershipAccountToGen(account)
	if err != nil {
		return err
	}

	query := table.DealershipAccounts.UPDATE(
		table.DealershipAccounts.Type,
		table.DealershipAccounts.Provider,
		table.DealershipAccounts.ProviderAccountID,
		table.DealershipAccounts.Version,
	).MODEL(
		genAcc,
	).WHERE(
		postgres.AND(
			table.DealershipAccounts.ID.EQ(postgres.Int(int64(account.ID))),
			table.DealershipAccounts.Version.EQ(postgres.Int(int64(account.Version))),
		),
	).RETURNING(
		table.DealershipAccounts.UpdatedAt,
		table.DealershipAccounts.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipAccounts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	account.UpdatedAt = dest.UpdatedAt
	account.Version = int(dest.Version)

	return nil
}

func (m DealershipAccountModel) Delete(id int) error {
	query := table.DealershipAccounts.DELETE().WHERE(
		table.DealershipAccounts.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
