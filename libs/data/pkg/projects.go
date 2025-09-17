package data

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
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
	DB *pgxpool.Pool
}

func (m ProjectModel) Insert(project *Project) error {
	query := `
        INSERT INTO projects (name, status, approved, dealership_id) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		project.Name,
		project.Status,
		project.Approved,
		project.DealershipID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&project.ID,
		&project.UUID,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.Version,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m ProjectModel) TxInsert(tx pgx.Tx, project *Project) error {
	query := `
        INSERT INTO projects (name, status, approved, dealership_id) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		project.Name,
		project.Status,
		project.Approved,
		project.DealershipID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRow(ctx, query, args...).Scan(
		&project.ID,
		&project.UUID,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.Version,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}

func (m ProjectModel) GetByID(id int) (*Project, bool, error) {
	query := `
		SELECT id, uuid, name, status, approved, dealership_id, created_at, updated_at, version
        FROM projects
        WHERE id = $1`

	var project Project

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.UUID,
		&project.Name,
		&project.Status,
		&project.Approved,
		&project.DealershipID,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &project, true, nil
}

func (m ProjectModel) GetByUUID(uuid string) (*Project, bool, error) {
	query := `
		SELECT id, uuid, name, status, approved, dealership_id, created_at, updated_at, version
        FROM projects
        WHERE uuid = $1`

	var project Project

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(
		&project.ID,
		&project.UUID,
		&project.Name,
		&project.Status,
		&project.Approved,
		&project.DealershipID,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &project, true, nil
}

type ExpandedProject struct {
	Project
	Inlays *[]Inlay `json:"inlays,omitempty"`
}

func (m ProjectModel) GetAll(expand []string) ([]*ExpandedProject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	includeInlays := slices.Contains(expand, "inlays")

	selects := "projects.id, projects.uuid, projects.name, projects.status, projects.approved, projects.dealership_id, projects.updated_at, projects.created_at, projects.version"
	if includeInlays {
		selects += ", inlays.id, inlays.uuid, inlays.project_id, inlays.name, inlays.preview_url, inlays.price_group, inlays.type, inlays.updated_at, inlays.created_at, inlays.version"
	}

	joins := ""
	if includeInlays {
		joins += "LEFT JOIN inlays ON inlays.project_id = projects.id"
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM projects
		%s
		WHERE projects.id IS NOT NULL;`, selects, joins)

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projectMap := make(map[int]*ExpandedProject)
	for rows.Next() {
		var project Project
		var inlayID *int
		var inlayUUID *string
		var inlayProjectID *int
		var inlayName *string
		var inlayPreviewURL *string
		var inlayPriceGroup *int
		var inlayType *InlayType
		var inlayUpdatedAt *time.Time
		var inlayCreatedAt *time.Time
		var inlayVersion *int

		var scans []any
		scans = append(scans, &project.ID)
		scans = append(scans, &project.UUID)
		scans = append(scans, &project.Name)
		scans = append(scans, &project.Status)
		scans = append(scans, &project.Approved)
		scans = append(scans, &project.DealershipID)
		scans = append(scans, &project.UpdatedAt)
		scans = append(scans, &project.CreatedAt)
		scans = append(scans, &project.Version)

		if includeInlays {
			scans = append(scans, &inlayID)
			scans = append(scans, &inlayUUID)
			scans = append(scans, &inlayProjectID)
			scans = append(scans, &inlayName)
			scans = append(scans, &inlayPreviewURL)
			scans = append(scans, &inlayPriceGroup)
			scans = append(scans, &inlayType)
			scans = append(scans, &inlayUpdatedAt)
			scans = append(scans, &inlayCreatedAt)
			scans = append(scans, &inlayVersion)

		}

		if err := rows.Scan(scans...); err != nil {
			return nil, err
		}

		if _, exists := projectMap[project.ID]; !exists {
			pwi := ExpandedProject{
				Project: project,
			}

			if includeInlays {
				pwi.Inlays = &[]Inlay{}
			}

			projectMap[project.ID] = &pwi
		}

		if includeInlays && inlayID != nil {
			var inlay Inlay
			inlay.ID = *inlayID
			inlay.UUID = *inlayUUID
			inlay.ProjectID = *inlayProjectID
			inlay.Name = *inlayName
			inlay.PreviewURL = *inlayPreviewURL
			inlay.PriceGroup = *inlayPriceGroup
			inlay.Type = *inlayType
			inlay.UpdatedAt = *inlayUpdatedAt
			inlay.CreatedAt = *inlayCreatedAt
			inlay.Version = *inlayVersion

			*projectMap[project.ID].Inlays = append(*projectMap[project.ID].Inlays, inlay)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*ExpandedProject, 0, len(projectMap))
	for _, project := range projectMap {
		result = append(result, project)
	}

	return result, nil
}

func (m ProjectModel) Update(project *Project) error {
	query := `
        UPDATE projects
		SET name = $1, status = $2, approved = $3, dealership_id = $4
		WHERE id = $5 AND version = $6
        RETURNING version`

	args := []any{
		project.Name,
		project.Status,
		project.Approved,
		project.DealershipID,
		project.ID,
		project.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&project.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m ProjectModel) Delete(id int) error {
	query := `
        DELETE FROM projects
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
