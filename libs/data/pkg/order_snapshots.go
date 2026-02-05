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

type OrderSnapshot struct {
	ID           int       `json:"id"`
	UUID         string    `json:"uuid"`
	ProjectID    int       `json:"project_id"`
	InlayID      int       `json:"inlay_id"`
	ProofID      int       `json:"proof_id"`
	PriceGroupID int       `json:"price_group_id"`
	PriceCents   int       `json:"price_cents"`
	Width        float64   `json:"width"`
	Height       float64   `json:"height"`
	CreatedAt    time.Time `json:"created_at"`
}

type OrderSnapshotModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func orderSnapshotFromGen(genSnapshot model.OrderSnapshots) *OrderSnapshot {
	snapshot := OrderSnapshot{
		ID:           int(genSnapshot.ID),
		UUID:         genSnapshot.UUID.String(),
		ProjectID:    int(genSnapshot.ProjectID),
		InlayID:      int(genSnapshot.InlayID),
		ProofID:      int(genSnapshot.ProofID),
		PriceGroupID: int(genSnapshot.PriceGroupID),
		PriceCents:   int(genSnapshot.PriceCents),
		Width:        genSnapshot.Width,
		Height:       genSnapshot.Height,
		CreatedAt:    genSnapshot.CreatedAt,
	}

	return &snapshot
}

func orderSnapshotToGen(os *OrderSnapshot) (*model.OrderSnapshots, error) {
	var snapshotUUID uuid.UUID
	var err error

	if os.UUID != "" {
		snapshotUUID, err = uuid.Parse(os.UUID)
		if err != nil {
			return nil, err
		}
	}

	genSnapshot := model.OrderSnapshots{
		ID:           int32(os.ID),
		UUID:         snapshotUUID,
		ProjectID:    int32(os.ProjectID),
		InlayID:      int32(os.InlayID),
		ProofID:      int32(os.ProofID),
		PriceGroupID: int32(os.PriceGroupID),
		PriceCents:   int32(os.PriceCents),
		Width:        os.Width,
		Height:       os.Height,
		CreatedAt:    os.CreatedAt,
	}

	return &genSnapshot, nil
}

func (m OrderSnapshotModel) Insert(orderSnapshot *OrderSnapshot) error {
	genSnapshot, err := orderSnapshotToGen(orderSnapshot)
	if err != nil {
		return err
	}

	query := table.OrderSnapshots.INSERT(
		table.OrderSnapshots.ProjectID,
		table.OrderSnapshots.InlayID,
		table.OrderSnapshots.ProofID,
		table.OrderSnapshots.PriceGroupID,
		table.OrderSnapshots.PriceCents,
		table.OrderSnapshots.Width,
		table.OrderSnapshots.Height,
	).MODEL(
		genSnapshot,
	).RETURNING(
		table.OrderSnapshots.ID,
		table.OrderSnapshots.UUID,
		table.OrderSnapshots.CreatedAt,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.OrderSnapshots
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	orderSnapshot.ID = int(dest.ID)
	orderSnapshot.UUID = dest.UUID.String()
	orderSnapshot.CreatedAt = dest.CreatedAt

	return nil
}

func (m OrderSnapshotModel) GetByID(id int) (*OrderSnapshot, bool, error) {
	query := postgres.SELECT(
		table.OrderSnapshots.AllColumns,
	).FROM(
		table.OrderSnapshots,
	).WHERE(
		table.OrderSnapshots.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.OrderSnapshots
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return orderSnapshotFromGen(dest), true, nil
}

func (m OrderSnapshotModel) GetByUUID(uuidStr string) (*OrderSnapshot, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.OrderSnapshots.AllColumns,
	).FROM(
		table.OrderSnapshots,
	).WHERE(
		table.OrderSnapshots.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.OrderSnapshots
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return orderSnapshotFromGen(dest), true, nil
}

func (m OrderSnapshotModel) GetByInlayID(inlayID int) (*OrderSnapshot, bool, error) {
	query := postgres.SELECT(
		table.OrderSnapshots.AllColumns,
	).FROM(
		table.OrderSnapshots,
	).WHERE(
		table.OrderSnapshots.InlayID.EQ(postgres.Int(int64(inlayID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.OrderSnapshots
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return orderSnapshotFromGen(dest), true, nil
}

func (m OrderSnapshotModel) GetByProjectID(projectID int) ([]*OrderSnapshot, error) {
	query := postgres.SELECT(
		table.OrderSnapshots.AllColumns,
	).FROM(
		table.OrderSnapshots,
	).WHERE(
		table.OrderSnapshots.ProjectID.EQ(postgres.Int(int64(projectID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.OrderSnapshots
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	snapshots := make([]*OrderSnapshot, len(dest))
	for i, d := range dest {
		snapshots[i] = orderSnapshotFromGen(d)
	}

	return snapshots, nil
}

func (m OrderSnapshotModel) GetAll() ([]*OrderSnapshot, error) {
	query := postgres.SELECT(
		table.OrderSnapshots.AllColumns,
	).FROM(
		table.OrderSnapshots,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.OrderSnapshots
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	snapshots := make([]*OrderSnapshot, len(dest))
	for i, d := range dest {
		snapshots[i] = orderSnapshotFromGen(d)
	}

	return snapshots, nil
}

func (m OrderSnapshotModel) Delete(id int) error {
	query := table.OrderSnapshots.DELETE().WHERE(
		table.OrderSnapshots.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
