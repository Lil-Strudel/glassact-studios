package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// RelatedCatalogItem is the slice of a catalog item a dealership needs to see a
// suggestion, add it as an inlay and draw it at its real size.
type RelatedCatalogItem struct {
	ID            int     `json:"id"`
	UUID          string  `json:"uuid"`
	CatalogCode   string  `json:"catalog_code"`
	Name          string  `json:"name"`
	SvgURL        string  `json:"svg_url"`
	DefaultWidth  float64 `json:"default_width"`
	DefaultHeight float64 `json:"default_height"`
	IsActive      bool    `json:"is_active"`
}

type SandblastBackground struct {
	StandardTable
	Code         string               `json:"code"`
	Name         string               `json:"name"`
	Description  *string              `json:"description"`
	SvgURL       string               `json:"svg_url"`
	Tags         []string             `json:"tags"`
	IsActive     bool                 `json:"is_active"`
	RelatedItems []RelatedCatalogItem `json:"related_items"`
}

// WithoutRetiredItems drops related items that have been disabled in the
// catalog, so a dealership is never offered an inlay it cannot add.
func (b *SandblastBackground) WithoutRetiredItems() *SandblastBackground {
	offered := *b
	offered.RelatedItems = make([]RelatedCatalogItem, 0, len(b.RelatedItems))
	for _, item := range b.RelatedItems {
		if item.IsActive {
			offered.RelatedItems = append(offered.RelatedItems, item)
		}
	}
	return &offered
}

type SandblastBackgroundModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func sandblastBackgroundFromGen(gen model.SandblastBackgrounds) *SandblastBackground {
	tags := []string(gen.Tags)
	if tags == nil {
		tags = []string{}
	}
	return &SandblastBackground{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		Code:         gen.Code,
		Name:         gen.Name,
		Description:  gen.Description,
		SvgURL:       gen.SvgURL,
		Tags:         tags,
		IsActive:     gen.IsActive,
		RelatedItems: []RelatedCatalogItem{},
	}
}

func sandblastBackgroundToGen(b *SandblastBackground) (*model.SandblastBackgrounds, error) {
	var backgroundUUID uuid.UUID
	if b.UUID != "" {
		parsed, err := uuid.Parse(b.UUID)
		if err != nil {
			return nil, err
		}
		backgroundUUID = parsed
	}

	tags := pq.StringArray(b.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}

	return &model.SandblastBackgrounds{
		ID:          int32(b.ID),
		UUID:        backgroundUUID,
		Code:        b.Code,
		Name:        b.Name,
		Description: b.Description,
		SvgURL:      b.SvgURL,
		Tags:        tags,
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		Version:     int32(b.Version),
	}, nil
}

// Insert stores the background and its related items, in the order given.
func (m SandblastBackgroundModel) Insert(background *SandblastBackground) error {
	tx, err := m.STDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	gen, err := sandblastBackgroundToGen(background)
	if err != nil {
		return err
	}

	query := table.SandblastBackgrounds.INSERT(
		table.SandblastBackgrounds.Code,
		table.SandblastBackgrounds.Name,
		table.SandblastBackgrounds.Description,
		table.SandblastBackgrounds.SvgURL,
		table.SandblastBackgrounds.Tags,
		table.SandblastBackgrounds.IsActive,
	).MODEL(gen).RETURNING(
		table.SandblastBackgrounds.ID,
		table.SandblastBackgrounds.UUID,
		table.SandblastBackgrounds.CreatedAt,
		table.SandblastBackgrounds.UpdatedAt,
		table.SandblastBackgrounds.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dest model.SandblastBackgrounds
	if err := query.QueryContext(ctx, tx, &dest); err != nil {
		return err
	}

	if err := insertRelatedItems(ctx, tx, int(dest.ID), background.RelatedItems); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	background.ID = int(dest.ID)
	background.UUID = dest.UUID.String()
	background.CreatedAt = dest.CreatedAt
	background.UpdatedAt = dest.UpdatedAt
	background.Version = int(dest.Version)
	return nil
}

// Update is version-locked and replaces the related items wholesale: the admin
// form always saves the whole list.
func (m SandblastBackgroundModel) Update(background *SandblastBackground) error {
	tx, err := m.STDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := updateBackgroundRow(ctx, tx, background); err != nil {
		return err
	}

	_, err = table.SandblastBackgroundRelatedItems.DELETE().WHERE(
		table.SandblastBackgroundRelatedItems.BackgroundID.EQ(postgres.Int(int64(background.ID))),
	).ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to clear related items of background %d: %w", background.ID, err)
	}

	if err := insertRelatedItems(ctx, tx, background.ID, background.RelatedItems); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateFields saves the background's own columns and leaves its related items
// alone, for toggles that do not round-trip the list.
func (m SandblastBackgroundModel) UpdateFields(background *SandblastBackground) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return updateBackgroundRow(ctx, m.STDB, background)
}

