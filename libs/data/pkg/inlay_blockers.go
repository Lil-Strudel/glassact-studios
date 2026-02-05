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

type BlockerType string

type blockerTypes struct {
	Soft BlockerType
	Hard BlockerType
}

var BlockerTypes = blockerTypes{
	Soft: BlockerType("soft"),
	Hard: BlockerType("hard"),
}

type InlayBlocker struct {
	StandardTable
	InlayID         int         `json:"inlay_id"`
	BlockerType     BlockerType `json:"blocker_type"`
	Reason          string      `json:"reason"`
	StepBlocked     string      `json:"step_blocked"`
	CreatedBy       *int        `json:"created_by"`
	ResolvedAt      *time.Time  `json:"resolved_at"`
	ResolvedBy      *int        `json:"resolved_by"`
	ResolutionNotes *string     `json:"resolution_notes"`
}

type InlayBlockerModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayBlockerFromGen(genBlocker model.InlayBlockers) *InlayBlocker {
	var createdBy *int
	if genBlocker.CreatedBy != nil {
		createdByVal := int(*genBlocker.CreatedBy)
		createdBy = &createdByVal
	}

	var resolvedBy *int
	if genBlocker.ResolvedBy != nil {
		resolvedByVal := int(*genBlocker.ResolvedBy)
		resolvedBy = &resolvedByVal
	}

	blocker := InlayBlocker{
		StandardTable: StandardTable{
			ID:        int(genBlocker.ID),
			UUID:      genBlocker.UUID.String(),
			CreatedAt: genBlocker.CreatedAt,
			UpdatedAt: genBlocker.UpdatedAt,
			Version:   int(genBlocker.Version),
		},
		InlayID:         int(genBlocker.InlayID),
		BlockerType:     BlockerType(genBlocker.BlockerType),
		Reason:          genBlocker.Reason,
		StepBlocked:     genBlocker.StepBlocked,
		CreatedBy:       createdBy,
		ResolvedAt:      genBlocker.ResolvedAt,
		ResolvedBy:      resolvedBy,
		ResolutionNotes: genBlocker.ResolutionNotes,
	}

	return &blocker
}

func inlayBlockerToGen(ib *InlayBlocker) (*model.InlayBlockers, error) {
	var blockerUUID uuid.UUID
	var err error

	if ib.UUID != "" {
		blockerUUID, err = uuid.Parse(ib.UUID)
		if err != nil {
			return nil, err
		}
	}

	var createdBy *int32
	if ib.CreatedBy != nil {
		createdByVal := int32(*ib.CreatedBy)
		createdBy = &createdByVal
	}

	var resolvedBy *int32
	if ib.ResolvedBy != nil {
		resolvedByVal := int32(*ib.ResolvedBy)
		resolvedBy = &resolvedByVal
	}

	genBlocker := model.InlayBlockers{
		ID:              int32(ib.ID),
		UUID:            blockerUUID,
		InlayID:         int32(ib.InlayID),
		BlockerType:     string(ib.BlockerType),
		Reason:          ib.Reason,
		StepBlocked:     ib.StepBlocked,
		CreatedBy:       createdBy,
		ResolvedAt:      ib.ResolvedAt,
		ResolvedBy:      resolvedBy,
		ResolutionNotes: ib.ResolutionNotes,
		UpdatedAt:       ib.UpdatedAt,
		CreatedAt:       ib.CreatedAt,
		Version:         int32(ib.Version),
	}

	return &genBlocker, nil
}

