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

type ManufacturingStep string

type manufacturingSteps struct {
	Ordered       ManufacturingStep
	MaterialsPrep ManufacturingStep
	Cutting       ManufacturingStep
	FirePolish    ManufacturingStep
	Packaging     ManufacturingStep
	Shipped       ManufacturingStep
	Delivered     ManufacturingStep
}

var ManufacturingSteps = manufacturingSteps{
	Ordered:       ManufacturingStep("ordered"),
	MaterialsPrep: ManufacturingStep("materials-prep"),
	Cutting:       ManufacturingStep("cutting"),
	FirePolish:    ManufacturingStep("fire-polish"),
	Packaging:     ManufacturingStep("packaging"),
	Shipped:       ManufacturingStep("shipped"),
	Delivered:     ManufacturingStep("delivered"),
}

type MilestoneEventType string

type milestoneEventTypes struct {
	Entered  MilestoneEventType
	Exited   MilestoneEventType
	Reverted MilestoneEventType
}

var MilestoneEventTypes = milestoneEventTypes{
	Entered:  MilestoneEventType("entered"),
	Exited:   MilestoneEventType("exited"),
	Reverted: MilestoneEventType("reverted"),
}

type InlayMilestone struct {
	StandardTable
	InlayID     int                `json:"inlay_id"`
	Step        ManufacturingStep  `json:"step"`
	EventType   MilestoneEventType `json:"event_type"`
	PerformedBy int                `json:"performed_by"`
	EventTime   time.Time          `json:"event_time"`
}

type InlayMilestoneModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayMilestoneFromGen(genMilestone model.InlayMilestones) *InlayMilestone {
	milestone := InlayMilestone{
		StandardTable: StandardTable{
			ID:        int(genMilestone.ID),
			UUID:      genMilestone.UUID.String(),
			CreatedAt: genMilestone.CreatedAt,
			UpdatedAt: genMilestone.UpdatedAt,
			Version:   int(genMilestone.Version),
		},
		InlayID:     int(genMilestone.InlayID),
		Step:        ManufacturingStep(genMilestone.Step),
		EventType:   MilestoneEventType(genMilestone.EventType),
		PerformedBy: int(genMilestone.PerformedBy),
		EventTime:   genMilestone.EventTime,
	}

	return &milestone
}

func inlayMilestoneToGen(im *InlayMilestone) (*model.InlayMilestones, error) {
	var milestoneUUID uuid.UUID
	var err error

	if im.UUID != "" {
		milestoneUUID, err = uuid.Parse(im.UUID)
		if err != nil {
			return nil, err
		}
	}

	genMilestone := model.InlayMilestones{
		ID:          int32(im.ID),
		UUID:        milestoneUUID,
		InlayID:     int32(im.InlayID),
		Step:        string(im.Step),
		EventType:   string(im.EventType),
		PerformedBy: int32(im.PerformedBy),
		EventTime:   im.EventTime,
		UpdatedAt:   im.UpdatedAt,
		CreatedAt:   im.CreatedAt,
		Version:     int32(im.Version),
	}

	return &genMilestone, nil
}

func (m InlayMilestoneModel) Insert(milestone *InlayMilestone) error {
	genMilestone, err := inlayMilestoneToGen(milestone)
	if err != nil {
		return err
	}

	query := table.InlayMilestones.INSERT(
		table.InlayMilestones.InlayID,
		table.InlayMilestones.Step,
		table.InlayMilestones.EventType,
		table.InlayMilestones.PerformedBy,
		table.InlayMilestones.EventTime,
	).MODEL(
		genMilestone,
	).RETURNING(
		table.InlayMilestones.ID,
		table.InlayMilestones.UUID,
		table.InlayMilestones.UpdatedAt,
		table.InlayMilestones.CreatedAt,
		table.InlayMilestones.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayMilestones
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	milestone.ID = int(dest.ID)
	milestone.UUID = dest.UUID.String()
	milestone.UpdatedAt = dest.UpdatedAt
	milestone.CreatedAt = dest.CreatedAt
	milestone.Version = int(dest.Version)

	return nil
}

func (m InlayMilestoneModel) GetByID(id int) (*InlayMilestone, bool, error) {
	query := postgres.SELECT(
		table.InlayMilestones.AllColumns,
	).FROM(
		table.InlayMilestones,
	).WHERE(
		table.InlayMilestones.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayMilestones
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayMilestoneFromGen(dest), true, nil
}

func (m InlayMilestoneModel) GetByUUID(uuidStr string) (*InlayMilestone, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InlayMilestones.AllColumns,
	).FROM(
		table.InlayMilestones,
	).WHERE(
		table.InlayMilestones.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayMilestones
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayMilestoneFromGen(dest), true, nil
}

func (m InlayMilestoneModel) GetByInlayID(inlayID int) ([]*InlayMilestone, error) {
	query := postgres.SELECT(
		table.InlayMilestones.AllColumns,
	).FROM(
		table.InlayMilestones,
	).WHERE(
		table.InlayMilestones.InlayID.EQ(postgres.Int(int64(inlayID))),
	).ORDER_BY(
		table.InlayMilestones.EventTime.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayMilestones
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	milestones := make([]*InlayMilestone, len(dest))
	for i, d := range dest {
		milestones[i] = inlayMilestoneFromGen(d)
	}

	return milestones, nil
}

func (m InlayMilestoneModel) GetAll() ([]*InlayMilestone, error) {
	query := postgres.SELECT(
		table.InlayMilestones.AllColumns,
	).FROM(
		table.InlayMilestones,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayMilestones
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	milestones := make([]*InlayMilestone, len(dest))
	for i, d := range dest {
		milestones[i] = inlayMilestoneFromGen(d)
	}

	return milestones, nil
}

func (m InlayMilestoneModel) Update(milestone *InlayMilestone) error {
	genMilestone, err := inlayMilestoneToGen(milestone)
	if err != nil {
		return err
	}

	query := table.InlayMilestones.UPDATE(
		table.InlayMilestones.Step,
		table.InlayMilestones.EventType,
		table.InlayMilestones.EventTime,
		table.InlayMilestones.Version,
	).MODEL(
		genMilestone,
	).WHERE(
		postgres.AND(
			table.InlayMilestones.ID.EQ(postgres.Int(int64(milestone.ID))),
			table.InlayMilestones.Version.EQ(postgres.Int(int64(milestone.Version))),
		),
	).RETURNING(
		table.InlayMilestones.UpdatedAt,
		table.InlayMilestones.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayMilestones
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	milestone.UpdatedAt = dest.UpdatedAt
	milestone.Version = int(dest.Version)

	return nil
}

func (m InlayMilestoneModel) Delete(id int) error {
	query := table.InlayMilestones.DELETE().WHERE(
		table.InlayMilestones.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
