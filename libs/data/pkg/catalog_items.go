package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogItem struct {
	StandardTable
	CatalogCode         string                 `json:"catalog_code"`
	Name                string                 `json:"name"`
	Description         *string                `json:"description"`
	Category            string                 `json:"category"`
	DefaultWidth        float64                `json:"default_width"`
	DefaultHeight       float64                `json:"default_height"`
	MinWidth            float64                `json:"min_width"`
	MinHeight           float64                `json:"min_height"`
	DefaultPriceGroupID int                    `json:"default_price_group_id"`
	SvgURL              string                 `json:"svg_url"`
	Manifest            map[string]interface{} `json:"manifest"`
	IsActive            bool                   `json:"is_active"`
	DisplayOrder        *int                   `json:"display_order"`
	Tags                []string               `json:"tags,omitempty"`
}

type CatalogItemTag struct {
	ID            int       `json:"id"`
	CatalogItemID int       `json:"catalog_item_id"`
	Tag           string    `json:"tag"`
	CreatedAt     time.Time `json:"created_at"`
}

type CatalogItemModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

// Catalog lists share one order wherever they surface: manual best-seller rank
// first, then catalog code, then name.
func catalogItemOrder() []postgres.OrderByClause {
	return []postgres.OrderByClause{
		table.CatalogItems.DisplayOrder.ASC().NULLS_LAST(),
		table.CatalogItems.CatalogCode.ASC(),
		table.CatalogItems.Name.ASC(),
	}
}

func catalogItemFromGen(genCatalogItem model.CatalogItems) *CatalogItem {
	var manifest map[string]interface{}
	if genCatalogItem.Manifest != "" {
		_ = json.Unmarshal([]byte(genCatalogItem.Manifest), &manifest)
	}

	catalogItem := CatalogItem{
		StandardTable: StandardTable{
			ID:        int(genCatalogItem.ID),
			UUID:      genCatalogItem.UUID.String(),
			CreatedAt: genCatalogItem.CreatedAt,
			UpdatedAt: genCatalogItem.UpdatedAt,
			Version:   int(genCatalogItem.Version),
		},
		CatalogCode:         genCatalogItem.CatalogCode,
		Name:                genCatalogItem.Name,
		Description:         genCatalogItem.Description,
		Category:            genCatalogItem.Category,
		DefaultWidth:        genCatalogItem.DefaultWidth,
		DefaultHeight:       genCatalogItem.DefaultHeight,
		MinWidth:            genCatalogItem.MinWidth,
		MinHeight:           genCatalogItem.MinHeight,
		DefaultPriceGroupID: int(genCatalogItem.DefaultPriceGroupID),
		SvgURL:              genCatalogItem.SvgURL,
		Manifest:            manifest,
		IsActive:            genCatalogItem.IsActive,
	}

	if genCatalogItem.DisplayOrder != nil {
		displayOrder := int(*genCatalogItem.DisplayOrder)
		catalogItem.DisplayOrder = &displayOrder
	}

	return &catalogItem
}

func catalogItemToGen(ci *CatalogItem) (*model.CatalogItems, error) {
	var catalogItemUUID uuid.UUID
	var err error

	if ci.UUID != "" {
		catalogItemUUID, err = uuid.Parse(ci.UUID)
		if err != nil {
			return nil, err
		}
	}

	manifestStr := "{}"
	if ci.Manifest != nil {
		manifestBytes, err := json.Marshal(ci.Manifest)
		if err != nil {
			return nil, err
		}
		manifestStr = string(manifestBytes)
	}

	genCatalogItem := model.CatalogItems{
		ID:                  int32(ci.ID),
		UUID:                catalogItemUUID,
		CatalogCode:         ci.CatalogCode,
		Name:                ci.Name,
		Description:         ci.Description,
		Category:            ci.Category,
		DefaultWidth:        ci.DefaultWidth,
		DefaultHeight:       ci.DefaultHeight,
		MinWidth:            ci.MinWidth,
		MinHeight:           ci.MinHeight,
		DefaultPriceGroupID: int32(ci.DefaultPriceGroupID),
		SvgURL:              ci.SvgURL,
		Manifest:            manifestStr,
		IsActive:            ci.IsActive,
		UpdatedAt:           ci.UpdatedAt,
		CreatedAt:           ci.CreatedAt,
		Version:             int32(ci.Version),
	}

	if ci.DisplayOrder != nil {
		displayOrder := int32(*ci.DisplayOrder)
		genCatalogItem.DisplayOrder = &displayOrder
	}

	return &genCatalogItem, nil
}

