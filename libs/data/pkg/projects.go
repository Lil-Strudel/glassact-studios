package data

import (
	"context"
	"errors"
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
	AwaitingPayment:   ProjectStatus("awaiting-payment"),
	Completed:         ProjectStatus("completed"),
}

type Project struct {
	ID           int           `json:"id"`
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Status       ProjectStatus `json:"status"`
	Approved     bool          `json:"approved"`
	DealershipID int           `json:"dealership_id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Version      int           `json:"version"`
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

func (m ProjectModel) GetAll() ([]*Project, error) {
	query := `
		SELECT id, uuid, name, status, approved, dealership_id, created_at, updated_at, version
		FROM projects
		WHERE id IS NOT NULL;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	projects := []*Project{}

	for rows.Next() {
		var project Project

		err := rows.Scan(

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
			return nil, err
		}

		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
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
