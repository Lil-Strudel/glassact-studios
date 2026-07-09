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
	Draft        ProjectStatus
	Ordered      ProjectStatus
	InProduction ProjectStatus
	Shipped      ProjectStatus
	Invoiced     ProjectStatus
	Completed    ProjectStatus
	Cancelled    ProjectStatus
}

var ProjectStatuses = projectStatuses{
	Draft:        ProjectStatus("draft"),
	Ordered:      ProjectStatus("ordered"),
	InProduction: ProjectStatus("in-production"),
	Shipped:      ProjectStatus("shipped"),
	Invoiced:     ProjectStatus("invoiced"),
	Completed:    ProjectStatus("completed"),
	Cancelled:    ProjectStatus("cancelled"),
}

type Project struct {
	StandardTable
	DealershipID      int           `json:"dealership_id"`
	Name              string        `json:"name"`
	InternalReference *string       `json:"internal_reference"`
	Status            ProjectStatus `json:"status"`
	TrackingNumber    *string       `json:"tracking_number"`
	OrderedAt         *time.Time    `json:"ordered_at"`
	OrderedBy         *int          `json:"ordered_by"`
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
		DealershipID:      int(genProj.DealershipID),
		Name:              genProj.Name,
		InternalReference: genProj.InternalReference,
		Status:            ProjectStatus(genProj.Status),
		TrackingNumber:    genProj.TrackingNumber,
		OrderedAt:         genProj.OrderedAt,
		OrderedBy:         orderedBy,
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
		ID:                int32(p.ID),
		UUID:              projectUUID,
		DealershipID:      int32(p.DealershipID),
		Name:              p.Name,
		InternalReference: p.InternalReference,
		Status:            string(p.Status),
		TrackingNumber:    p.TrackingNumber,
		OrderedAt:         p.OrderedAt,
		OrderedBy:         orderedBy,
		UpdatedAt:         p.UpdatedAt,
		CreatedAt:         p.CreatedAt,
		Version:           int32(p.Version),
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
		table.Projects.InternalReference,
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
		table.Projects.InternalReference,
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

// ProjectActionSummary captures the per-project counts that tell an internal
// user, at a glance, which projects need their attention.
type ProjectActionSummary struct {
	NeedsInternalApproval int `json:"needs_internal_approval"`
	NeedsProof            int `json:"needs_proof"`
	AwaitingReply         int `json:"awaiting_reply"`
}

// GetActionSummaries returns a map keyed by project ID of the internal
// action counts across all projects. It runs three set-based queries (no
// N+1): customized catalog inlays awaiting internal approval, custom inlays
// still needing a proof, and inlays whose most recent chat message came from
// the dealership (awaiting an internal reply). Projects with no outstanding
// action simply won't appear in the map.
func (m ProjectModel) GetActionSummaries() (map[int]ProjectActionSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	summaries := make(map[int]ProjectActionSummary)

	applyCounts := func(query string, set func(s *ProjectActionSummary, count int)) error {
		rows, err := m.DB.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var projectID, count int
			if err := rows.Scan(&projectID, &count); err != nil {
				return err
			}
			s := summaries[projectID]
			set(&s, count)
			summaries[projectID] = s
		}

		return rows.Err()
	}

	needsApproval := `
		SELECT i.project_id, COUNT(*)
		FROM inlays i
		WHERE i.type = 'catalog'
		  AND i.is_customized = true
		  AND i.approved_proof_id IS NULL
		  AND EXISTS (
		    SELECT 1 FROM inlay_proofs p
		    WHERE p.inlay_id = i.id
		      AND p.status = 'pending'
		      AND p.approval_authority = 'internal'
		  )
		GROUP BY i.project_id`
	if err := applyCounts(needsApproval, func(s *ProjectActionSummary, count int) {
		s.NeedsInternalApproval = count
	}); err != nil {
		return nil, err
	}

	needsProof := `
		SELECT i.project_id, COUNT(*)
		FROM inlays i
		WHERE i.type = 'custom'
		  AND i.approved_proof_id IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM inlay_proofs p
		    WHERE p.inlay_id = i.id
		      AND p.status = 'pending'
		  )
		GROUP BY i.project_id`
	if err := applyCounts(needsProof, func(s *ProjectActionSummary, count int) {
		s.NeedsProof = count
	}); err != nil {
		return nil, err
	}

	awaitingReply := `
		SELECT latest.project_id, COUNT(*)
		FROM (
		  SELECT DISTINCT ON (c.inlay_id) i.project_id, c.dealership_user_id
		  FROM inlay_chats c
		  JOIN inlays i ON i.id = c.inlay_id
		  ORDER BY c.inlay_id, c.created_at DESC, c.id DESC
		) latest
		WHERE latest.dealership_user_id IS NOT NULL
		GROUP BY latest.project_id`
	if err := applyCounts(awaitingReply, func(s *ProjectActionSummary, count int) {
		s.AwaitingReply = count
	}); err != nil {
		return nil, err
	}

	return summaries, nil
}

// GetDealershipNames returns a map keyed by project ID of the owning
// dealership's name for the given project IDs. Used to surface the dealership
// on the project list for internal users (who see projects across dealerships).
func (m ProjectModel) GetDealershipNames(projectIDs []int) (map[int]string, error) {
	names := make(map[int]string)
	if len(projectIDs) == 0 {
		return names, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, `
		SELECT p.id, d.name
		FROM projects p
		JOIN dealerships d ON d.id = p.dealership_id
		WHERE p.id = ANY($1)`, projectIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var projectID int
		var name string
		if err := rows.Scan(&projectID, &name); err != nil {
			return nil, err
		}
		names[projectID] = name
	}

	return names, rows.Err()
}

func (m ProjectModel) updateProject(ctx context.Context, executor qrm.Queryable, project *Project) error {
	genProj, err := projectToGen(project)
	if err != nil {
		return err
	}

	query := table.Projects.UPDATE(
		table.Projects.Name,
		table.Projects.InternalReference,
		table.Projects.Status,
		table.Projects.TrackingNumber,
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

	var dest model.Projects
	err = query.QueryContext(ctx, executor, &dest)
	if err != nil {
		return err
	}

	project.UpdatedAt = dest.UpdatedAt
	project.Version = int(dest.Version)

	return nil
}

func (m ProjectModel) Update(project *Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.updateProject(ctx, m.STDB, project)
}

func (m ProjectModel) TxUpdate(tx *sql.Tx, project *Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.updateProject(ctx, tx, project)
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
