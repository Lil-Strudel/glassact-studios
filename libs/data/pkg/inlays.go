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

type InlayType string

type inlayTypes struct {
	Catalog InlayType
	Custom  InlayType
}

var InlayTypes = inlayTypes{
	Catalog: InlayType("catalog"),
	Custom:  InlayType("custom"),
}

type InlayCatalogInfo struct {
	StandardTable
	InlayID            int    `json:"inlay_id"`
	CatalogItemID      int    `json:"catalog_item_id"`
	CustomizationNotes string `json:"customization_notes"`
}

type InlayCustomInfo struct {
	StandardTable
	InlayID         int     `json:"inlay_id"`
	Description     string  `json:"description"`
	RequestedWidth  float64 `json:"requested_width"`
	RequestedHeight float64 `json:"requested_height"`
}

type Inlay struct {
	StandardTable
	ProjectID         int               `json:"project_id"`
	Name              string            `json:"name"`
	Type              InlayType         `json:"type"`
	PreviewURL        string            `json:"preview_url"`
	ApprovedProofID   *int              `json:"approved_proof_id,omitempty"`
	ManufacturingStep *string           `json:"manufacturing_step,omitempty"`
	CatalogInfo       *InlayCatalogInfo `json:"catalog_info,omitempty"`
	CustomInfo        *InlayCustomInfo  `json:"custom_info,omitempty"`
}

type InlayModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayFromGen(genInlay model.Inlays, genCatalogInfo *model.InlayCatalogInfos, genCustomInfo *model.InlayCustomInfos) *Inlay {
	inlay := Inlay{
		StandardTable: StandardTable{
			ID:        int(genInlay.ID),
			UUID:      genInlay.UUID.String(),
			CreatedAt: genInlay.CreatedAt,
			UpdatedAt: genInlay.UpdatedAt,
			Version:   int(genInlay.Version),
		},
		ProjectID:  int(genInlay.ProjectID),
		Name:       genInlay.Name,
		Type:       InlayType(genInlay.Type),
		PreviewURL: genInlay.PreviewURL,
	}

	if genInlay.ApprovedProofID != nil {
		approvedProofID := int(*genInlay.ApprovedProofID)
		inlay.ApprovedProofID = &approvedProofID
	}

	if genInlay.ManufacturingStep != nil {
		inlay.ManufacturingStep = genInlay.ManufacturingStep
	}

	if genCatalogInfo != nil {
		inlay.CatalogInfo = &InlayCatalogInfo{
			StandardTable: StandardTable{
				ID:        int(genCatalogInfo.ID),
				UUID:      genCatalogInfo.UUID.String(),
				CreatedAt: genCatalogInfo.CreatedAt,
				UpdatedAt: genCatalogInfo.UpdatedAt,
				Version:   int(genCatalogInfo.Version),
			},
			InlayID:            int(genCatalogInfo.InlayID),
			CatalogItemID:      int(genCatalogInfo.CatalogItemID),
			CustomizationNotes: genCatalogInfo.CustomizationNotes,
		}
	}

	if genCustomInfo != nil {
		inlay.CustomInfo = &InlayCustomInfo{
			StandardTable: StandardTable{
				ID:        int(genCustomInfo.ID),
				UUID:      genCustomInfo.UUID.String(),
				CreatedAt: genCustomInfo.CreatedAt,
				UpdatedAt: genCustomInfo.UpdatedAt,
				Version:   int(genCustomInfo.Version),
			},
			InlayID:         int(genCustomInfo.InlayID),
			Description:     genCustomInfo.Description,
			RequestedWidth:  genCustomInfo.RequestedWidth,
			RequestedHeight: genCustomInfo.RequestedHeight,
		}
	}

	return &inlay
}

func inlayToGen(in *Inlay) (*model.Inlays, error) {
	var inlayUUID uuid.UUID
	var err error

	if in.UUID != "" {
		inlayUUID, err = uuid.Parse(in.UUID)
		if err != nil {
			return nil, err
		}
	}

	genInlay := model.Inlays{
		ID:         int32(in.ID),
		UUID:       inlayUUID,
		ProjectID:  int32(in.ProjectID),
		Name:       in.Name,
		Type:       string(in.Type),
		PreviewURL: in.PreviewURL,
		UpdatedAt:  in.UpdatedAt,
		CreatedAt:  in.CreatedAt,
		Version:    int32(in.Version),
	}

	if in.ApprovedProofID != nil {
		approvedProofID := int32(*in.ApprovedProofID)
		genInlay.ApprovedProofID = &approvedProofID
	}

	if in.ManufacturingStep != nil {
		genInlay.ManufacturingStep = in.ManufacturingStep
	}

	return &genInlay, nil
}

