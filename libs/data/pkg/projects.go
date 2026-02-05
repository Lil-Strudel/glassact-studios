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

type ProjectStatus string

type projectStatuses struct {
	Draft           ProjectStatus
	Designing       ProjectStatus
	PendingApproval ProjectStatus
	Approved        ProjectStatus
	Ordered         ProjectStatus
	InProduction    ProjectStatus
	Shipped         ProjectStatus
	Delivered       ProjectStatus
	Invoiced        ProjectStatus
	Completed       ProjectStatus
	Cancelled       ProjectStatus
}

var ProjectStatuses = projectStatuses{
	Draft:           ProjectStatus("draft"),
	Designing:       ProjectStatus("designing"),
	PendingApproval: ProjectStatus("pending-approval"),
	Approved:        ProjectStatus("approved"),
	Ordered:         ProjectStatus("ordered"),
	InProduction:    ProjectStatus("in-production"),
	Shipped:         ProjectStatus("shipped"),
	Delivered:       ProjectStatus("delivered"),
	Invoiced:        ProjectStatus("invoiced"),
	Completed:       ProjectStatus("completed"),
	Cancelled:       ProjectStatus("cancelled"),
}

type Project struct {
	StandardTable
	DealershipID int           `json:"dealership_id"`
	Name         string        `json:"name"`
	Status       ProjectStatus `json:"status"`
	OrderedAt    *time.Time    `json:"ordered_at"`
	OrderedBy    *int          `json:"ordered_by"`
}

type ProjectModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func projectFromGen(genProj model.Projects) *Project {
	var orderedBy *int
	if genProj.OrderedBy != nil {
		orderedByVal := int(*genProj.OrderedBy)
		orderedBy = &orderedByVal
	}

	project := Project{
		StandardTable: StandardTable{
			ID:        int(genProj.ID),
			UUID:      genProj.UUID.String(),
			CreatedAt: genProj.CreatedAt,
			UpdatedAt: genProj.UpdatedAt,
			Version:   int(genProj.Version),
		},
		DealershipID: int(genProj.DealershipID),
		Name:         genProj.Name,
		Status:       ProjectStatus(genProj.Status),
		OrderedAt:    genProj.OrderedAt,
		OrderedBy:    orderedBy,
	}

	return &project
}

func projectToGen(p *Project) (*model.Projects, error) {
	var projectUUID uuid.UUID
	var err error

	if p.UUID != "" {
		projectUUID, err = uuid.Parse(p.UUID)
		if err != nil {
			return nil, err
		}
	}

	var orderedBy *int32
	if p.OrderedBy != nil {
		orderedByVal := int32(*p.OrderedBy)
		orderedBy = &orderedByVal
	}

	genProj := model.Projects{
		ID:           int32(p.ID),
		UUID:         projectUUID,
		DealershipID: int32(p.DealershipID),
		Name:         p.Name,
		Status:       string(p.Status),
		OrderedAt:    p.OrderedAt,
		OrderedBy:    orderedBy,
		UpdatedAt:    p.UpdatedAt,
		CreatedAt:    p.CreatedAt,
		Version:      int32(p.Version),
	}

	return &genProj, nil
}

func (m ProjectModel) Insert(project *Project) error {
	genProj, err := projectToGen(project)
	if err != nil {
		return err
	}

	query := table.Projects.INSERT(
		table.Projects.DealershipID,
		table.Projects.Name,
		table.Projects.Status,
		table.Projects.OrderedAt,
		table.Projects.OrderedBy,
	).MODEL(
		genProj,
	).RETURNING(
		table.Projects.ID,
		table.Projects.UUID,
		table.Projects.UpdatedAt,
		table.Projects.CreatedAt,
		table.Projects.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Projects
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	project.ID = int(dest.ID)
	project.UUID = dest.UUID.String()
	project.UpdatedAt = dest.UpdatedAt
	project.CreatedAt = dest.CreatedAt
	project.Version = int(dest.Version)

	return nil
}

func (m ProjectModel) TxInsert(tx *sql.Tx, project *Project) error {
	genProj, err := projectToGen(project)
	if err != nil {
		return err
	}

	query := table.Projects.INSERT(
		table.Projects.DealershipID,
		table.Projects.Name,
		table.Projects.Status,
		table.Projects.OrderedAt,
		table.Projects.OrderedBy,
	).MODEL(
		genProj,
	).RETURNING(
		table.Projects.ID,
		table.Projects.UUID,
		table.Projects.UpdatedAt,
		table.Projects.CreatedAt,
		table.Projects.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Projects
	err = query.QueryContext(ctx, tx, &dest)
	if err != nil {
		return err
	}

	project.ID = int(dest.ID)
	project.UUID = dest.UUID.String()
	project.UpdatedAt = dest.UpdatedAt
	project.CreatedAt = dest.CreatedAt
	project.Version = int(dest.Version)

	return nil
}

func (m ProjectModel) GetByID(id int) (*Project, bool, error) {
	query := postgres.SELECT(
		table.Projects.AllColumns,
	).FROM(
		table.Projects,
	).WHERE(
		table.Projects.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Projects
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return projectFromGen(dest), true, nil
}

func (m ProjectModel) GetByUUID(uuidStr string) (*Project, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.Projects.AllColumns,
	).FROM(
		table.Projects,
	).WHERE(
		table.Projects.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Projects
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return projectFromGen(dest), true, nil
}

func (m ProjectModel) GetByDealershipID(dealershipID int) ([]*Project, error) {
	query := postgres.SELECT(
		table.Projects.AllColumns,
	).FROM(
		table.Projects,
	).WHERE(
		table.Projects.DealershipID.EQ(postgres.Int(int64(dealershipID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Projects
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, len(dest))
	for i, d := range dest {
		projects[i] = projectFromGen(d)
	}

	return projects, nil
}

func (m ProjectModel) GetAll() ([]*Project, error) {
	query := postgres.SELECT(
		table.Projects.AllColumns,
	).FROM(
		table.Projects,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Projects
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, len(dest))
	for i, d := range dest {
		projects[i] = projectFromGen(d)
	}

	return projects, nil
}

func (m ProjectModel) Update(project *Project) error {
	genProj, err := projectToGen(project)
	if err != nil {
		return err
	}

	query := table.Projects.UPDATE(
		table.Projects.Name,
		table.Projects.Status,
		table.Projects.OrderedAt,
		table.Projects.OrderedBy,
		table.Projects.Version,
	).MODEL(
		genProj,
	).WHERE(
		postgres.AND(
			table.Projects.ID.EQ(postgres.Int(int64(project.ID))),
			table.Projects.Version.EQ(postgres.Int(int64(project.Version))),
		),
	).RETURNING(
		table.Projects.UpdatedAt,
		table.Projects.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Projects
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	project.UpdatedAt = dest.UpdatedAt
	project.Version = int(dest.Version)

	return nil
}

func (m ProjectModel) Delete(id int) error {
	query := table.Projects.DELETE().WHERE(
		table.Projects.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