func catalogItemTagFromGen(genTag model.CatalogItemTags) *CatalogItemTag {
	return &CatalogItemTag{
		ID:            int(genTag.ID),
		CatalogItemID: int(genTag.CatalogItemID),
		Tag:           genTag.Tag,
		CreatedAt:     genTag.CreatedAt,
	}
}

func (m CatalogItemModel) Insert(catalogItem *CatalogItem) error {
	genCatalogItem, err := catalogItemToGen(catalogItem)
	if err != nil {
		return err
	}

	query := table.CatalogItems.INSERT(
		table.CatalogItems.CatalogCode,
		table.CatalogItems.Name,
		table.CatalogItems.Description,
		table.CatalogItems.Category,
		table.CatalogItems.DefaultWidth,
		table.CatalogItems.DefaultHeight,
		table.CatalogItems.MinWidth,
		table.CatalogItems.MinHeight,
		table.CatalogItems.DefaultPriceGroupID,
		table.CatalogItems.SvgURL,
		table.CatalogItems.Manifest,
		table.CatalogItems.IsActive,
		table.CatalogItems.DisplayOrder,
	).MODEL(
		genCatalogItem,
	).RETURNING(
		table.CatalogItems.ID,
		table.CatalogItems.UUID,
		table.CatalogItems.UpdatedAt,
		table.CatalogItems.CreatedAt,
		table.CatalogItems.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.CatalogItems
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	catalogItem.ID = int(dest.ID)
	catalogItem.UUID = dest.UUID.String()
	catalogItem.UpdatedAt = dest.UpdatedAt
	catalogItem.CreatedAt = dest.CreatedAt
	catalogItem.Version = int(dest.Version)

	return nil
}

func (m CatalogItemModel) GetByID(id int) (*CatalogItem, bool, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.CatalogItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	catalogItem := catalogItemFromGen(dest)

	tags, err := m.GetTags(int(dest.ID))
	if err != nil {
		return nil, false, err
	}
	catalogItem.Tags = tags

	return catalogItem, true, nil
}

func (m CatalogItemModel) GetByUUID(uuidStr string) (*CatalogItem, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.CatalogItems
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	catalogItem := catalogItemFromGen(dest)

	tags, err := m.GetTags(int(dest.ID))
	if err != nil {
		return nil, false, err
	}
	catalogItem.Tags = tags

	return catalogItem, true, nil
}

func (m CatalogItemModel) GetByCatalogCode(catalogCode string) (*CatalogItem, bool, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.CatalogCode.EQ(postgres.String(catalogCode)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.CatalogItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	catalogItem := catalogItemFromGen(dest)

	tags, err := m.GetTags(int(dest.ID))
	if err != nil {
		return nil, false, err
	}
	catalogItem.Tags = tags

	return catalogItem, true, nil
}

func (m CatalogItemModel) GetAll() ([]*CatalogItem, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).ORDER_BY(
		catalogItemOrder()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.CatalogItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	catalogItems := make([]*CatalogItem, len(dest))
	for i, d := range dest {
		catalogItem := catalogItemFromGen(d)

		tags, err := m.GetTags(int(d.ID))
		if err != nil {
			return nil, err
		}
		catalogItem.Tags = tags

		catalogItems[i] = catalogItem
	}

	return catalogItems, nil
}

func (m CatalogItemModel) Update(catalogItem *CatalogItem) error {
	genCatalogItem, err := catalogItemToGen(catalogItem)
	if err != nil {
		return err
	}

	query := table.CatalogItems.UPDATE(
		table.CatalogItems.CatalogCode,
		table.CatalogItems.Name,
		table.CatalogItems.Description,
		table.CatalogItems.Category,
		table.CatalogItems.DefaultWidth,
		table.CatalogItems.DefaultHeight,
		table.CatalogItems.MinWidth,
		table.CatalogItems.MinHeight,
		table.CatalogItems.DefaultPriceGroupID,
		table.CatalogItems.SvgURL,
		table.CatalogItems.Manifest,
		table.CatalogItems.IsActive,
		table.CatalogItems.DisplayOrder,
		table.CatalogItems.Version,
	).MODEL(
		genCatalogItem,
	).WHERE(
		postgres.AND(
			table.CatalogItems.ID.EQ(postgres.Int(int64(catalogItem.ID))),
			table.CatalogItems.Version.EQ(postgres.Int(int64(catalogItem.Version))),
		),
	).RETURNING(
		table.CatalogItems.UpdatedAt,
		table.CatalogItems.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.CatalogItems
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	catalogItem.UpdatedAt = dest.UpdatedAt
	catalogItem.Version = int(dest.Version)

	return nil
}

func (m CatalogItemModel) Delete(id int) error {
	query := table.CatalogItems.DELETE().WHERE(
		table.CatalogItems.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}

// SetDisplayOrder replaces the entire best-seller ranking: every item not named
// in orderedUUIDs becomes unranked, and the named ones take positions 1..N.
//
// Written as raw SQL rather than per-row Update() calls because position is not
// state worth optimistic-locking on, and the version-bump trigger would make N
// round-trips both slow and spuriously conflict-prone.
func (m CatalogItemModel) SetDisplayOrder(orderedUUIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := m.STDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin display order transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"UPDATE catalog_items SET display_order = NULL WHERE display_order IS NOT NULL",
	); err != nil {
		return fmt.Errorf("failed to clear existing display order: %w", err)
	}

	if len(orderedUUIDs) > 0 {
		for i, u := range orderedUUIDs {
			if _, err := uuid.Parse(u); err != nil {
				return fmt.Errorf("invalid catalog item uuid %q at position %d: %w", u, i, err)
			}
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE catalog_items
			SET display_order = ranked.position
			FROM (
				SELECT ci.id, t.position
				FROM unnest($1::text[]) WITH ORDINALITY AS t(item_uuid, position)
				JOIN catalog_items ci ON ci.uuid = t.item_uuid::uuid
			) AS ranked
			WHERE catalog_items.id = ranked.id
		`, orderedUUIDs)
		if err != nil {
			return fmt.Errorf("failed to assign display order: %w", err)
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to read display order update count: %w", err)
		}
		if int(affected) != len(orderedUUIDs) {
			return fmt.Errorf("expected to rank %d catalog items but matched %d", len(orderedUUIDs), affected)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit display order: %w", err)
	}

	return nil
}

// GetRanked returns the items with a display_order set, in rank order.
func (m CatalogItemModel) GetRanked() ([]*CatalogItem, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.DisplayOrder.IS_NOT_NULL(),
	).ORDER_BY(
		catalogItemOrder()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.CatalogItems
	if err := query.QueryContext(ctx, m.STDB, &dest); err != nil {
		return nil, err
	}

	catalogItems := make([]*CatalogItem, len(dest))
	for i, d := range dest {
		catalogItems[i] = catalogItemFromGen(d)
	}

	return catalogItems, nil
}

func (m CatalogItemModel) AddTag(catalogItemID int, tag string) error {
	query := table.CatalogItemTags.INSERT(
		table.CatalogItemTags.CatalogItemID,
		table.CatalogItemTags.Tag,
	).VALUES(
		int32(catalogItemID),
		tag,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m CatalogItemModel) RemoveTag(catalogItemID int, tag string) error {
	query := table.CatalogItemTags.DELETE().WHERE(
		postgres.AND(
			table.CatalogItemTags.CatalogItemID.EQ(postgres.Int(int64(catalogItemID))),
			table.CatalogItemTags.Tag.EQ(postgres.String(tag)),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	return err
}

func (m CatalogItemModel) GetTags(catalogItemID int) ([]string, error) {
	query := postgres.SELECT(
		table.CatalogItemTags.Tag,
	).FROM(
		table.CatalogItemTags,
	).WHERE(
		table.CatalogItemTags.CatalogItemID.EQ(postgres.Int(int64(catalogItemID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.CatalogItemTags
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(dest))
	for i, d := range dest {
		tags[i] = d.Tag
	}

	return tags, nil
}

func (m CatalogItemModel) GetByTag(tag string) ([]*CatalogItem, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems.INNER_JOIN(
			table.CatalogItemTags,
			table.CatalogItemTags.CatalogItemID.EQ(table.CatalogItems.ID),
		),
	).WHERE(
		table.CatalogItemTags.Tag.EQ(postgres.String(tag)),
	).ORDER_BY(
		catalogItemOrder()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.CatalogItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	catalogItems := make([]*CatalogItem, len(dest))
	for i, d := range dest {
		catalogItem := catalogItemFromGen(d)

		tags, err := m.GetTags(int(d.ID))
		if err != nil {
			return nil, err
		}
		catalogItem.Tags = tags

		catalogItems[i] = catalogItem
	}

	return catalogItems, nil
}

func (m CatalogItemModel) GetByCategory(category string) ([]*CatalogItem, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.Category.EQ(postgres.String(category)),
	).ORDER_BY(
		catalogItemOrder()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.CatalogItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	catalogItems := make([]*CatalogItem, len(dest))
	for i, d := range dest {
		catalogItem := catalogItemFromGen(d)

		tags, err := m.GetTags(int(d.ID))
		if err != nil {
			return nil, err
		}
		catalogItem.Tags = tags

		catalogItems[i] = catalogItem
	}

	return catalogItems, nil
}

func (m CatalogItemModel) GetAllActive() ([]*CatalogItem, error) {
	query := postgres.SELECT(
		table.CatalogItems.AllColumns,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.IsActive.EQ(postgres.Bool(true)),
	).ORDER_BY(
		catalogItemOrder()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.CatalogItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	catalogItems := make([]*CatalogItem, len(dest))
	for i, d := range dest {
		catalogItem := catalogItemFromGen(d)

		tags, err := m.GetTags(int(d.ID))
		if err != nil {
			return nil, err
		}
		catalogItem.Tags = tags

		catalogItems[i] = catalogItem
	}

	return catalogItems, nil
}

func (m CatalogItemModel) GetCategories() ([]string, error) {
	query := postgres.SELECT(
		table.CatalogItems.Category,
	).FROM(
		table.CatalogItems,
	).WHERE(
		table.CatalogItems.IsActive.EQ(postgres.Bool(true)),
	).DISTINCT().ORDER_BY(
		table.CatalogItems.Category.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var categories []string
	err := query.QueryContext(ctx, m.STDB, &categories)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (m CatalogItemModel) GetAllTags() ([]string, error) {
	query := postgres.SELECT(
		table.CatalogItemTags.Tag,
	).FROM(
		table.CatalogItemTags.INNER_JOIN(
			table.CatalogItems,
			table.CatalogItemTags.CatalogItemID.EQ(table.CatalogItems.ID),
		),
	).WHERE(
		table.CatalogItems.IsActive.EQ(postgres.Bool(true)),
	).DISTINCT().ORDER_BY(
		table.CatalogItemTags.Tag.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var tags []string
	err := query.QueryContext(ctx, m.STDB, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
