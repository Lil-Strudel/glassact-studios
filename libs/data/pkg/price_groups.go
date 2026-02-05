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

type PriceGroup struct {
	StandardTable
	Name           string  `json:"name"`
	BasePriceCents int     `json:"base_price_cents"`
	Description    *string `json:"description"`
	IsActive       bool    `json:"is_active"`
}

type PriceGroupModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func priceGroupFromGen(genPriceGroup model.PriceGroups) *PriceGroup {
	priceGroup := PriceGroup{
		StandardTable: StandardTable{
			ID:        int(genPriceGroup.ID),
			UUID:      genPriceGroup.UUID.String(),
			CreatedAt: genPriceGroup.CreatedAt,
			UpdatedAt: genPriceGroup.UpdatedAt,
			Version:   int(genPriceGroup.Version),
		},
		Name:           genPriceGroup.Name,
		BasePriceCents: int(genPriceGroup.BasePriceCents),
		Description:    genPriceGroup.Description,
		IsActive:       genPriceGroup.IsActive,
	}

	return &priceGroup
}

func priceGroupToGen(pg *PriceGroup) (*model.PriceGroups, error) {
	var priceGroupUUID uuid.UUID
	var err error

	if pg.UUID != "" {
		priceGroupUUID, err = uuid.Parse(pg.UUID)
		if err != nil {
			return nil, err
		}
	}

	genPriceGroup := model.PriceGroups{
		ID:             int32(pg.ID),
		UUID:           priceGroupUUID,
		Name:           pg.Name,
		BasePriceCents: int32(pg.BasePriceCents),
		Description:    pg.Description,
		IsActive:       pg.IsActive,
		UpdatedAt:      pg.UpdatedAt,
		CreatedAt:      pg.CreatedAt,
		Version:        int32(pg.Version),
	}

	return &genPriceGroup, nil
}

func (m PriceGroupModel) Insert(priceGroup *PriceGroup) error {
	genPriceGroup, err := priceGroupToGen(priceGroup)
	if err != nil {
		return err
	}

	query := table.PriceGroups.INSERT(
		table.PriceGroups.Name,
		table.PriceGroups.BasePriceCents,
		table.PriceGroups.Description,
		table.PriceGroups.IsActive,
	).MODEL(
		genPriceGroup,
	).RETURNING(
		table.PriceGroups.ID,
		table.PriceGroups.UUID,
		table.PriceGroups.UpdatedAt,
		table.PriceGroups.CreatedAt,
		table.PriceGroups.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.PriceGroups
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	priceGroup.ID = int(dest.ID)
	priceGroup.UUID = dest.UUID.String()
	priceGroup.UpdatedAt = dest.UpdatedAt
	priceGroup.CreatedAt = dest.CreatedAt
	priceGroup.Version = int(dest.Version)

	return nil
}

func (m PriceGroupModel) GetByID(id int) (*PriceGroup, bool, error) {
	query := postgres.SELECT(
		table.PriceGroups.AllColumns,
	).FROM(
		table.PriceGroups,
	).WHERE(
		table.PriceGroups.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.PriceGroups
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return priceGroupFromGen(dest), true, nil
}

func (m PriceGroupModel) GetByUUID(uuidStr string) (*PriceGroup, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.PriceGroups.AllColumns,
	).FROM(
		table.PriceGroups,
	).WHERE(
		table.PriceGroups.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.PriceGroups
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return priceGroupFromGen(dest), true, nil
}

func (m PriceGroupModel) GetAll() ([]*PriceGroup, error) {
	query := postgres.SELECT(
		table.PriceGroups.AllColumns,
	).FROM(
		table.PriceGroups,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.PriceGroups
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	priceGroups := make([]*PriceGroup, len(dest))
	for i, d := range dest {
		priceGroups[i] = priceGroupFromGen(d)
	}

	return priceGroups, nil
}

func (m PriceGroupModel) Update(priceGroup *PriceGroup) error {
	genPriceGroup, err := priceGroupToGen(priceGroup)
	if err != nil {
		return err
	}

	query := table.PriceGroups.UPDATE(
		table.PriceGroups.Name,
		table.PriceGroups.BasePriceCents,
		table.PriceGroups.Description,
		table.PriceGroups.IsActive,
		table.PriceGroups.Version,
	).MODEL(
		genPriceGroup,
	).WHERE(
		postgres.AND(
			table.PriceGroups.ID.EQ(postgres.Int(int64(priceGroup.ID))),
			table.PriceGroups.Version.EQ(postgres.Int(int64(priceGroup.Version))),
		),
	).RETURNING(
		table.PriceGroups.UpdatedAt,
		table.PriceGroups.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.PriceGroups
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	priceGroup.UpdatedAt = dest.UpdatedAt
	priceGroup.Version = int(dest.Version)

	return nil
}

func (m PriceGroupModel) Delete(id int) error {
	query := table.PriceGroups.DELETE().WHERE(
		table.PriceGroups.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
