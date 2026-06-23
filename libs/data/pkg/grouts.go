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

type Grout struct {
	StandardTable
	Name      string `json:"name"`
	Hex       string `json:"hex"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type GroutModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func groutFromGen(gen model.Grouts) *Grout {
	return &Grout{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		Name:      gen.Name,
		Hex:       gen.Hex,
		SortOrder: int(gen.SortOrder),
		IsActive:  gen.IsActive,
	}
}

func groutToGen(g *Grout) (*model.Grouts, error) {
	var groutUUID uuid.UUID
	var err error

	if g.UUID != "" {
		groutUUID, err = uuid.Parse(g.UUID)
		if err != nil {
			return nil, err
		}
	}

	return &model.Grouts{
		ID:        int32(g.ID),
		UUID:      groutUUID,
		Name:      g.Name,
		Hex:       g.Hex,
		SortOrder: int32(g.SortOrder),
		IsActive:  g.IsActive,
		UpdatedAt: g.UpdatedAt,
		CreatedAt: g.CreatedAt,
		Version:   int32(g.Version),
	}, nil
}

func (m GroutModel) Insert(grout *Grout) error {
	gen, err := groutToGen(grout)
	if err != nil {
		return err
	}

	query := table.Grouts.INSERT(
		table.Grouts.Name,
		table.Grouts.Hex,
		table.Grouts.SortOrder,
		table.Grouts.IsActive,
	).MODEL(
		gen,
	).RETURNING(
		table.Grouts.ID,
		table.Grouts.UUID,
		table.Grouts.UpdatedAt,
		table.Grouts.CreatedAt,
		table.Grouts.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Grouts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	grout.ID = int(dest.ID)
	grout.UUID = dest.UUID.String()
	grout.UpdatedAt = dest.UpdatedAt
	grout.CreatedAt = dest.CreatedAt
	grout.Version = int(dest.Version)

	return nil
}

func (m GroutModel) GetByID(id int) (*Grout, bool, error) {
	query := postgres.SELECT(
		table.Grouts.AllColumns,
	).FROM(
		table.Grouts,
	).WHERE(
		table.Grouts.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Grouts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return groutFromGen(dest), true, nil
}

func (m GroutModel) GetByUUID(uuidStr string) (*Grout, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.Grouts.AllColumns,
	).FROM(
		table.Grouts,
	).WHERE(
		table.Grouts.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Grouts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return groutFromGen(dest), true, nil
}

func (m GroutModel) GetAll() ([]*Grout, error) {
	query := postgres.SELECT(
		table.Grouts.AllColumns,
	).FROM(
		table.Grouts,
	).ORDER_BY(
		table.Grouts.SortOrder.ASC(),
		table.Grouts.Name.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Grouts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	grouts := make([]*Grout, len(dest))
	for i, d := range dest {
		grouts[i] = groutFromGen(d)
	}

	return grouts, nil
}

func (m GroutModel) GetAllActive() ([]*Grout, error) {
	query := postgres.SELECT(
		table.Grouts.AllColumns,
	).FROM(
		table.Grouts,
	).WHERE(
		table.Grouts.IsActive.EQ(postgres.Bool(true)),
	).ORDER_BY(
		table.Grouts.SortOrder.ASC(),
		table.Grouts.Name.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Grouts
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	grouts := make([]*Grout, len(dest))
	for i, d := range dest {
		grouts[i] = groutFromGen(d)
	}

	return grouts, nil
}

func (m GroutModel) Update(grout *Grout) error {
	gen, err := groutToGen(grout)
	if err != nil {
		return err
	}

	query := table.Grouts.UPDATE(
		table.Grouts.Name,
		table.Grouts.Hex,
		table.Grouts.SortOrder,
		table.Grouts.IsActive,
		table.Grouts.Version,
	).MODEL(
		gen,
	).WHERE(
		postgres.AND(
			table.Grouts.ID.EQ(postgres.Int(int64(grout.ID))),
			table.Grouts.Version.EQ(postgres.Int(int64(grout.Version))),
		),
	).RETURNING(
		table.Grouts.UpdatedAt,
		table.Grouts.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Grouts
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	grout.UpdatedAt = dest.UpdatedAt
	grout.Version = int(dest.Version)

	return nil
}

func (m GroutModel) Delete(id int) error {
	query := table.Grouts.DELETE().WHERE(
		table.Grouts.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
