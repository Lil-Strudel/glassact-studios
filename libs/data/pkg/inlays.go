package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InlayType string

type inlayTypes struct {
	Catalog InlayType
	Custom  InlayType
}

var InlayTypes = inlayTypes{
	Catalog: InlayType("catalog"),
	Custom:  InlayType("custom"),
}

type Inlay struct {
	ID         int       `json:"id"`
	UUID       string    `json:"uuid"`
	ProjectID  int       `json:"project_id"`
	PreviewUrl string    `json:"preview_url"`
	Name       string    `json:"name"`
	PriceGroup int       `json:"price_group"`
	Type       InlayType `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Version    int       `json:"version"`
}

type InlayModel struct {
	DB *pgxpool.Pool
}

func (m InlayModel) Insert(inlay *Inlay) error {
	query := `
        INSERT INTO inlays () 
        VALUES ()
        RETURNING id, uuid, created_at, updated_at, version`

	args := []any{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&inlay.ID,
		&inlay.UUID,
		&inlay.CreatedAt,
		&inlay.UpdatedAt,
		&inlay.Version,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayModel) GetByID(id int) (*Inlay, bool, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, version
        FROM inlays
        WHERE id = $1`

	var inlay Inlay

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&inlay.ID,
		&inlay.UUID,
		&inlay.CreatedAt,
		&inlay.UpdatedAt,
		&inlay.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &inlay, true, nil
}

func (m InlayModel) GetByUUID(uuid string) (*Inlay, bool, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, version
        FROM inlays
        WHERE uuid = $1`

	var inlay Inlay

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(
		&inlay.ID,
		&inlay.UUID,
		&inlay.CreatedAt,
		&inlay.UpdatedAt,
		&inlay.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &inlay, true, nil
}

func (m InlayModel) GetAll() ([]*Inlay, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, version
		FROM inlays
		WHERE id IS NOT NULL;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	inlays := []*Inlay{}

	for rows.Next() {
		var inlay Inlay

		err := rows.Scan(

			&inlay.ID,
			&inlay.UUID,
			&inlay.CreatedAt,
			&inlay.UpdatedAt,
			&inlay.Version,
		)
		if err != nil {
			return nil, err
		}

		inlays = append(inlays, &inlay)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return inlays, nil
}

func (m InlayModel) Update(inlay *Inlay) error {
	query := `
        UPDATE inlays
		SET 
		WHERE id = $1 AND version = $2
        RETURNING version`

	args := []any{
		inlay.ID,
		inlay.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&inlay.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayModel) Delete(id int) error {
	query := `
        DELETE FROM inlays
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
