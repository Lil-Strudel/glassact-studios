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

type ProofApprovalAuthority string

type proofApprovalAuthorities struct {
	Dealership ProofApprovalAuthority
	Internal   ProofApprovalAuthority
}

var ProofApprovalAuthorities = proofApprovalAuthorities{
	Dealership: ProofApprovalAuthority("dealership"),
	Internal:   ProofApprovalAuthority("internal"),
}

type InlayProof struct {
	StandardTable
	InlayID                    int                    `json:"inlay_id"`
	VersionNumber              int                    `json:"version_number"`
	DesignAssetURL             string                 `json:"design_asset_url"`
	Width                      float64                `json:"width"`
	Height                     float64                `json:"height"`
	PriceGroupID               *int                   `json:"price_group_id"`
	PriceCents                 *int                   `json:"price_cents"`
	ScaleFactor                float64                `json:"scale_factor"`
	ColorOverrides             map[string]interface{} `json:"color_overrides"`
	ApprovalAuthority          ProofApprovalAuthority `json:"approval_authority"`
	Status                     ProofStatus            `json:"status"`
	ApprovedAt                 *time.Time             `json:"approved_at"`
	ApprovedByDealershipUserID *int                   `json:"approved_by_dealership_user_id"`
	ApprovedByInternalUserID   *int                   `json:"approved_by_internal_user_id"`
	DeclinedAt                 *time.Time             `json:"declined_at"`
	DeclinedByDealershipUserID *int                   `json:"declined_by_dealership_user_id"`
	DeclinedByInternalUserID   *int                   `json:"declined_by_internal_user_id"`
	DeclineReason              *string                `json:"decline_reason"`
	SentInChatID               *int                   `json:"sent_in_chat_id"`
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

	var approvedByDealershipUserID *int
	if genProof.ApprovedByDealershipUserID != nil {
		v := int(*genProof.ApprovedByDealershipUserID)
		approvedByDealershipUserID = &v
	}

	var approvedByInternalUserID *int
	if genProof.ApprovedByInternalUserID != nil {
		v := int(*genProof.ApprovedByInternalUserID)
		approvedByInternalUserID = &v
	}

	var declinedByDealershipUserID *int
	if genProof.DeclinedByDealershipUserID != nil {
		v := int(*genProof.DeclinedByDealershipUserID)
		declinedByDealershipUserID = &v
	}

	var declinedByInternalUserID *int
	if genProof.DeclinedByInternalUserID != nil {
		v := int(*genProof.DeclinedByInternalUserID)
		declinedByInternalUserID = &v
	}

	var sentInChatID *int
	if genProof.SentInChatID != nil {
		v := int(*genProof.SentInChatID)
		sentInChatID = &v
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
		InlayID:                    int(genProof.InlayID),
		VersionNumber:              int(genProof.VersionNumber),
		DesignAssetURL:             genProof.DesignAssetURL,
		Width:                      genProof.Width,
		Height:                     genProof.Height,
		PriceGroupID:               priceGroupID,
		PriceCents:                 priceCents,
		ScaleFactor:                genProof.ScaleFactor,
		ColorOverrides:             colorOverrides,
		ApprovalAuthority:          ProofApprovalAuthority(genProof.ApprovalAuthority),
		Status:                     ProofStatus(genProof.Status),
		ApprovedAt:                 genProof.ApprovedAt,
		ApprovedByDealershipUserID: approvedByDealershipUserID,
		ApprovedByInternalUserID:   approvedByInternalUserID,
		DeclinedAt:                 genProof.DeclinedAt,
		DeclinedByDealershipUserID: declinedByDealershipUserID,
		DeclinedByInternalUserID:   declinedByInternalUserID,
		DeclineReason:              genProof.DeclineReason,
		SentInChatID:               sentInChatID,
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

	var approvedByDealershipUserID *int32
	if ip.ApprovedByDealershipUserID != nil {
		v := int32(*ip.ApprovedByDealershipUserID)
		approvedByDealershipUserID = &v
	}

	var approvedByInternalUserID *int32
	if ip.ApprovedByInternalUserID != nil {
		v := int32(*ip.ApprovedByInternalUserID)
		approvedByInternalUserID = &v
	}

	var declinedByDealershipUserID *int32
	if ip.DeclinedByDealershipUserID != nil {
		v := int32(*ip.DeclinedByDealershipUserID)
		declinedByDealershipUserID = &v
	}

	var declinedByInternalUserID *int32
	if ip.DeclinedByInternalUserID != nil {
		v := int32(*ip.DeclinedByInternalUserID)
		declinedByInternalUserID = &v
	}

	var sentInChatID *int32
	if ip.SentInChatID != nil {
		v := int32(*ip.SentInChatID)
		sentInChatID = &v
	}

	colorOverridesStr := "{}"
	if ip.ColorOverrides != nil {
		colorOverridesBytes, _ := json.Marshal(ip.ColorOverrides)
		colorOverridesStr = string(colorOverridesBytes)
	}

	authority := string(ip.ApprovalAuthority)
	if authority == "" {
		authority = string(ProofApprovalAuthorities.Dealership)
	}

	genProof := model.InlayProofs{
		ID:                         int32(ip.ID),
		UUID:                       proofUUID,
		InlayID:                    int32(ip.InlayID),
		VersionNumber:              int32(ip.VersionNumber),
		DesignAssetURL:             ip.DesignAssetURL,
		Width:                      ip.Width,
		Height:                     ip.Height,
		PriceGroupID:               priceGroupID,
		PriceCents:                 priceCents,
		ScaleFactor:                ip.ScaleFactor,
		ColorOverrides:             colorOverridesStr,
		ApprovalAuthority:          authority,
		Status:                     string(ip.Status),
		ApprovedAt:                 ip.ApprovedAt,
		ApprovedByDealershipUserID: approvedByDealershipUserID,
		ApprovedByInternalUserID:   approvedByInternalUserID,
		DeclinedAt:                 ip.DeclinedAt,
		DeclinedByDealershipUserID: declinedByDealershipUserID,
		DeclinedByInternalUserID:   declinedByInternalUserID,
		DeclineReason:              ip.DeclineReason,
		SentInChatID:               sentInChatID,
		UpdatedAt:                  ip.UpdatedAt,
		CreatedAt:                  ip.CreatedAt,
		Version:                    int32(ip.Version),
	}

	return &genProof, nil
}

func (m InlayProofModel) insertProof(ctx context.Context, executor qrm.Queryable, proof *InlayProof) error {
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
		table.InlayProofs.ApprovalAuthority,
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

	var dest model.InlayProofs
	err = query.QueryContext(ctx, executor, &dest)
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

func (m InlayProofModel) Insert(proof *InlayProof) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.insertProof(ctx, m.STDB, proof)
}

func (m InlayProofModel) TxInsert(tx *sql.Tx, proof *InlayProof) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.insertProof(ctx, tx, proof)
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

func (m InlayProofModel) updateProof(ctx context.Context, executor qrm.Queryable, proof *InlayProof) error {
	genProof, err := inlayProofToGen(proof)
	if err != nil {
		return err
	}

	query := table.InlayProofs.UPDATE(
		table.InlayProofs.PriceGroupID,
		table.InlayProofs.PriceCents,
		table.InlayProofs.Status,
		table.InlayProofs.ApprovedAt,
		table.InlayProofs.ApprovedByDealershipUserID,
		table.InlayProofs.ApprovedByInternalUserID,
		table.InlayProofs.DeclinedAt,
		table.InlayProofs.DeclinedByDealershipUserID,
		table.InlayProofs.DeclinedByInternalUserID,
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

	var dest model.InlayProofs
	err = query.QueryContext(ctx, executor, &dest)
	if err != nil {
		return err
	}

	proof.UpdatedAt = dest.UpdatedAt
	proof.Version = int(dest.Version)

	return nil
}

func (m InlayProofModel) Update(proof *InlayProof) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.updateProof(ctx, m.STDB, proof)
}

func (m InlayProofModel) TxUpdate(tx *sql.Tx, proof *InlayProof) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.updateProof(ctx, tx, proof)
}

func (m InlayProofModel) CountByInlayID(inlayID int) (int, error) {
	query := postgres.SELECT(
		postgres.COUNT(table.InlayProofs.ID),
	).FROM(
		table.InlayProofs,
	).WHERE(
		table.InlayProofs.InlayID.EQ(postgres.Int(int64(inlayID))),
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

func (m InlayProofModel) TxSupersedePendingByInlayID(tx *sql.Tx, inlayID int, excludeProofID int) error {
	query := table.InlayProofs.UPDATE(
		table.InlayProofs.Status,
	).SET(
		postgres.String(string(ProofStatuses.Superseded)),
	).WHERE(
		postgres.AND(
			table.InlayProofs.InlayID.EQ(postgres.Int(int64(inlayID))),
			table.InlayProofs.Status.EQ(postgres.String(string(ProofStatuses.Pending))),
			table.InlayProofs.ID.NOT_EQ(postgres.Int(int64(excludeProofID))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, tx)
	return err
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
