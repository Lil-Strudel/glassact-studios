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

type GlassColor struct {
	StandardTable
	Name      string  `json:"name"`
	Hex       string  `json:"hex"`
	Family    *string `json:"family"`
	SortOrder int     `json:"sort_order"`
	IsActive  bool    `json:"is_active"`
}

type GlassColorModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func glassColorFromGen(gen model.GlassColors) *GlassColor {
	return &GlassColor{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		Name:      gen.Name,
		Hex:       gen.Hex,
		Family:    gen.Family,
		SortOrder: int(gen.SortOrder),
		IsActive:  gen.IsActive,
	}
}

func glassColorToGen(gc *GlassColor) (*model.GlassColors, error) {
	var glassColorUUID uuid.UUID
	var err error

	if gc.UUID != "" {
		glassColorUUID, err = uuid.Parse(gc.UUID)
		if err != nil {
			return nil, err
		}
	}

	return &model.GlassColors{
		ID:        int32(gc.ID),
		UUID:      glassColorUUID,
		Name:      gc.Name,
		Hex:       gc.Hex,
		Family:    gc.Family,
		SortOrder: int32(gc.SortOrder),
		IsActive:  gc.IsActive,
		UpdatedAt: gc.UpdatedAt,
		CreatedAt: gc.CreatedAt,
		Version:   int32(gc.Version),
	}, nil
}

func (m GlassColorModel) Insert(glassColor *GlassColor) error {
	gen, err := glassColorToGen(glassColor)
	if err != nil {
		return err
	}

	query := table.GlassColors.INSERT(
		table.GlassColors.Name,
		table.GlassColors.Hex,
		table.GlassColors.Family,
		table.GlassColors.SortOrder,
		table.GlassColors.IsActive,
	).MODEL(
		gen,
	).RETURNING(
		table.GlassColors.ID,
		table.GlassColors.UUID,
		table.GlassColors.UpdatedAt,
		table.GlassColors.CreatedAt,
		table.GlassColors.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.GlassColors
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	glassColor.ID = int(dest.ID)
	glassColor.UUID = dest.UUID.String()
	glassColor.UpdatedAt = dest.UpdatedAt
	glassColor.CreatedAt = dest.CreatedAt
	glassColor.Version = int(dest.Version)

	return nil
}

func (m GlassColorModel) GetByID(id int) (*GlassColor, bool, error) {
	query := postgres.SELECT(
		table.GlassColors.AllColumns,
	).FROM(
		table.GlassColors,
	).WHERE(
		table.GlassColors.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.GlassColors
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return glassColorFromGen(dest), true, nil
}

func (m GlassColorModel) GetByUUID(uuidStr string) (*GlassColor, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.GlassColors.AllColumns,
	).FROM(
		table.GlassColors,
	).WHERE(
		table.GlassColors.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.GlassColors
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return glassColorFromGen(dest), true, nil
}

func (m GlassColorModel) GetAll() ([]*GlassColor, error) {
	query := postgres.SELECT(
		table.GlassColors.AllColumns,
	).FROM(
		table.GlassColors,
	).ORDER_BY(
		table.GlassColors.SortOrder.ASC(),
		table.GlassColors.Name.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.GlassColors
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	glassColors := make([]*GlassColor, len(dest))
	for i, d := range dest {
		glassColors[i] = glassColorFromGen(d)
	}

	return glassColors, nil
}

func (m GlassColorModel) GetAllActive() ([]*GlassColor, error) {
	query := postgres.SELECT(
		table.GlassColors.AllColumns,
	).FROM(
		table.GlassColors,
	).WHERE(
		table.GlassColors.IsActive.EQ(postgres.Bool(true)),
	).ORDER_BY(
		table.GlassColors.SortOrder.ASC(),
		table.GlassColors.Name.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.GlassColors
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	glassColors := make([]*GlassColor, len(dest))
	for i, d := range dest {
		glassColors[i] = glassColorFromGen(d)
	}

	return glassColors, nil
}

func (m GlassColorModel) Update(glassColor *GlassColor) error {
	gen, err := glassColorToGen(glassColor)
	if err != nil {
		return err
	}

	query := table.GlassColors.UPDATE(
		table.GlassColors.Name,
		table.GlassColors.Hex,
		table.GlassColors.Family,
		table.GlassColors.SortOrder,
		table.GlassColors.IsActive,
		table.GlassColors.Version,
	).MODEL(
		gen,
	).WHERE(
		postgres.AND(
			table.GlassColors.ID.EQ(postgres.Int(int64(glassColor.ID))),
			table.GlassColors.Version.EQ(postgres.Int(int64(glassColor.Version))),
		),
	).RETURNING(
		table.GlassColors.UpdatedAt,
		table.GlassColors.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.GlassColors
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	glassColor.UpdatedAt = dest.UpdatedAt
	glassColor.Version = int(dest.Version)

	return nil
}

func (m GlassColorModel) Delete(id int) error {
	query := table.GlassColors.DELETE().WHERE(
		table.GlassColors.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
