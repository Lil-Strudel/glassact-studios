package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProofStatus string

type proofStatuses struct {
	Pending    ProofStatus
	Approved   ProofStatus
	Declined   ProofStatus
	Superseded ProofStatus
}

var ProofStatuses = proofStatuses{
	Pending:    ProofStatus("pending"),
	Approved:   ProofStatus("approved"),
	Declined:   ProofStatus("declined"),
	Superseded: ProofStatus("superseded"),
}

type InlayProof struct {
	StandardTable
	InlayID        int                    `json:"inlay_id"`
	VersionNumber  int                    `json:"version_number"`
	DesignAssetURL string                 `json:"design_asset_url"`
	Width          float64                `json:"width"`
	Height         float64                `json:"height"`
	PriceGroupID   *int                   `json:"price_group_id"`
	PriceCents     *int                   `json:"price_cents"`
	ScaleFactor    float64                `json:"scale_factor"`
	ColorOverrides map[string]interface{} `json:"color_overrides"`
	Status         ProofStatus            `json:"status"`
	ApprovedAt     *time.Time             `json:"approved_at"`
	ApprovedBy     *int                   `json:"approved_by"`
	DeclinedAt     *time.Time             `json:"declined_at"`
	DeclinedBy     *int                   `json:"declined_by"`
	DeclineReason  *string                `json:"decline_reason"`
	SentInChatID   int                    `json:"sent_in_chat_id"`
}

type InlayProofModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayProofFromGen(genProof model.InlayProofs) *InlayProof {
	var priceGroupID *int
	if genProof.PriceGroupID != nil {
		priceGroupIDVal := int(*genProof.PriceGroupID)
		priceGroupID = &priceGroupIDVal
	}

	var priceCents *int
	if genProof.PriceCents != nil {
		priceCentsVal := int(*genProof.PriceCents)
		priceCents = &priceCentsVal
	}

	var approvedBy *int
	if genProof.ApprovedBy != nil {
		approvedByVal := int(*genProof.ApprovedBy)
		approvedBy = &approvedByVal
	}

	var declinedBy *int
	if genProof.DeclinedBy != nil {
		declinedByVal := int(*genProof.DeclinedBy)
		declinedBy = &declinedByVal
	}

	var colorOverrides map[string]interface{}
	if genProof.ColorOverrides != "" {
		_ = json.Unmarshal([]byte(genProof.ColorOverrides), &colorOverrides)
	}

	proof := InlayProof{
		StandardTable: StandardTable{
			ID:        int(genProof.ID),
			UUID:      genProof.UUID.String(),
			CreatedAt: genProof.CreatedAt,
			UpdatedAt: genProof.UpdatedAt,
			Version:   int(genProof.Version),
		},
		InlayID:        int(genProof.InlayID),
		VersionNumber:  int(genProof.VersionNumber),
		DesignAssetURL: genProof.DesignAssetURL,
		Width:          genProof.Width,
		Height:         genProof.Height,
		PriceGroupID:   priceGroupID,
		PriceCents:     priceCents,
		ScaleFactor:    genProof.ScaleFactor,
		ColorOverrides: colorOverrides,
		Status:         ProofStatus(genProof.Status),
		ApprovedAt:     genProof.ApprovedAt,
		ApprovedBy:     approvedBy,
		DeclinedAt:     genProof.DeclinedAt,
		DeclinedBy:     declinedBy,
		DeclineReason:  genProof.DeclineReason,
		SentInChatID:   int(genProof.SentInChatID),
	}

	return &proof
}

func inlayProofToGen(ip *InlayProof) (*model.InlayProofs, error) {
	var proofUUID uuid.UUID
	var err error

	if ip.UUID != "" {
		proofUUID, err = uuid.Parse(ip.UUID)
		if err != nil {
			return nil, err
		}
	}

	var priceGroupID *int32
	if ip.PriceGroupID != nil {
		priceGroupIDVal := int32(*ip.PriceGroupID)
		priceGroupID = &priceGroupIDVal
	}

	var priceCents *int32
	if ip.PriceCents != nil {
		priceCentsVal := int32(*ip.PriceCents)
		priceCents = &priceCentsVal
	}

	var approvedBy *int32
	if ip.ApprovedBy != nil {
		approvedByVal := int32(*ip.ApprovedBy)
		approvedBy = &approvedByVal
	}

	var declinedBy *int32
	if ip.DeclinedBy != nil {
		declinedByVal := int32(*ip.DeclinedBy)
		declinedBy = &declinedByVal
	}

	colorOverridesStr := ""
	if ip.ColorOverrides != nil && len(ip.ColorOverrides) > 0 {
		colorOverridesBytes, _ := json.Marshal(ip.ColorOverrides)
		colorOverridesStr = string(colorOverridesBytes)
	}

	genProof := model.InlayProofs{
		ID:             int32(ip.ID),
		UUID:           proofUUID,
		InlayID:        int32(ip.InlayID),
		VersionNumber:  int32(ip.VersionNumber),
		DesignAssetURL: ip.DesignAssetURL,
		Width:          ip.Width,
		Height:         ip.Height,
		PriceGroupID:   priceGroupID,
		PriceCents:     priceCents,
		ScaleFactor:    ip.ScaleFactor,
		ColorOverrides: colorOverridesStr,
		Status:         string(ip.Status),
		ApprovedAt:     ip.ApprovedAt,
		ApprovedBy:     approvedBy,
		DeclinedAt:     ip.DeclinedAt,
		DeclinedBy:     declinedBy,
		DeclineReason:  ip.DeclineReason,
		SentInChatID:   int32(ip.SentInChatID),
		UpdatedAt:      ip.UpdatedAt,
		CreatedAt:      ip.CreatedAt,
		Version:        int32(ip.Version),
	}

	return &genProof, nil
}

