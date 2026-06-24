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

type InlayUpdateType string

type inlayUpdateTypes struct {
	Info  InlayUpdateType
	Issue InlayUpdateType
}

var InlayUpdateTypes = inlayUpdateTypes{
	Info:  InlayUpdateType("info"),
	Issue: InlayUpdateType("issue"),
}

type InlayUpdate struct {
	StandardTable
	InlayID    int             `json:"inlay_id"`
	UpdateType InlayUpdateType `json:"update_type"`
	Message    string          `json:"message"`
	Step       *string         `json:"step"`
	CreatedBy  *int            `json:"created_by"`
}

type InlayUpdateModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayUpdateFromGen(genUpdate model.InlayUpdates) *InlayUpdate {
	var createdBy *int
	if genUpdate.CreatedBy != nil {
		createdByVal := int(*genUpdate.CreatedBy)
		createdBy = &createdByVal
	}

	update := InlayUpdate{
		StandardTable: StandardTable{
			ID:        int(genUpdate.ID),
			UUID:      genUpdate.UUID.String(),
			CreatedAt: genUpdate.CreatedAt,
			UpdatedAt: genUpdate.UpdatedAt,
			Version:   int(genUpdate.Version),
		},
		InlayID:    int(genUpdate.InlayID),
		UpdateType: InlayUpdateType(genUpdate.UpdateType),
		Message:    genUpdate.Message,
		Step:       genUpdate.Step,
		CreatedBy:  createdBy,
	}

	return &update
}

func inlayUpdateToGen(iu *InlayUpdate) (*model.InlayUpdates, error) {
	var updateUUID uuid.UUID
	var err error

	if iu.UUID != "" {
		updateUUID, err = uuid.Parse(iu.UUID)
		if err != nil {
			return nil, err
		}
	}

	var createdBy *int32
	if iu.CreatedBy != nil {
		createdByVal := int32(*iu.CreatedBy)
		createdBy = &createdByVal
	}

	genUpdate := model.InlayUpdates{
		ID:         int32(iu.ID),
		UUID:       updateUUID,
		InlayID:    int32(iu.InlayID),
		UpdateType: string(iu.UpdateType),
		Message:    iu.Message,
		Step:       iu.Step,
		CreatedBy:  createdBy,
		UpdatedAt:  iu.UpdatedAt,
		CreatedAt:  iu.CreatedAt,
		Version:    int32(iu.Version),
	}

	return &genUpdate, nil
}

func (m InlayUpdateModel) Insert(update *InlayUpdate) error {
	genUpdate, err := inlayUpdateToGen(update)
	if err != nil {
		return err
	}

	query := table.InlayUpdates.INSERT(
		table.InlayUpdates.InlayID,
		table.InlayUpdates.UpdateType,
		table.InlayUpdates.Message,
		table.InlayUpdates.Step,
		table.InlayUpdates.CreatedBy,
	).MODEL(
		genUpdate,
	).RETURNING(
		table.InlayUpdates.ID,
		table.InlayUpdates.UUID,
		table.InlayUpdates.UpdatedAt,
		table.InlayUpdates.CreatedAt,
		table.InlayUpdates.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayUpdates
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	update.ID = int(dest.ID)
	update.UUID = dest.UUID.String()
	update.UpdatedAt = dest.UpdatedAt
	update.CreatedAt = dest.CreatedAt
	update.Version = int(dest.Version)

	return nil
}

func (m InlayUpdateModel) TxInsert(tx *sql.Tx, update *InlayUpdate) error {
	genUpdate, err := inlayUpdateToGen(update)
	if err != nil {
		return err
	}

	query := table.InlayUpdates.INSERT(
		table.InlayUpdates.InlayID,
		table.InlayUpdates.UpdateType,
		table.InlayUpdates.Message,
		table.InlayUpdates.Step,
		table.InlayUpdates.CreatedBy,
	).MODEL(
		genUpdate,
	).RETURNING(
		table.InlayUpdates.ID,
		table.InlayUpdates.UUID,
		table.InlayUpdates.UpdatedAt,
		table.InlayUpdates.CreatedAt,
		table.InlayUpdates.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayUpdates
	err = query.QueryContext(ctx, tx, &dest)
	if err != nil {
		return err
	}

	update.ID = int(dest.ID)
	update.UUID = dest.UUID.String()
	update.UpdatedAt = dest.UpdatedAt
	update.CreatedAt = dest.CreatedAt
	update.Version = int(dest.Version)

	return nil
}

func (m InlayUpdateModel) GetByID(id int) (*InlayUpdate, bool, error) {
	query := postgres.SELECT(
		table.InlayUpdates.AllColumns,
	).FROM(
		table.InlayUpdates,
	).WHERE(
		table.InlayUpdates.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayUpdates
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayUpdateFromGen(dest), true, nil
}

func (m InlayUpdateModel) GetByUUID(uuidStr string) (*InlayUpdate, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InlayUpdates.AllColumns,
	).FROM(
		table.InlayUpdates,
	).WHERE(
		table.InlayUpdates.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayUpdates
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayUpdateFromGen(dest), true, nil
}

func (m InlayUpdateModel) GetByInlayID(inlayID int) ([]*InlayUpdate, error) {
	query := postgres.SELECT(
		table.InlayUpdates.AllColumns,
	).FROM(
		table.InlayUpdates,
	).WHERE(
		table.InlayUpdates.InlayID.EQ(postgres.Int(int64(inlayID))),
	).ORDER_BY(
		table.InlayUpdates.CreatedAt.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayUpdates
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	updates := make([]*InlayUpdate, len(dest))
	for i, d := range dest {
		updates[i] = inlayUpdateFromGen(d)
	}

	return updates, nil
}

func (m InlayUpdateModel) GetAll() ([]*InlayUpdate, error) {
	query := postgres.SELECT(
		table.InlayUpdates.AllColumns,
	).FROM(
		table.InlayUpdates,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayUpdates
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	updates := make([]*InlayUpdate, len(dest))
	for i, d := range dest {
		updates[i] = inlayUpdateFromGen(d)
	}

	return updates, nil
}

func (m InlayUpdateModel) Update(update *InlayUpdate) error {
	genUpdate, err := inlayUpdateToGen(update)
	if err != nil {
		return err
	}

	query := table.InlayUpdates.UPDATE(
		table.InlayUpdates.UpdateType,
		table.InlayUpdates.Message,
		table.InlayUpdates.Step,
		table.InlayUpdates.Version,
	).MODEL(
		genUpdate,
	).WHERE(
		postgres.AND(
			table.InlayUpdates.ID.EQ(postgres.Int(int64(update.ID))),
			table.InlayUpdates.Version.EQ(postgres.Int(int64(update.Version))),
		),
	).RETURNING(
		table.InlayUpdates.UpdatedAt,
		table.InlayUpdates.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayUpdates
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	update.UpdatedAt = dest.UpdatedAt
	update.Version = int(dest.Version)

	return nil
}

func (m InlayUpdateModel) Delete(id int) error {
	query := table.InlayUpdates.DELETE().WHERE(
		table.InlayUpdates.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
