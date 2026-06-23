package proof

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type ProofModule struct {
	*app.Application
}

func NewProofModule(app *app.Application) *ProofModule {
	return &ProofModule{app}
}

func (m ProofModule) getInlayWithAccessCheck(w http.ResponseWriter, r *http.Request) (*data.Inlay, *data.Project, bool) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return nil, nil, false
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, nil, false
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, nil, false
	}

	project, found, err := m.Db.Projects.GetByID(inlay.ProjectID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, nil, false
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, nil, false
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return nil, nil, false
		}
	}

	return inlay, project, true
}

func (m ProofModule) HandleGetProofsByInlay(w http.ResponseWriter, r *http.Request) {
	inlay, _, ok := m.getInlayWithAccessCheck(w, r)
	if !ok {
		return
	}

	proofs, err := m.Db.InlayProofs.GetByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, proofs)
}

func (m ProofModule) HandleGetProof(w http.ResponseWriter, r *http.Request) {
	proofUUID := r.PathValue("uuid")

	err := m.Validate.Var(proofUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	proof, found, err := m.Db.InlayProofs.GetByUUID(proofUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByID(proof.InlayID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	project, found, err := m.Db.Projects.GetByID(inlay.ProjectID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	m.WriteJSON(w, r, http.StatusOK, proof)
}

// HandleCreateProof is used by internal designers to upload a proof for a
// custom inlay. It always creates a dealership-authority proof (the dealership
// then approves or declines it). Customizer-baked proofs for catalog inlays
// are created inside inlay.HandlePostCatalogInlay, not here.
func (m ProofModule) HandleCreateProof(w http.ResponseWriter, r *http.Request) {
	inlay, project, ok := m.getInlayWithAccessCheck(w, r)
	if !ok {
		return
	}

	var body struct {
		DesignAssetURL       string                 `json:"design_asset_url" validate:"required"`
		Width                float64                `json:"width" validate:"required,gt=0"`
		Height               float64                `json:"height" validate:"required,gt=0"`
		PriceGroupID         *int                   `json:"price_group_id"`
		PriceAdjustmentType  *string                `json:"price_adjustment_type" validate:"omitempty,oneof=none percent fixed"`
		PriceAdjustmentValue *float64               `json:"price_adjustment_value"`
		ScaleFactor          *float64               `json:"scale_factor"`
		ColorOverrides       map[string]interface{} `json:"color_overrides"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user := m.ContextGetUser(r)
	if !user.IsInternal() {
		m.WriteError(w, r, m.Err.Forbidden, fmt.Errorf("only internal users can create proofs"))
		return
	}

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only create proofs for draft projects, currently: %s", project.Status))
		return
	}

	proofCount, err := m.Db.InlayProofs.CountByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	versionNumber := proofCount + 1

	scaleFactor := 1.0
	if body.ScaleFactor != nil {
		scaleFactor = *body.ScaleFactor
	}

	adjustmentType := data.PriceAdjustmentTypes.None
	if body.PriceAdjustmentType != nil {
		adjustmentType = data.PriceAdjustmentType(*body.PriceAdjustmentType)
	}
	adjustmentValue := 0.0
	if body.PriceAdjustmentValue != nil {
		adjustmentValue = *body.PriceAdjustmentValue
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

	internalUserID := user.GetID()
	chatMessage := data.InlayChat{
		InlayID:        inlay.ID,
		InternalUserID: &internalUserID,
		MessageType:    data.ChatMessageTypes.ProofSent,
		Message:        fmt.Sprintf("Proof v%d ready for review", versionNumber),
	}

	err = m.Db.InlayChats.TxInsert(tx, &chatMessage)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create chat message for proof: %w", err))
		return
	}

	chatID := chatMessage.ID
	proof := data.InlayProof{
		InlayID:              inlay.ID,
		VersionNumber:        versionNumber,
		DesignAssetURL:       body.DesignAssetURL,
		Width:                body.Width,
		Height:               body.Height,
		PriceGroupID:         body.PriceGroupID,
		PriceAdjustmentType:  adjustmentType,
		PriceAdjustmentValue: adjustmentValue,
		ScaleFactor:          scaleFactor,
		ColorOverrides:       colorOverrides,
		ApprovalAuthority:    data.ProofApprovalAuthorities.Dealership,
		Status:               data.ProofStatuses.Pending,
		SentInChatID:         &chatID,
	}

	err = m.Db.InlayProofs.TxInsert(tx, &proof)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create proof: %w", err))
		return
	}

	err = m.Db.InlayProofs.TxSupersedePendingByInlayID(tx, inlay.ID, proof.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to supersede pending proofs: %w", err))
		return
	}

	inlay.PreviewURL = body.DesignAssetURL
	err = m.Db.Inlays.TxUpdateFields(tx, inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update inlay preview: %w", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.SendNotificationToAllDealershipUsersForProject(
		project.ID,
		data.NotificationEventTypes.ProofReady,
		fmt.Sprintf("Proof ready for review: %s", inlay.Name),
		fmt.Sprintf("A new proof (v%d) is ready for review on inlay %q.", versionNumber, inlay.Name),
		&inlay.ID,
	)

	m.WriteJSON(w, r, http.StatusCreated, proof)
}

// loadProofWithContext fetches a proof and its surrounding inlay + project,
// running the dealership scope check before returning.
func (m ProofModule) loadProofWithContext(w http.ResponseWriter, r *http.Request) (*data.InlayProof, *data.Inlay, *data.Project, bool) {
	proofUUID := r.PathValue("uuid")

	err := m.Validate.Var(proofUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return nil, nil, nil, false
	}

	proof, found, err := m.Db.InlayProofs.GetByUUID(proofUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, nil, nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, nil, nil, false
	}

	inlay, found, err := m.Db.Inlays.GetByID(proof.InlayID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, nil, nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, nil, nil, false
	}

	project, found, err := m.Db.Projects.GetByID(inlay.ProjectID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, nil, nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, nil, nil, false
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return nil, nil, nil, false
		}
	}

	return proof, inlay, project, true
}

// authorizeProofAction enforces that the calling user has the right to act on
// this proof, branching on the proof's approval authority.
func (m ProofModule) authorizeProofAction(w http.ResponseWriter, r *http.Request, proof *data.InlayProof) bool {
	user := m.ContextGetUser(r)

	switch proof.ApprovalAuthority {
	case data.ProofApprovalAuthorities.Internal:
		if !user.IsInternal() || !user.Can(data.ActionInternalApproveProof) {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return false
		}
	default: // dealership
		if !user.Can(data.ActionApproveProof) {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return false
		}
	}

	return true
}

func (m ProofModule) HandleApproveProof(w http.ResponseWriter, r *http.Request) {
	proof, inlay, project, ok := m.loadProofWithContext(w, r)
	if !ok {
		return
	}

	if !m.authorizeProofAction(w, r, proof) {
		return
	}

	if proof.Status != data.ProofStatuses.Pending {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("proof is not in pending status"))
		return
	}

	// Internal-authority approvals may override the price group and apply a
	// price adjustment (percent or fixed) on top of the group's base price.
	var body struct {
		PriceGroupID         *int     `json:"price_group_id"`
		PriceAdjustmentType  *string  `json:"price_adjustment_type" validate:"omitempty,oneof=none percent fixed"`
		PriceAdjustmentValue *float64 `json:"price_adjustment_value"`
	}
	if r.ContentLength > 0 {
		if err := m.ReadJSONBody(w, r, &body); err != nil {
			m.WriteError(w, r, m.Err.BadRequest, err)
			return
		}
	}

	user := m.ContextGetUser(r)

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	if proof.ApprovalAuthority == data.ProofApprovalAuthorities.Internal {
		if body.PriceGroupID != nil {
			proof.PriceGroupID = body.PriceGroupID
		}
		if body.PriceAdjustmentType != nil {
			proof.PriceAdjustmentType = data.PriceAdjustmentType(*body.PriceAdjustmentType)
			proof.PriceAdjustmentValue = 0
			if body.PriceAdjustmentValue != nil {
				proof.PriceAdjustmentValue = *body.PriceAdjustmentValue
			}
		}
	}

	now := time.Now()
	userID := user.GetID()
	proof.Status = data.ProofStatuses.Approved
	proof.ApprovedAt = &now
	if user.IsInternal() {
		proof.ApprovedByInternalUserID = &userID
		proof.ApprovedByDealershipUserID = nil
	} else {
		proof.ApprovedByDealershipUserID = &userID
		proof.ApprovedByInternalUserID = nil
	}

	err = m.Db.InlayProofs.TxUpdate(tx, proof)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to approve proof: %w", err))
		return
	}

	// Only customer-facing approvals get a chat message; internal review of a
	// customized catalog inlay is an internal process.
	if proof.ApprovalAuthority == data.ProofApprovalAuthorities.Dealership {
		actorID := user.GetID()
		chatMessage := data.InlayChat{
			InlayID:     inlay.ID,
			MessageType: data.ChatMessageTypes.ProofApproved,
			Message:     fmt.Sprintf("Proof v%d approved", proof.VersionNumber),
		}
		if user.IsInternal() {
			chatMessage.InternalUserID = &actorID
		} else {
			chatMessage.DealershipUserID = &actorID
		}

		err = m.Db.InlayChats.TxInsert(tx, &chatMessage)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create approval chat message: %w", err))
			return
		}
	}

	inlay.ApprovedProofID = &proof.ID
	err = m.Db.Inlays.TxUpdateFields(tx, inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update inlay approved proof: %w", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	// Only notify internal users about dealership approvals (the internal user
	// who approves an internal-authority proof already knows about it).
	if proof.ApprovalAuthority == data.ProofApprovalAuthorities.Dealership {
		m.SendNotificationToAllInternalUsers(
			data.NotificationEventTypes.ProofApproved,
			fmt.Sprintf("Proof approved: %s", inlay.Name),
			fmt.Sprintf("Proof v%d for inlay %q has been approved.", proof.VersionNumber, inlay.Name),
			&project.ID, &inlay.ID,
		)
	}

	m.WriteJSON(w, r, http.StatusOK, proof)
}

func (m ProofModule) HandleDeclineProof(w http.ResponseWriter, r *http.Request) {
	proof, inlay, project, ok := m.loadProofWithContext(w, r)
	if !ok {
		return
	}

	if !m.authorizeProofAction(w, r, proof) {
		return
	}

	if proof.Status != data.ProofStatuses.Pending {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("proof is not in pending status"))
		return
	}

	var body struct {
		DeclineReason string `json:"decline_reason" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user := m.ContextGetUser(r)

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	now := time.Now()
	userID := user.GetID()
	proof.Status = data.ProofStatuses.Declined
	proof.DeclinedAt = &now
	proof.DeclineReason = &body.DeclineReason
	if user.IsInternal() {
		proof.DeclinedByInternalUserID = &userID
		proof.DeclinedByDealershipUserID = nil
	} else {
		proof.DeclinedByDealershipUserID = &userID
		proof.DeclinedByInternalUserID = nil
	}

	err = m.Db.InlayProofs.TxUpdate(tx, proof)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to decline proof: %w", err))
		return
	}

	if proof.ApprovalAuthority == data.ProofApprovalAuthorities.Dealership {
		actorID := user.GetID()
		chatMessage := data.InlayChat{
			InlayID:     inlay.ID,
			MessageType: data.ChatMessageTypes.ProofDeclined,
			Message:     fmt.Sprintf("Proof v%d declined: %s", proof.VersionNumber, body.DeclineReason),
		}
		if user.IsInternal() {
			chatMessage.InternalUserID = &actorID
		} else {
			chatMessage.DealershipUserID = &actorID
		}

		err = m.Db.InlayChats.TxInsert(tx, &chatMessage)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create decline chat message: %w", err))
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if proof.ApprovalAuthority == data.ProofApprovalAuthorities.Dealership {
		m.SendNotificationToAllInternalUsers(
			data.NotificationEventTypes.ProofDeclined,
			fmt.Sprintf("Proof declined: %s", inlay.Name),
			fmt.Sprintf("Proof v%d for inlay %q has been declined: %s", proof.VersionNumber, inlay.Name, body.DeclineReason),
			&project.ID, &inlay.ID,
		)
	}

	m.WriteJSON(w, r, http.StatusOK, proof)
}