func (m InlayBlockerModel) Insert(blocker *InlayBlocker) error {
	genBlocker, err := inlayBlockerToGen(blocker)
	if err != nil {
		return err
	}

	query := table.InlayBlockers.INSERT(
		table.InlayBlockers.InlayID,
		table.InlayBlockers.BlockerType,
		table.InlayBlockers.Reason,
		table.InlayBlockers.StepBlocked,
		table.InlayBlockers.CreatedBy,
	).MODEL(
		genBlocker,
	).RETURNING(
		table.InlayBlockers.ID,
		table.InlayBlockers.UUID,
		table.InlayBlockers.UpdatedAt,
		table.InlayBlockers.CreatedAt,
		table.InlayBlockers.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayBlockers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	blocker.ID = int(dest.ID)
	blocker.UUID = dest.UUID.String()
	blocker.UpdatedAt = dest.UpdatedAt
	blocker.CreatedAt = dest.CreatedAt
	blocker.Version = int(dest.Version)

	return nil
}

func (m InlayBlockerModel) GetByID(id int) (*InlayBlocker, bool, error) {
	query := postgres.SELECT(
		table.InlayBlockers.AllColumns,
	).FROM(
		table.InlayBlockers,
	).WHERE(
		table.InlayBlockers.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayBlockers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayBlockerFromGen(dest), true, nil
}

func (m InlayBlockerModel) GetByUUID(uuidStr string) (*InlayBlocker, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InlayBlockers.AllColumns,
	).FROM(
		table.InlayBlockers,
	).WHERE(
		table.InlayBlockers.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayBlockers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayBlockerFromGen(dest), true, nil
}

func (m InlayBlockerModel) GetByInlayID(inlayID int) ([]*InlayBlocker, error) {
	query := postgres.SELECT(
		table.InlayBlockers.AllColumns,
	).FROM(
		table.InlayBlockers,
	).WHERE(
		table.InlayBlockers.InlayID.EQ(postgres.Int(int64(inlayID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayBlockers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	blockers := make([]*InlayBlocker, len(dest))
	for i, d := range dest {
		blockers[i] = inlayBlockerFromGen(d)
	}

	return blockers, nil
}

func (m InlayBlockerModel) GetUnresolved(inlayID int) ([]*InlayBlocker, error) {
	query := postgres.SELECT(
		table.InlayBlockers.AllColumns,
	).FROM(
		table.InlayBlockers,
	).WHERE(
		postgres.AND(
			table.InlayBlockers.InlayID.EQ(postgres.Int(int64(inlayID))),
			table.InlayBlockers.ResolvedAt.IS_NULL(),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayBlockers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	blockers := make([]*InlayBlocker, len(dest))
	for i, d := range dest {
		blockers[i] = inlayBlockerFromGen(d)
	}

	return blockers, nil
}

func (m InlayBlockerModel) GetAll() ([]*InlayBlocker, error) {
	query := postgres.SELECT(
		table.InlayBlockers.AllColumns,
	).FROM(
		table.InlayBlockers,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayBlockers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	blockers := make([]*InlayBlocker, len(dest))
	for i, d := range dest {
		blockers[i] = inlayBlockerFromGen(d)
	}

	return blockers, nil
}

func (m InlayBlockerModel) Update(blocker *InlayBlocker) error {
	genBlocker, err := inlayBlockerToGen(blocker)
	if err != nil {
		return err
	}

	query := table.InlayBlockers.UPDATE(
		table.InlayBlockers.BlockerType,
		table.InlayBlockers.Reason,
		table.InlayBlockers.StepBlocked,
		table.InlayBlockers.ResolvedAt,
		table.InlayBlockers.ResolvedBy,
		table.InlayBlockers.ResolutionNotes,
		table.InlayBlockers.Version,
	).MODEL(
		genBlocker,
	).WHERE(
		postgres.AND(
			table.InlayBlockers.ID.EQ(postgres.Int(int64(blocker.ID))),
			table.InlayBlockers.Version.EQ(postgres.Int(int64(blocker.Version))),
		),
	).RETURNING(
		table.InlayBlockers.UpdatedAt,
		table.InlayBlockers.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayBlockers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	blocker.UpdatedAt = dest.UpdatedAt
	blocker.Version = int(dest.Version)

	return nil
}

func (m InlayBlockerModel) Delete(id int) error {
	query := table.InlayBlockers.DELETE().WHERE(
		table.InlayBlockers.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
