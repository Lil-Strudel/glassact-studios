package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InlayCatalogInfo struct {
	StandardTable
	InlayID       int `json:"inlay_id"`
	CatalogItemID int `json:"catalog_item_id"`
}

type InlayCustomInfo struct {
	StandardTable
	InlayID     int     `json:"inlay_id"`
	Description string  `json:"description"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
}

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
	StandardTable
	ProjectID   int               `json:"project_id"`
	Name        string            `json:"name"`
	PreviewURL  string            `json:"preview_url"`
	PriceGroup  int               `json:"price_group"`
	Type        InlayType         `json:"type"`
	CatalogInfo *InlayCatalogInfo `json:"catalog_info,omitempty"`
	CustomInfo  *InlayCustomInfo  `json:"custom_info,omitempty"`
}

type InlayModel struct {
	DB *pgxpool.Pool
}

func (m InlayModel) Insert(inlay *Inlay) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin(ctx)
	query := `
			INSERT INTO inlays (project_id, name, preview_url, price_group, type)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		inlay.ProjectID,
		inlay.Name,
		inlay.PreviewURL,
		inlay.PriceGroup,
		inlay.Type,
	}

	err = tx.QueryRow(ctx, query, args...).Scan(
		&inlay.ID,
		&inlay.UUID,
		&inlay.CreatedAt,
		&inlay.UpdatedAt,
		&inlay.Version,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if inlay.Type == InlayTypes.Catalog {
		if inlay.CatalogInfo == nil {
			return errors.New("CatalogInfo required when inserting Inlay with type catalog")
		}

		query := `
			INSERT INTO inlay_catalog_infos (inlay_id, catalog_item_id)
			VALUES ($1, $2)
			RETURNING id, uuid, created_at, updated_at, version`

		args := []any{
			inlay.ID,
			inlay.CatalogInfo.CatalogItemID,
		}

		err = tx.QueryRow(ctx, query, args...).Scan(
			&inlay.CatalogInfo.ID,
			&inlay.CatalogInfo.UUID,
			&inlay.CatalogInfo.CreatedAt,
			&inlay.CatalogInfo.UpdatedAt,
			&inlay.CatalogInfo.Version,
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	if inlay.Type == InlayTypes.Custom {
		if inlay.CustomInfo == nil {
			return errors.New("CustomInfo required when inserting Inlay with type custom")
		}

		query := `
			INSERT INTO inlay_custom_infos (inlay_id, description, width, height)
			VALUES ($1, $2, $3, $4)
			RETURNING id, uuid, created_at, updated_at, version`

		args := []any{
			inlay.ID,
			inlay.CustomInfo.Description,
			inlay.CustomInfo.Width,
			inlay.CustomInfo.Height,
		}

		err = tx.QueryRow(ctx, query, args...).Scan(
			&inlay.CustomInfo.ID,
			&inlay.CustomInfo.UUID,
			&inlay.CustomInfo.CreatedAt,
			&inlay.CustomInfo.UpdatedAt,
			&inlay.CustomInfo.Version,
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayModel) TxInsert(tx pgx.Tx, inlay *Inlay) error {
	query := `
			INSERT INTO inlays (project_id, name, preview_url, price_group, type)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		inlay.ProjectID,
		inlay.Name,
		inlay.PreviewURL,
		inlay.PriceGroup,
		inlay.Type,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRow(ctx, query, args...).Scan(
		&inlay.ID,
		&inlay.UUID,
		&inlay.CreatedAt,
		&inlay.UpdatedAt,
		&inlay.Version,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if inlay.Type == InlayTypes.Catalog {
		if inlay.CatalogInfo == nil {
			return errors.New("CatalogInfo required when inserting Inlay with type catalog")
		}

		query := `
			INSERT INTO inlay_catalog_infos (inlay_id, catalog_item_id)
			VALUES ($1, $2)
			RETURNING id, uuid, created_at, updated_at, version`

		args := []any{
			inlay.ID,
			inlay.CatalogInfo.CatalogItemID,
		}

		err = tx.QueryRow(ctx, query, args...).Scan(
			&inlay.CatalogInfo.ID,
			&inlay.CatalogInfo.UUID,
			&inlay.CatalogInfo.CreatedAt,
			&inlay.CatalogInfo.UpdatedAt,
			&inlay.CatalogInfo.Version,
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	if inlay.Type == InlayTypes.Custom {
		if inlay.CustomInfo == nil {
			return errors.New("CustomInfo required when inserting Inlay with type custom")
		}

		query := `
			INSERT INTO inlay_custom_infos (inlay_id, description, width, height)
			VALUES ($1, $2, $3, $4)
			RETURNING id, uuid, created_at, updated_at, version`

		args := []any{
			inlay.ID,
			inlay.CustomInfo.Description,
			inlay.CustomInfo.Width,
			inlay.CustomInfo.Height,
		}

		err = tx.QueryRow(ctx, query, args...).Scan(
			&inlay.CustomInfo.ID,
			&inlay.CustomInfo.UUID,
			&inlay.CustomInfo.CreatedAt,
			&inlay.CustomInfo.UpdatedAt,
			&inlay.CustomInfo.Version,
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	return nil
}