func (m InlayProofModel) Insert(proof *InlayProof) error {
	genProof, err := inlayProofToGen(proof)
	if err != nil {
		return err
	}

	query := table.InlayProofs.INSERT(
		table.InlayProofs.InlayID,
		table.InlayProofs.VersionNumber,
		table.InlayProofs.DesignAssetURL,
		table.InlayProofs.Width,
		table.InlayProofs.Height,
		table.InlayProofs.PriceGroupID,
		table.InlayProofs.PriceCents,
		table.InlayProofs.ScaleFactor,
		table.InlayProofs.ColorOverrides,
		table.InlayProofs.Status,
		table.InlayProofs.SentInChatID,
	).MODEL(
		genProof,
	).RETURNING(
		table.InlayProofs.ID,
		table.InlayProofs.UUID,
		table.InlayProofs.UpdatedAt,
		table.InlayProofs.CreatedAt,
		table.InlayProofs.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayProofs
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	proof.ID = int(dest.ID)
	proof.UUID = dest.UUID.String()
	proof.UpdatedAt = dest.UpdatedAt
	proof.CreatedAt = dest.CreatedAt
	proof.Version = int(dest.Version)

	return nil
}

func (m InlayProofModel) GetByID(id int) (*InlayProof, bool, error) {
	query := postgres.SELECT(
		table.InlayProofs.AllColumns,
	).FROM(
		table.InlayProofs,
	).WHERE(
		table.InlayProofs.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayProofs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayProofFromGen(dest), true, nil
}

func (m InlayProofModel) GetByUUID(uuidStr string) (*InlayProof, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InlayProofs.AllColumns,
	).FROM(
		table.InlayProofs,
	).WHERE(
		table.InlayProofs.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayProofs
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayProofFromGen(dest), true, nil
}

func (m InlayProofModel) GetByInlayID(inlayID int) ([]*InlayProof, error) {
	query := postgres.SELECT(
		table.InlayProofs.AllColumns,
	).FROM(
		table.InlayProofs,
	).WHERE(
		table.InlayProofs.InlayID.EQ(postgres.Int(int64(inlayID))),
	).ORDER_BY(
		table.InlayProofs.VersionNumber.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayProofs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	proofs := make([]*InlayProof, len(dest))
	for i, d := range dest {
		proofs[i] = inlayProofFromGen(d)
	}

	return proofs, nil
}

func (m InlayProofModel) GetLatestByInlayID(inlayID int) (*InlayProof, bool, error) {
	query := postgres.SELECT(
		table.InlayProofs.AllColumns,
	).FROM(
		table.InlayProofs,
	).WHERE(
		table.InlayProofs.InlayID.EQ(postgres.Int(int64(inlayID))),
	).ORDER_BY(
		table.InlayProofs.VersionNumber.DESC(),
	).LIMIT(1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayProofs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayProofFromGen(dest), true, nil
}

func (m InlayProofModel) GetApprovedByInlayID(inlayID int) (*InlayProof, bool, error) {
	query := postgres.SELECT(
		table.InlayProofs.AllColumns,
	).FROM(
		table.InlayProofs,
	).WHERE(
		postgres.AND(
			table.InlayProofs.InlayID.EQ(postgres.Int(int64(inlayID))),
			table.InlayProofs.Status.EQ(postgres.String(string(ProofStatuses.Approved))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayProofs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayProofFromGen(dest), true, nil
}

func (m InlayProofModel) GetAll() ([]*InlayProof, error) {
	query := postgres.SELECT(
		table.InlayProofs.AllColumns,
	).FROM(
		table.InlayProofs,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayProofs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	proofs := make([]*InlayProof, len(dest))
	for i, d := range dest {
		proofs[i] = inlayProofFromGen(d)
	}

	return proofs, nil
}

func (m InlayProofModel) Update(proof *InlayProof) error {
	genProof, err := inlayProofToGen(proof)
	if err != nil {
		return err
	}

	query := table.InlayProofs.UPDATE(
		table.InlayProofs.Status,
		table.InlayProofs.ApprovedAt,
		table.InlayProofs.ApprovedBy,
		table.InlayProofs.DeclinedAt,
		table.InlayProofs.DeclinedBy,
		table.InlayProofs.DeclineReason,
		table.InlayProofs.Version,
	).MODEL(
		genProof,
	).WHERE(
		postgres.AND(
			table.InlayProofs.ID.EQ(postgres.Int(int64(proof.ID))),
			table.InlayProofs.Version.EQ(postgres.Int(int64(proof.Version))),
		),
	).RETURNING(
		table.InlayProofs.UpdatedAt,
		table.InlayProofs.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayProofs
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	proof.UpdatedAt = dest.UpdatedAt
	proof.Version = int(dest.Version)

	return nil
}

func (m InlayProofModel) Delete(id int) error {
	query := table.InlayProofs.DELETE().WHERE(
		table.InlayProofs.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
