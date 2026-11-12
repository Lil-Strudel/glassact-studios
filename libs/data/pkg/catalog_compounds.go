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
)

// CatalogCompoundMemberItem is the slice of a catalog item a compound needs to
// draw its members and to decide whether it can still be offered.
type CatalogCompoundMemberItem struct {
	ID            int     `json:"id"`
	UUID          string  `json:"uuid"`
	CatalogCode   string  `json:"catalog_code"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	SvgURL        string  `json:"svg_url"`
	DefaultWidth  float64 `json:"default_width"`
	DefaultHeight float64 `json:"default_height"`
	IsActive      bool    `json:"is_active"`
}

type CatalogCompoundMember struct {
	ID            int                        `json:"id"`
	CatalogItemID int                        `json:"catalog_item_id"`
	CenterX       float64                    `json:"center_x"`
	CenterY       float64                    `json:"center_y"`
	RotationDeg   float64                    `json:"rotation_deg"`
	SortOrder     int                        `json:"sort_order"`
	CatalogItem   *CatalogCompoundMemberItem `json:"catalog_item"`
}

type CatalogCompound struct {
	StandardTable
	Name         string                  `json:"name"`
	Description  *string                 `json:"description"`
	IsActive     bool                    `json:"is_active"`
	DisplayOrder *int                    `json:"display_order"`
	Members      []CatalogCompoundMember `json:"members"`
}

// IsOfferable is false when any member has been retired from the catalog, so a
// dealership is never handed a set it could only partly order.
func (c *CatalogCompound) IsOfferable() bool {
	if !c.IsActive {
		return false
	}
	for _, member := range c.Members {
		if member.CatalogItem == nil || !member.CatalogItem.IsActive {
			return false
		}
	}
	return true
}

type CatalogCompoundModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func catalogCompoundFromGen(gen model.CatalogCompounds) *CatalogCompound {
	compound := &CatalogCompound{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		Name:        gen.Name,
		Description: gen.Description,
		IsActive:    gen.IsActive,
		Members:     []CatalogCompoundMember{},
	}
	if gen.DisplayOrder != nil {
		order := int(*gen.DisplayOrder)
		compound.DisplayOrder = &order
	}
	return compound
}

func catalogCompoundToGen(c *CatalogCompound) (*model.CatalogCompounds, error) {
	var compoundUUID uuid.UUID
	if c.UUID != "" {
		parsed, err := uuid.Parse(c.UUID)
		if err != nil {
			return nil, err
		}
		compoundUUID = parsed
	}

	gen := &model.CatalogCompounds{
		ID:          int32(c.ID),
		UUID:        compoundUUID,
		Name:        c.Name,
		Description: c.Description,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Version:     int32(c.Version),
	}
	if c.DisplayOrder != nil {
		order := int32(*c.DisplayOrder)
		gen.DisplayOrder = &order
	}
	return gen, nil
}

func (m CatalogCompoundModel) Insert(compound *CatalogCompound) error {
	tx, err := m.STDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	gen, err := catalogCompoundToGen(compound)
	if err != nil {
		return err
	}

	query := table.CatalogCompounds.INSERT(
		table.CatalogCompounds.Name,
		table.CatalogCompounds.Description,
		table.CatalogCompounds.IsActive,
		table.CatalogCompounds.DisplayOrder,
	).MODEL(gen).RETURNING(
		table.CatalogCompounds.ID,
		table.CatalogCompounds.UUID,
		table.CatalogCompounds.CreatedAt,
		table.CatalogCompounds.UpdatedAt,
		table.CatalogCompounds.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dest model.CatalogCompounds
	if err := query.QueryContext(ctx, tx, &dest); err != nil {
		return err
	}

	if err := insertCompoundMembers(ctx, tx, int(dest.ID), compound.Members); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	compound.ID = int(dest.ID)
	compound.UUID = dest.UUID.String()
	compound.CreatedAt = dest.CreatedAt
	compound.UpdatedAt = dest.UpdatedAt
	compound.Version = int(dest.Version)
	return nil
}

// Update is version-locked on the compound row and replaces its members
// wholesale: the editor always saves the whole arrangement.
func (m CatalogCompoundModel) Update(compound *CatalogCompound) error {
	tx, err := m.STDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	gen, err := catalogCompoundToGen(compound)
	if err != nil {
		return err
	}

	query := table.CatalogCompounds.UPDATE(
		table.CatalogCompounds.Name,
		table.CatalogCompounds.Description,
		table.CatalogCompounds.IsActive,
		table.CatalogCompounds.DisplayOrder,
		table.CatalogCompounds.Version,
	).MODEL(gen).WHERE(
		postgres.AND(
			table.CatalogCompounds.ID.EQ(postgres.Int(int64(compound.ID))),
			table.CatalogCompounds.Version.EQ(postgres.Int(int64(compound.Version))),
		),
	).RETURNING(
		table.CatalogCompounds.UpdatedAt,
		table.CatalogCompounds.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dest model.CatalogCompounds
	if err := query.QueryContext(ctx, tx, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	_, err = table.CatalogCompoundMembers.DELETE().WHERE(
		table.CatalogCompoundMembers.CompoundID.EQ(postgres.Int(int64(compound.ID))),
	).ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to clear members of compound %d: %w", compound.ID, err)
	}

	if err := insertCompoundMembers(ctx, tx, compound.ID, compound.Members); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	compound.UpdatedAt = dest.UpdatedAt
	compound.Version = int(dest.Version)
	return nil
}

func insertCompoundMembers(ctx context.Context, tx *sql.Tx, compoundID int, members []CatalogCompoundMember) error {
	if len(members) == 0 {
		return nil
	}

	rows := make([]model.CatalogCompoundMembers, len(members))
	for i, member := range members {
		rows[i] = model.CatalogCompoundMembers{
			CompoundID:    int32(compoundID),
			CatalogItemID: int32(member.CatalogItemID),
			CenterX:       member.CenterX,
			CenterY:       member.CenterY,
			RotationDeg:   member.RotationDeg,
			SortOrder:     int32(i),
		}
	}

	_, err := table.CatalogCompoundMembers.INSERT(
		table.CatalogCompoundMembers.CompoundID,
		table.CatalogCompoundMembers.CatalogItemID,
		table.CatalogCompoundMembers.CenterX,
		table.CatalogCompoundMembers.CenterY,
		table.CatalogCompoundMembers.RotationDeg,
		table.CatalogCompoundMembers.SortOrder,
	).MODELS(rows).ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to insert members of compound %d: %w", compoundID, err)
	}
	return nil
}

func (m CatalogCompoundModel) GetByUUID(uuidStr string) (*CatalogCompound, bool, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}
	compounds, err := m.query(table.CatalogCompounds.UUID.EQ(postgres.UUID(parsed)))
	if err != nil {
		return nil, false, err
	}
	if len(compounds) == 0 {
		return nil, false, nil
	}
	return compounds[0], true, nil
}

func (m CatalogCompoundModel) GetAll() ([]*CatalogCompound, error) {
	return m.query(nil)
}

func (m CatalogCompoundModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := table.CatalogCompounds.DELETE().WHERE(
		table.CatalogCompounds.ID.EQ(postgres.Int(int64(id))),
	).ExecContext(ctx, m.STDB)
	return err
}

func (m CatalogCompoundModel) query(where postgres.BoolExpression) ([]*CatalogCompound, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := postgres.SELECT(table.CatalogCompounds.AllColumns).FROM(table.CatalogCompounds)
	if where != nil {
		stmt = stmt.WHERE(where)
	}
	stmt = stmt.ORDER_BY(
		table.CatalogCompounds.DisplayOrder.ASC().NULLS_LAST(),
		table.CatalogCompounds.Name.ASC(),
	)

	var dest []model.CatalogCompounds
	if err := stmt.QueryContext(ctx, m.STDB, &dest); err != nil {
		return nil, err
	}
	if len(dest) == 0 {
		return []*CatalogCompound{}, nil
	}

	compounds := make([]*CatalogCompound, len(dest))
	byID := make(map[int32]*CatalogCompound, len(dest))
	ids := make([]postgres.Expression, len(dest))
	for i, d := range dest {
		compounds[i] = catalogCompoundFromGen(d)
		byID[d.ID] = compounds[i]
		ids[i] = postgres.Int(int64(d.ID))
	}

	var memberRows []struct {
		model.CatalogCompoundMembers
		CatalogItem model.CatalogItems
	}
	err := postgres.SELECT(
		table.CatalogCompoundMembers.AllColumns,
		table.CatalogItems.ID,
		table.CatalogItems.UUID,
		table.CatalogItems.CatalogCode,
		table.CatalogItems.Name,
		table.CatalogItems.Category,
		table.CatalogItems.SvgURL,
		table.CatalogItems.DefaultWidth,
		table.CatalogItems.DefaultHeight,
		table.CatalogItems.IsActive,
	).FROM(
		table.CatalogCompoundMembers.
			INNER_JOIN(table.CatalogItems, table.CatalogItems.ID.EQ(table.CatalogCompoundMembers.CatalogItemID)),
	).WHERE(
		table.CatalogCompoundMembers.CompoundID.IN(ids...),
	).ORDER_BY(
		table.CatalogCompoundMembers.CompoundID,
		table.CatalogCompoundMembers.SortOrder,
	).QueryContext(ctx, m.STDB, &memberRows)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		return nil, fmt.Errorf("failed to load compound members: %w", err)
	}

	for _, row := range memberRows {
		compound := byID[row.CompoundID]
		compound.Members = append(compound.Members, CatalogCompoundMember{
			ID:            int(row.ID),
			CatalogItemID: int(row.CatalogItemID),
			CenterX:       row.CenterX,
			CenterY:       row.CenterY,
			RotationDeg:   row.RotationDeg,
			SortOrder:     int(row.SortOrder),
			CatalogItem: &CatalogCompoundMemberItem{
				ID:            int(row.CatalogItem.ID),
				UUID:          row.CatalogItem.UUID.String(),
				CatalogCode:   row.CatalogItem.CatalogCode,
				Name:          row.CatalogItem.Name,
				Category:      row.CatalogItem.Category,
				SvgURL:        row.CatalogItem.SvgURL,
				DefaultWidth:  row.CatalogItem.DefaultWidth,
				DefaultHeight: row.CatalogItem.DefaultHeight,
				IsActive:      row.CatalogItem.IsActive,
			},
		})
	}

	return compounds, nil
}
