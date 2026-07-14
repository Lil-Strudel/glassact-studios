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

type InlayCustomReferenceImage struct {
	ID                int    `json:"id"`
	UUID              string `json:"uuid"`
	InlayCustomInfoID int    `json:"inlay_custom_info_id"`
	ImageURL          string `json:"image_url"`
	SortOrder         int    `json:"sort_order"`
}

type InlayCustomInfo struct {
	StandardTable
	InlayID         int                         `json:"inlay_id"`
	Description     string                      `json:"description"`
	RequestedWidth  float64                     `json:"requested_width"`
	RequestedHeight float64                     `json:"requested_height"`
	ReferenceImages []InlayCustomReferenceImage `json:"reference_images"`
}

type Inlay struct {
	StandardTable
	ProjectID         int               `json:"project_id"`
	Name              string            `json:"name"`
	Type              InlayType         `json:"type"`
	IsCustomized      bool              `json:"is_customized"`
	InstallationKit   bool              `json:"installation_kit"`
	PreviewURL        string            `json:"preview_url"`
	SandblastFileURL  *string           `json:"sandblast_file_url"`
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
		ProjectID:       int(genInlay.ProjectID),
		Name:            genInlay.Name,
		Type:            InlayType(genInlay.Type),
		IsCustomized:    genInlay.IsCustomized,
		InstallationKit: genInlay.InstallationKit,
		PreviewURL:      genInlay.PreviewURL,
	}

	if genInlay.ApprovedProofID != nil {
		approvedProofID := int(*genInlay.ApprovedProofID)
		inlay.ApprovedProofID = &approvedProofID
	}

	if genInlay.ManufacturingStep != nil {
		inlay.ManufacturingStep = genInlay.ManufacturingStep
	}

	if genInlay.SandblastFileURL != nil {
		inlay.SandblastFileURL = genInlay.SandblastFileURL
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
		ID:              int32(in.ID),
		UUID:            inlayUUID,
		ProjectID:       int32(in.ProjectID),
		Name:            in.Name,
		Type:            string(in.Type),
		IsCustomized:    in.IsCustomized,
		InstallationKit: in.InstallationKit,
		PreviewURL:      in.PreviewURL,
		UpdatedAt:       in.UpdatedAt,
		CreatedAt:       in.CreatedAt,
		Version:         int32(in.Version),
	}

	if in.ApprovedProofID != nil {
		approvedProofID := int32(*in.ApprovedProofID)
		genInlay.ApprovedProofID = &approvedProofID
	}

	if in.ManufacturingStep != nil {
		genInlay.ManufacturingStep = in.ManufacturingStep
	}

	if in.SandblastFileURL != nil {
		genInlay.SandblastFileURL = in.SandblastFileURL
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

func referenceImageFromGen(gen *model.InlayCustomReferenceImages) *InlayCustomReferenceImage {
	return &InlayCustomReferenceImage{
		ID:                int(gen.ID),
		UUID:              gen.UUID.String(),
		InlayCustomInfoID: int(gen.InlayCustomInfoID),
		ImageURL:          gen.ImageURL,
		SortOrder:         int(gen.SortOrder),
	}
}

// insertReferenceImages bulk-inserts reference images for a custom info,
// assigning SortOrder from the slice index, and backfills the generated
// ID/UUID onto each element of images.
func (m InlayModel) insertReferenceImages(ctx context.Context, executor qrm.Queryable, customInfoID int, images []InlayCustomReferenceImage) error {
	if len(images) == 0 {
		return nil
	}

	genImages := make([]model.InlayCustomReferenceImages, len(images))
	for i := range images {
		genImages[i] = model.InlayCustomReferenceImages{
			InlayCustomInfoID: int32(customInfoID),
			ImageURL:          images[i].ImageURL,
			SortOrder:         int32(i),
		}
	}

	query := table.InlayCustomReferenceImages.INSERT(
		table.InlayCustomReferenceImages.InlayCustomInfoID,
		table.InlayCustomReferenceImages.ImageURL,
		table.InlayCustomReferenceImages.SortOrder,
	).MODELS(
		genImages,
	).RETURNING(
		table.InlayCustomReferenceImages.AllColumns,
	)

	var dest []model.InlayCustomReferenceImages
	err := query.QueryContext(ctx, executor, &dest)
	if err != nil {
		return err
	}

	for i := range dest {
		order := int(dest[i].SortOrder)
		if order >= 0 && order < len(images) {
			images[order] = *referenceImageFromGen(&dest[i])
		}
	}

	return nil
}

// getReferenceImages loads all reference images for a custom info, ordered by
// sort_order. Returns an empty (non-nil) slice when none exist.
func (m InlayModel) getReferenceImages(ctx context.Context, customInfoID int) ([]InlayCustomReferenceImage, error) {
	query := postgres.SELECT(
		table.InlayCustomReferenceImages.AllColumns,
	).FROM(
		table.InlayCustomReferenceImages,
	).WHERE(
		table.InlayCustomReferenceImages.InlayCustomInfoID.EQ(postgres.Int(int64(customInfoID))),
	).ORDER_BY(
		table.InlayCustomReferenceImages.SortOrder.ASC(),
	)

	var dest []model.InlayCustomReferenceImages
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	images := make([]InlayCustomReferenceImage, len(dest))
	for i := range dest {
		images[i] = *referenceImageFromGen(&dest[i])
	}

	return images, nil
}

// ReplaceReferenceImages atomically replaces the full set of reference images
// for a custom info with the given ordered list of URLs. Used when a dealership
// edits a draft custom inlay. The image files already live in S3; only the
// pointer rows are rewritten.
func (m InlayModel) ReplaceReferenceImages(customInfoID int, urls []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.STDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	delQuery := table.InlayCustomReferenceImages.DELETE().WHERE(
		table.InlayCustomReferenceImages.InlayCustomInfoID.EQ(postgres.Int(int64(customInfoID))),
	)
	_, err = delQuery.ExecContext(ctx, tx)
	if err != nil {
		return err
	}

	if len(urls) > 0 {
		images := make([]InlayCustomReferenceImage, len(urls))
		for i, url := range urls {
			images[i] = InlayCustomReferenceImage{ImageURL: url}
		}

		err = m.insertReferenceImages(ctx, tx, customInfoID, images)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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

		err = m.insertReferenceImages(ctx, executor, inlay.CustomInfo.ID, inlay.CustomInfo.ReferenceImages)
		if err != nil {
			return err
		}
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
		table.Inlays.IsCustomized,
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
		table.Inlays.IsCustomized,
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

	inlay := inlayFromGen(dest.Inlays, dest.InlayCatalogInfos, dest.InlayCustomInfos)

	if inlay.CustomInfo != nil {
		images, err := m.getReferenceImages(ctx, inlay.CustomInfo.ID)
		if err != nil {
			return nil, false, err
		}
		inlay.CustomInfo.ReferenceImages = images
	}

	return inlay, true, nil
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

	inlay := inlayFromGen(dest.Inlays, dest.InlayCatalogInfos, dest.InlayCustomInfos)

	if inlay.CustomInfo != nil {
		images, err := m.getReferenceImages(ctx, inlay.CustomInfo.ID)
		if err != nil {
			return nil, false, err
		}
		inlay.CustomInfo.ReferenceImages = images
	}

	return inlay, true, nil
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

// GetNeedingInternalApproval returns every customized catalog inlay across all
// projects that has an outstanding internal-authority proof awaiting approval.
// This powers the internal review queue.
func (m InlayModel) GetNeedingInternalApproval() ([]*Inlay, error) {
	pendingInternalProof := postgres.EXISTS(
		postgres.SELECT(table.InlayProofs.ID).FROM(table.InlayProofs).WHERE(
			postgres.AND(
				table.InlayProofs.InlayID.EQ(table.Inlays.ID),
				table.InlayProofs.Status.EQ(postgres.String(string(ProofStatuses.Pending))),
				table.InlayProofs.ApprovalAuthority.EQ(postgres.String(string(ProofApprovalAuthorities.Internal))),
			),
		),
	)

	return m.queryInlaysWithInfo(
		postgres.AND(
			table.Inlays.Type.EQ(postgres.String(string(InlayTypes.Catalog))),
			table.Inlays.IsCustomized.EQ(postgres.Bool(true)),
			table.Inlays.ApprovedProofID.IS_NULL(),
			pendingInternalProof,
		),
	)
}

// GetCustomNeedingProof returns every custom inlay across all projects that is
// not yet ready and has no pending proof — i.e. a designer still needs to
// create the first proof for it.
func (m InlayModel) GetCustomNeedingProof() ([]*Inlay, error) {
	pendingProof := postgres.EXISTS(
		postgres.SELECT(table.InlayProofs.ID).FROM(table.InlayProofs).WHERE(
			postgres.AND(
				table.InlayProofs.InlayID.EQ(table.Inlays.ID),
				table.InlayProofs.Status.EQ(postgres.String(string(ProofStatuses.Pending))),
			),
		),
	)

	return m.queryInlaysWithInfo(
		postgres.AND(
			table.Inlays.Type.EQ(postgres.String(string(InlayTypes.Custom))),
			table.Inlays.ApprovedProofID.IS_NULL(),
			postgres.NOT(pendingProof),
		),
	)
}

func (m InlayModel) queryInlaysWithInfo(where postgres.BoolExpression) ([]*Inlay, error) {
	query := postgres.SELECT(
		table.Inlays.AllColumns,
		table.InlayCatalogInfos.AllColumns,
		table.InlayCustomInfos.AllColumns,
	).FROM(
		table.Inlays.
			LEFT_JOIN(table.InlayCatalogInfos, table.InlayCatalogInfos.InlayID.EQ(table.Inlays.ID)).
			LEFT_JOIN(table.InlayCustomInfos, table.InlayCustomInfos.InlayID.EQ(table.Inlays.ID)),
	).WHERE(
		where,
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
		table.Inlays.InstallationKit,
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

func (m InlayModel) TxUpdateFields(tx *sql.Tx, inlay *Inlay) error {
	genInlay, err := inlayToGen(inlay)
	if err != nil {
		return err
	}

	query := table.Inlays.UPDATE(
		table.Inlays.IsCustomized,
		table.Inlays.PreviewURL,
		table.Inlays.ApprovedProofID,
		table.Inlays.ManufacturingStep,
		table.Inlays.Version,
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Inlays
	err = query.QueryContext(ctx, tx, &dest)
	if err != nil {
		return err
	}

	inlay.UpdatedAt = dest.UpdatedAt
	inlay.Version = int(dest.Version)

	return nil
}

func (m InlayModel) UpdateSandblastFile(inlay *Inlay) error {
	genInlay, err := inlayToGen(inlay)
	if err != nil {
		return err
	}

	query := table.Inlays.UPDATE(
		table.Inlays.SandblastFileURL,
		table.Inlays.Version,
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Inlays
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	inlay.UpdatedAt = dest.UpdatedAt
	inlay.Version = int(dest.Version)

	return nil
}

func (m InlayModel) CountByProjectID(projectID int) (int, error) {
	query := postgres.SELECT(
		postgres.COUNT(table.Inlays.ID),
	).FROM(
		table.Inlays,
	).WHERE(
		table.Inlays.ProjectID.EQ(postgres.Int(int64(projectID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest struct {
		Count int64
	}
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return 0, err
	}

	return int(dest.Count), nil
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
