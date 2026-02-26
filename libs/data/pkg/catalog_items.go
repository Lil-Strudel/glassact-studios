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

type CatalogItem struct {
	StandardTable
	CatalogCode         string   `json:"catalog_code"`
	Name                string   `json:"name"`
	Description         *string  `json:"description"`
	Category            string   `json:"category"`
	DefaultWidth        float64  `json:"default_width"`
	DefaultHeight       float64  `json:"default_height"`
	MinWidth            float64  `json:"min_width"`
	MinHeight           float64  `json:"min_height"`
	DefaultPriceGroupID int      `json:"default_price_group_id"`
	SvgURL              string   `json:"svg_url"`
	IsActive            bool     `json:"is_active"`
	Tags                []string `json:"tags,omitempty"`
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

func catalogItemFromGen(genCatalogItem model.CatalogItems) *CatalogItem {
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
		IsActive:            genCatalogItem.IsActive,
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
		IsActive:            ci.IsActive,
		UpdatedAt:           ci.UpdatedAt,
		CreatedAt:           ci.CreatedAt,
		Version:             int32(ci.Version),
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
		table.CatalogItems.IsActive,
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
		table.CatalogItems.IsActive,
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
		table.CatalogItems.CreatedAt.DESC(),
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
	).DISTINCT()

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