func updateBackgroundRow(ctx context.Context, db qrm.Queryable, background *SandblastBackground) error {
	gen, err := sandblastBackgroundToGen(background)
	if err != nil {
		return err
	}

	query := table.SandblastBackgrounds.UPDATE(
		table.SandblastBackgrounds.Code,
		table.SandblastBackgrounds.Name,
		table.SandblastBackgrounds.Description,
		table.SandblastBackgrounds.SvgURL,
		table.SandblastBackgrounds.Tags,
		table.SandblastBackgrounds.IsActive,
		table.SandblastBackgrounds.Version,
	).MODEL(gen).WHERE(
		postgres.AND(
			table.SandblastBackgrounds.ID.EQ(postgres.Int(int64(background.ID))),
			table.SandblastBackgrounds.Version.EQ(postgres.Int(int64(background.Version))),
		),
	).RETURNING(
		table.SandblastBackgrounds.UpdatedAt,
		table.SandblastBackgrounds.Version,
	)

	var dest model.SandblastBackgrounds
	if err := query.QueryContext(ctx, db, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	background.UpdatedAt = dest.UpdatedAt
	background.Version = int(dest.Version)
	return nil
}

func insertRelatedItems(ctx context.Context, tx *sql.Tx, backgroundID int, items []RelatedCatalogItem) error {
	if len(items) == 0 {
		return nil
	}

	rows := make([]model.SandblastBackgroundRelatedItems, len(items))
	for i, item := range items {
		rows[i] = model.SandblastBackgroundRelatedItems{
			BackgroundID:  int32(backgroundID),
			CatalogItemID: int32(item.ID),
			SortOrder:     int32(i),
		}
	}

	_, err := table.SandblastBackgroundRelatedItems.INSERT(
		table.SandblastBackgroundRelatedItems.BackgroundID,
		table.SandblastBackgroundRelatedItems.CatalogItemID,
		table.SandblastBackgroundRelatedItems.SortOrder,
	).MODELS(rows).ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to insert related items of background %d: %w", backgroundID, err)
	}
	return nil
}

func (m SandblastBackgroundModel) GetByUUID(uuidStr string) (*SandblastBackground, bool, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}
	backgrounds, err := m.query(table.SandblastBackgrounds.UUID.EQ(postgres.UUID(parsed)))
	if err != nil {
		return nil, false, err
	}
	if len(backgrounds) == 0 {
		return nil, false, nil
	}
	return backgrounds[0], true, nil
}

func (m SandblastBackgroundModel) GetAll() ([]*SandblastBackground, error) {
	return m.query(nil)
}

func (m SandblastBackgroundModel) GetAllActive() ([]*SandblastBackground, error) {
	return m.query(table.SandblastBackgrounds.IsActive.IS_TRUE())
}

func (m SandblastBackgroundModel) query(where postgres.BoolExpression) ([]*SandblastBackground, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := postgres.SELECT(table.SandblastBackgrounds.AllColumns).FROM(table.SandblastBackgrounds)
	if where != nil {
		stmt = stmt.WHERE(where)
	}
	stmt = stmt.ORDER_BY(table.SandblastBackgrounds.Code.ASC())

	var dest []model.SandblastBackgrounds
	if err := stmt.QueryContext(ctx, m.STDB, &dest); err != nil {
		return nil, err
	}
	if len(dest) == 0 {
		return []*SandblastBackground{}, nil
	}

	backgrounds := make([]*SandblastBackground, len(dest))
	byID := make(map[int32]*SandblastBackground, len(dest))
	ids := make([]postgres.Expression, len(dest))
	for i, d := range dest {
		backgrounds[i] = sandblastBackgroundFromGen(d)
		byID[d.ID] = backgrounds[i]
		ids[i] = postgres.Int(int64(d.ID))
	}

	var relatedRows []struct {
		model.SandblastBackgroundRelatedItems
		CatalogItem model.CatalogItems
	}
	err := postgres.SELECT(
		table.SandblastBackgroundRelatedItems.AllColumns,
		table.CatalogItems.ID,
		table.CatalogItems.UUID,
		table.CatalogItems.CatalogCode,
		table.CatalogItems.Name,
		table.CatalogItems.SvgURL,
		table.CatalogItems.DefaultWidth,
		table.CatalogItems.DefaultHeight,
		table.CatalogItems.IsActive,
	).FROM(
		table.SandblastBackgroundRelatedItems.
			INNER_JOIN(table.CatalogItems, table.CatalogItems.ID.EQ(table.SandblastBackgroundRelatedItems.CatalogItemID)),
	).WHERE(
		table.SandblastBackgroundRelatedItems.BackgroundID.IN(ids...),
	).ORDER_BY(
		table.SandblastBackgroundRelatedItems.BackgroundID,
		table.SandblastBackgroundRelatedItems.SortOrder,
	).QueryContext(ctx, m.STDB, &relatedRows)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		return nil, fmt.Errorf("failed to load background related items: %w", err)
	}

	for _, row := range relatedRows {
		background := byID[row.BackgroundID]
		background.RelatedItems = append(background.RelatedItems, RelatedCatalogItem{
			ID:            int(row.CatalogItem.ID),
			UUID:          row.CatalogItem.UUID.String(),
			CatalogCode:   row.CatalogItem.CatalogCode,
			Name:          row.CatalogItem.Name,
			SvgURL:        row.CatalogItem.SvgURL,
			DefaultWidth:  row.CatalogItem.DefaultWidth,
			DefaultHeight: row.CatalogItem.DefaultHeight,
			IsActive:      row.CatalogItem.IsActive,
		})
	}

	return backgrounds, nil
}
