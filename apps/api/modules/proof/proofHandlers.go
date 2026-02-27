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

func (m ProofModule) HandleCreateProof(w http.ResponseWriter, r *http.Request) {
	inlay, project, ok := m.getInlayWithAccessCheck(w, r)
	if !ok {
		return
	}

	var body struct {
		DesignAssetURL string                 `json:"design_asset_url" validate:"required"`
		Width          float64                `json:"width" validate:"required,gt=0"`
		Height         float64                `json:"height" validate:"required,gt=0"`
		PriceGroupID   *int                   `json:"price_group_id"`
		PriceCents     *int                   `json:"price_cents"`
		ScaleFactor    *float64               `json:"scale_factor"`
		ColorOverrides map[string]interface{} `json:"color_overrides"`
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

	if project.Status == data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("cannot create proofs for a project that has not been submitted for design"))
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

	proof := data.InlayProof{
		InlayID:        inlay.ID,
		VersionNumber:  versionNumber,
		DesignAssetURL: body.DesignAssetURL,
		Width:          body.Width,
		Height:         body.Height,
		PriceGroupID:   body.PriceGroupID,
		PriceCents:     body.PriceCents,
		ScaleFactor:    scaleFactor,
		ColorOverrides: colorOverrides,
		Status:         data.ProofStatuses.Pending,
		SentInChatID:   chatMessage.ID,
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

	m.WriteJSON(w, r, http.StatusCreated, proof)
}

func (m ProofModule) HandleApproveProof(w http.ResponseWriter, r *http.Request) {
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

	if proof.Status != data.ProofStatuses.Pending {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("proof is not in pending status"))
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

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	now := time.Now()
	userID := user.GetID()
	proof.Status = data.ProofStatuses.Approved
	proof.ApprovedAt = &now
	proof.ApprovedBy = &userID

	err = m.Db.InlayProofs.TxUpdate(tx, proof)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to approve proof: %w", err))
		return
	}

	dealershipUserID := user.GetID()
	chatMessage := data.InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUserID,
		MessageType:      data.ChatMessageTypes.ProofApproved,
		Message:          fmt.Sprintf("Proof v%d approved", proof.VersionNumber),
	}

	err = m.Db.InlayChats.TxInsert(tx, &chatMessage)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create approval chat message: %w", err))
		return
	}

	inlay.ApprovedProofID = &proof.ID
	err = m.Db.Inlays.TxUpdateFields(tx, inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update inlay approved proof: %w", err))
		return
	}

	allInlays, err := m.Db.Inlays.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	allApproved := true
	for _, projectInlay := range allInlays {
		if projectInlay.ExcludedFromOrder {
			continue
		}
		if projectInlay.ID == inlay.ID {
			continue
		}
		if projectInlay.ApprovedProofID == nil {
			allApproved = false
			break
		}
	}

	if allApproved {
		project.Status = data.ProjectStatuses.Approved
	} else {
		project.Status = data.ProjectStatuses.PendingApproval
	}

	err = m.Db.Projects.TxUpdate(tx, project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update project status: %w", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, proof)
}

func (m ProofModule) HandleDeclineProof(w http.ResponseWriter, r *http.Request) {
	proofUUID := r.PathValue("uuid")

	err := m.Validate.Var(proofUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		DeclineReason string `json:"decline_reason" validate:"required"`
	}

	err = m.ReadJSONBody(w, r, &body)
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

	if proof.Status != data.ProofStatuses.Pending {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("proof is not in pending status"))
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
	proof.DeclinedBy = &userID
	proof.DeclineReason = &body.DeclineReason

	err = m.Db.InlayProofs.TxUpdate(tx, proof)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to decline proof: %w", err))
		return
	}

	dealershipUserID := user.GetID()
	chatMessage := data.InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUserID,
		MessageType:      data.ChatMessageTypes.ProofDeclined,
		Message:          fmt.Sprintf("Proof v%d declined: %s", proof.VersionNumber, body.DeclineReason),
	}

	err = m.Db.InlayChats.TxInsert(tx, &chatMessage)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create decline chat message: %w", err))
		return
	}

	if project.Status == data.ProjectStatuses.PendingApproval || project.Status == data.ProjectStatuses.Approved {
		project.Status = data.ProjectStatuses.Designing
		err = m.Db.Projects.TxUpdate(tx, project)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update project status: %w", err))
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, proof)
}
