package inlay

import (
	"fmt"
	"net/http"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// HandleRecustomizeInlay replaces the customization on an existing customized
// catalog inlay: the dealership re-enters the customizer, re-bakes, and the
// result becomes a new pending internal-authority proof. Without this the only
// way to change your mind is to delete the inlay and add it again, losing the
// installation-kit choice and the chat thread.
//
// The previous approval is discarded — a new coloring needs its own pricing
// review, so the inlay goes back to "not ready" until internal approves it.
func (m InlayModule) HandleRecustomizeInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		BakedDesignAssetURL string                 `json:"baked_design_asset_url" validate:"required"`
		ScaleFactor         float64                `json:"scale_factor" validate:"required,gt=0"`
		Width               float64                `json:"width" validate:"required,gt=0"`
		Height              float64                `json:"height" validate:"required,gt=0"`
		ColorOverrides      map[string]interface{} `json:"color_overrides"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	project, ok := m.validateInlayOwnership(w, r, inlay)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only re-customize inlays on draft projects, current status: %s", project.Status))
		return
	}

	if inlay.Type != data.InlayTypes.Catalog || !inlay.IsCustomized {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("only customized catalog inlays can be re-customized"))
		return
	}

	if inlay.CatalogInfo == nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("catalog inlay %d is missing its catalog info", inlay.ID))
		return
	}

	catalogItem, found, err := m.Db.CatalogItems.GetByID(inlay.CatalogInfo.CatalogItemID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, fmt.Errorf("catalog item %d not found", inlay.CatalogInfo.CatalogItemID))
		return
	}

	proofCount, err := m.Db.InlayProofs.CountByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	colorOverrides := map[string]interface{}{}
	if body.ColorOverrides != nil {
		colorOverrides = body.ColorOverrides
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	defaultPriceGroupID := catalogItem.DefaultPriceGroupID
	proof := data.InlayProof{
		InlayID:           inlay.ID,
		VersionNumber:     proofCount + 1,
		DesignAssetURL:    body.BakedDesignAssetURL,
		Width:             body.Width,
		Height:            body.Height,
		PriceGroupID:      &defaultPriceGroupID,
		ScaleFactor:       body.ScaleFactor,
		ColorOverrides:    colorOverrides,
		ApprovalAuthority: data.ProofApprovalAuthorities.Internal,
		Status:            data.ProofStatuses.Pending,
		SentInChatID:      nil,
	}

	if err := m.Db.InlayProofs.TxInsert(tx, &proof); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create re-customized proof for inlay %d: %w", inlay.ID, err))
		return
	}

	if err := m.Db.InlayProofs.TxSupersedePendingByInlayID(tx, inlay.ID, proof.ID); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to supersede pending proofs for inlay %d: %w", inlay.ID, err))
		return
	}

	inlay.PreviewURL = body.BakedDesignAssetURL
	inlay.ApprovedProofID = nil
	if err := m.Db.Inlays.TxUpdateFields(tx, inlay); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to reset inlay %d after re-customization: %w", inlay.ID, err))
		return
	}

	if err := tx.Commit(); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.SendNotificationToAllInternalUsers(
		data.NotificationEventTypes.InternalReviewRequired,
		fmt.Sprintf("Re-customized inlay needs review: %s", inlay.Name),
		fmt.Sprintf("The customization on %q (from catalog %s) changed and v%d is ready for internal pricing review.", inlay.Name, catalogItem.CatalogCode, proof.VersionNumber),
		&project.ID, &inlay.ID,
	)

	detail, err := m.buildInlayDetail(inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, detail)
}