func catalogInfoToGen(ci *InlayCatalogInfo) (*model.InlayCatalogInfos, error) {
	var ciUUID uuid.UUID
	var err error

	if ci.UUID != "" {
		ciUUID, err = uuid.Parse(ci.UUID)
		if err != nil {
			return nil, err
		}
	}

	genCatalogInfo := model.InlayCatalogInfos{
		ID:                 int32(ci.ID),
		UUID:               ciUUID,
		InlayID:            int32(ci.InlayID),
		CatalogItemID:      int32(ci.CatalogItemID),
		CustomizationNotes: ci.CustomizationNotes,
		UpdatedAt:          ci.UpdatedAt,
		CreatedAt:          ci.CreatedAt,
		Version:            int32(ci.Version),
	}

	return &genCatalogInfo, nil
}

func customInfoToGen(ci *InlayCustomInfo) (*model.InlayCustomInfos, error) {
	var ciUUID uuid.UUID
	var err error

	if ci.UUID != "" {
		ciUUID, err = uuid.Parse(ci.UUID)
		if err != nil {
			return nil, err
		}
	}

	genCustomInfo := model.InlayCustomInfos{
		ID:              int32(ci.ID),
		UUID:            ciUUID,
		InlayID:         int32(ci.InlayID),
		Description:     ci.Description,
		RequestedWidth:  ci.RequestedWidth,
		RequestedHeight: ci.RequestedHeight,
		UpdatedAt:       ci.UpdatedAt,
		CreatedAt:       ci.CreatedAt,
		Version:         int32(ci.Version),
	}

	return &genCustomInfo, nil
}

func (m InlayModel) insertInlaySubtypes(ctx context.Context, executor qrm.Queryable, inlay *Inlay) error {
	if inlay.Type == InlayTypes.Catalog {
		if inlay.CatalogInfo == nil {
			return errors.New("CatalogInfo required when inserting Inlay with type catalog")
		}

		inlay.CatalogInfo.InlayID = inlay.ID
		genCatalogInfo, err := catalogInfoToGen(inlay.CatalogInfo)
		if err != nil {
			return err
		}

		catalogQuery := table.InlayCatalogInfos.INSERT(
			table.InlayCatalogInfos.InlayID,
			table.InlayCatalogInfos.CatalogItemID,
			table.InlayCatalogInfos.CustomizationNotes,
		).MODEL(
			genCatalogInfo,
		).RETURNING(
			table.InlayCatalogInfos.ID,
			table.InlayCatalogInfos.UUID,
			table.InlayCatalogInfos.UpdatedAt,
			table.InlayCatalogInfos.CreatedAt,
			table.InlayCatalogInfos.Version,
		)

		var catalogDest model.InlayCatalogInfos
		err = catalogQuery.QueryContext(ctx, executor, &catalogDest)
		if err != nil {
			return err
		}

		inlay.CatalogInfo.ID = int(catalogDest.ID)
		inlay.CatalogInfo.UUID = catalogDest.UUID.String()
		inlay.CatalogInfo.CreatedAt = catalogDest.CreatedAt
		inlay.CatalogInfo.UpdatedAt = catalogDest.UpdatedAt
		inlay.CatalogInfo.Version = int(catalogDest.Version)
	}

	if inlay.Type == InlayTypes.Custom {
		if inlay.CustomInfo == nil {
			return errors.New("CustomInfo required when inserting Inlay with type custom")
		}

		inlay.CustomInfo.InlayID = inlay.ID
		genCustomInfo, err := customInfoToGen(inlay.CustomInfo)
		if err != nil {
			return err
		}

		customQuery := table.InlayCustomInfos.INSERT(
			table.InlayCustomInfos.InlayID,
			table.InlayCustomInfos.Description,
			table.InlayCustomInfos.RequestedWidth,
			table.InlayCustomInfos.RequestedHeight,
		).MODEL(
			genCustomInfo,
		).RETURNING(
			table.InlayCustomInfos.ID,
			table.InlayCustomInfos.UUID,
			table.InlayCustomInfos.UpdatedAt,
			table.InlayCustomInfos.CreatedAt,
			table.InlayCustomInfos.Version,
		)

		var customDest model.InlayCustomInfos
		err = customQuery.QueryContext(ctx, executor, &customDest)
		if err != nil {
			return err
		}

		inlay.CustomInfo.ID = int(customDest.ID)
		inlay.CustomInfo.UUID = customDest.UUID.String()
		inlay.CustomInfo.CreatedAt = customDest.CreatedAt
		inlay.CustomInfo.UpdatedAt = customDest.UpdatedAt
		inlay.CustomInfo.Version = int(customDest.Version)
	}

	return nil
}

