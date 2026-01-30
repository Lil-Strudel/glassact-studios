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

type projectStatusi struct {
	AwaitingProof     ProjectStatus
	ProofInRevision   ProjectStatus
	AllProofsAccepted ProjectStatus
	Cancelled         ProjectStatus
	Ordered           ProjectStatus
	InProduction      ProjectStatus
	AwaitingInvoice   ProjectStatus
	AwaitingPayment   ProjectStatus
	Completed         ProjectStatus
}

var ProjectStatusi = projectStatusi{
	AwaitingProof:     ProjectStatus("awaiting-proof"),
	ProofInRevision:   ProjectStatus("proof-in-revision"),
	AllProofsAccepted: ProjectStatus("all-proofs-accepted"),
	Cancelled:         ProjectStatus("cancelled"),
	Ordered:           ProjectStatus("ordered"),
	InProduction:      ProjectStatus("in-production"),
	AwaitingInvoice:   ProjectStatus("awaiting-invoice"),
	AwaitingPayment:   ProjectStatus("awaiting-payment"),
	Completed:         ProjectStatus("completed"),
}

type Project struct {
	StandardTable
	Name         string        `json:"name"`
	Status       ProjectStatus `json:"status"`
	Approved     bool          `json:"approved"`
	DealershipID int           `json:"dealership_id"`
}

type ProjectModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func projectFromGen(genProj model.Projects) *Project {
	project := Project{
		StandardTable: StandardTable{
			ID:        int(genProj.ID),
			UUID:      genProj.UUID.String(),
			CreatedAt: genProj.CreatedAt,
			UpdatedAt: genProj.UpdatedAt,
			Version:   int(genProj.Version),
		},
		Name:         genProj.Name,
		Status:       ProjectStatus(genProj.Status),
		Approved:     genProj.Approved,
		DealershipID: int(genProj.DealershipID),
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

	genProj := model.Projects{
		ID:           int32(p.ID),
		UUID:         projectUUID,
		Name:         p.Name,
		Status:       string(p.Status),
		Approved:     p.Approved,
		DealershipID: int32(p.DealershipID),
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
		table.Projects.Name,
		table.Projects.Status,
		table.Projects.Approved,
		table.Projects.DealershipID,
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
		table.Projects.Name,
		table.Projects.Status,
		table.Projects.Approved,
		table.Projects.DealershipID,
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
		table.Projects.Approved,
		table.Projects.DealershipID,
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