func (m InlayModel) Insert(inlay *Inlay) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	genInlay, err := inlayToGen(inlay)
	if err != nil {
		return err
	}

	query := table.Inlays.INSERT(
		table.Inlays.ProjectID,
		table.Inlays.Name,
		table.Inlays.Type,
		table.Inlays.PreviewURL,
	).MODEL(
		genInlay,
	).RETURNING(
		table.Inlays.ID,
		table.Inlays.UUID,
		table.Inlays.UpdatedAt,
		table.Inlays.CreatedAt,
		table.Inlays.Version,
	)

	var dest model.Inlays
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	inlay.ID = int(dest.ID)
	inlay.UUID = dest.UUID.String()
	inlay.CreatedAt = dest.CreatedAt
	inlay.UpdatedAt = dest.UpdatedAt
	inlay.Version = int(dest.Version)

	err = m.insertInlaySubtypes(ctx, m.STDB, inlay)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayModel) TxInsert(tx *sql.Tx, inlay *Inlay) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	genInlay, err := inlayToGen(inlay)
	if err != nil {
		return err
	}

	query := table.Inlays.INSERT(
		table.Inlays.ProjectID,
		table.Inlays.Name,
		table.Inlays.Type,
		table.Inlays.PreviewURL,
	).MODEL(
		genInlay,
	).RETURNING(
		table.Inlays.ID,
		table.Inlays.UUID,
		table.Inlays.UpdatedAt,
		table.Inlays.CreatedAt,
		table.Inlays.Version,
	)

	var dest model.Inlays
	err = query.QueryContext(ctx, tx, &dest)
	if err != nil {
		return err
	}

	inlay.ID = int(dest.ID)
	inlay.UUID = dest.UUID.String()
	inlay.CreatedAt = dest.CreatedAt
	inlay.UpdatedAt = dest.UpdatedAt
	inlay.Version = int(dest.Version)

	err = m.insertInlaySubtypes(ctx, tx, inlay)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayModel) GetByID(id int) (*Inlay, bool, error) {
	query := postgres.SELECT(
		table.Inlays.AllColumns,
		table.InlayCatalogInfos.AllColumns,
		table.InlayCustomInfos.AllColumns,
	).FROM(
		table.Inlays.
			LEFT_JOIN(table.InlayCatalogInfos, table.InlayCatalogInfos.InlayID.EQ(table.Inlays.ID)).
			LEFT_JOIN(table.InlayCustomInfos, table.InlayCustomInfos.InlayID.EQ(table.Inlays.ID)),
	).WHERE(
		table.Inlays.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest struct {
		model.Inlays
		InlayCatalogInfos *model.InlayCatalogInfos
		InlayCustomInfos  *model.InlayCustomInfos
	}
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayFromGen(dest.Inlays, dest.InlayCatalogInfos, dest.InlayCustomInfos), true, nil
}

func (m InlayModel) GetByUUID(uuidStr string) (*Inlay, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.Inlays.AllColumns,
		table.InlayCatalogInfos.AllColumns,
		table.InlayCustomInfos.AllColumns,
	).FROM(
		table.Inlays.
			LEFT_JOIN(table.InlayCatalogInfos, table.InlayCatalogInfos.InlayID.EQ(table.Inlays.ID)).
			LEFT_JOIN(table.InlayCustomInfos, table.InlayCustomInfos.InlayID.EQ(table.Inlays.ID)),
	).WHERE(
		table.Inlays.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest struct {
		model.Inlays
		InlayCatalogInfos *model.InlayCatalogInfos
		InlayCustomInfos  *model.InlayCustomInfos
	}
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayFromGen(dest.Inlays, dest.InlayCatalogInfos, dest.InlayCustomInfos), true, nil
}

func (m InlayModel) GetAll() ([]*Inlay, error) {
	query := postgres.SELECT(
		table.Inlays.AllColumns,
		table.InlayCatalogInfos.AllColumns,
		table.InlayCustomInfos.AllColumns,
	).FROM(
		table.Inlays.
			LEFT_JOIN(table.InlayCatalogInfos, table.InlayCatalogInfos.InlayID.EQ(table.Inlays.ID)).
			LEFT_JOIN(table.InlayCustomInfos, table.InlayCustomInfos.InlayID.EQ(table.Inlays.ID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []struct {
		model.Inlays
		InlayCatalogInfos *model.InlayCatalogInfos
		InlayCustomInfos  *model.InlayCustomInfos
	}
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	inlays := make([]*Inlay, len(dest))
	for i, d := range dest {
		inlays[i] = inlayFromGen(d.Inlays, d.InlayCatalogInfos, d.InlayCustomInfos)
	}

	return inlays, nil
}

func (m InlayModel) GetByProjectID(projectID int) ([]*Inlay, error) {
	query := postgres.SELECT(
		table.Inlays.AllColumns,
		table.InlayCatalogInfos.AllColumns,
		table.InlayCustomInfos.AllColumns,
	).FROM(
		table.Inlays.
			LEFT_JOIN(table.InlayCatalogInfos, table.InlayCatalogInfos.InlayID.EQ(table.Inlays.ID)).
			LEFT_JOIN(table.InlayCustomInfos, table.InlayCustomInfos.InlayID.EQ(table.Inlays.ID)),
	).WHERE(
		table.Inlays.ProjectID.EQ(postgres.Int(int64(projectID))),
	).ORDER_BY(
		table.Inlays.CreatedAt.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []struct {
		model.Inlays
		InlayCatalogInfos *model.InlayCatalogInfos
		InlayCustomInfos  *model.InlayCustomInfos
	}
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	inlays := make([]*Inlay, len(dest))
	for i, d := range dest {
		inlays[i] = inlayFromGen(d.Inlays, d.InlayCatalogInfos, d.InlayCustomInfos)
	}

	return inlays, nil
}

func (m InlayModel) Update(inlay *Inlay) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	genInlay, err := inlayToGen(inlay)
	if err != nil {
		return err
	}

	query := table.Inlays.UPDATE(
		table.Inlays.ProjectID,
		table.Inlays.Name,
		table.Inlays.Type,
		table.Inlays.PreviewURL,
	).MODEL(
		genInlay,
	).WHERE(
		postgres.AND(
			table.Inlays.ID.EQ(postgres.Int(int64(inlay.ID))),
			table.Inlays.Version.EQ(postgres.Int(int64(inlay.Version))),
		),
	).RETURNING(
		table.Inlays.UpdatedAt,
		table.Inlays.Version,
	)

	var dest model.Inlays
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	inlay.UpdatedAt = dest.UpdatedAt
	inlay.Version = int(dest.Version)

	if inlay.Type == InlayTypes.Catalog && inlay.CatalogInfo != nil {
		genCatalogInfo, err := catalogInfoToGen(inlay.CatalogInfo)
		if err != nil {
			return err
		}

		catalogQuery := table.InlayCatalogInfos.UPDATE(
			table.InlayCatalogInfos.CatalogItemID,
			table.InlayCatalogInfos.CustomizationNotes,
		).MODEL(
			genCatalogInfo,
		).WHERE(
			postgres.AND(
				table.InlayCatalogInfos.InlayID.EQ(postgres.Int(int64(inlay.ID))),
				table.InlayCatalogInfos.Version.EQ(postgres.Int(int64(inlay.CatalogInfo.Version))),
			),
		).RETURNING(
			table.InlayCatalogInfos.UpdatedAt,
			table.InlayCatalogInfos.Version,
		)

		var catalogDest model.InlayCatalogInfos
		err = catalogQuery.QueryContext(ctx, m.STDB, &catalogDest)
		if err != nil {
			return err
		}

		inlay.CatalogInfo.UpdatedAt = catalogDest.UpdatedAt
		inlay.CatalogInfo.Version = int(catalogDest.Version)
	}

	if inlay.Type == InlayTypes.Custom && inlay.CustomInfo != nil {
		genCustomInfo, err := customInfoToGen(inlay.CustomInfo)
		if err != nil {
			return err
		}

		customQuery := table.InlayCustomInfos.UPDATE(
			table.InlayCustomInfos.Description,
			table.InlayCustomInfos.RequestedWidth,
			table.InlayCustomInfos.RequestedHeight,
		).MODEL(
			genCustomInfo,
		).WHERE(
			postgres.AND(
				table.InlayCustomInfos.InlayID.EQ(postgres.Int(int64(inlay.ID))),
				table.InlayCustomInfos.Version.EQ(postgres.Int(int64(inlay.CustomInfo.Version))),
			),
		).RETURNING(
			table.InlayCustomInfos.UpdatedAt,
			table.InlayCustomInfos.Version,
		)

		var customDest model.InlayCustomInfos
		err = customQuery.QueryContext(ctx, m.STDB, &customDest)
		if err != nil {
			return err
		}

		inlay.CustomInfo.UpdatedAt = customDest.UpdatedAt
		inlay.CustomInfo.Version = int(customDest.Version)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayModel) Delete(id int) error {
	query := table.Inlays.DELETE().WHERE(
		table.Inlays.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
