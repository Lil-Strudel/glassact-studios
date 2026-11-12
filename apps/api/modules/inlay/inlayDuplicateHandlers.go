package inlay

import (
	"errors"
	"fmt"
	"net/http"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// HandleDuplicateInlay copies an inlay within its draft project. An approved
// design carries over, so a copy of a ready inlay is ready too; an unapproved
// one starts wherever its original stands. If the original is on the stone,
// the copy is placed beside it.
func (m InlayModule) HandleDuplicateInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	source, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	project, ok := m.validateInlayOwnership(w, r, source)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only duplicate inlays on projects in draft status, current status: %s", project.Status))
		return
	}

	sourceProof, isApproved, err := m.proofToDuplicate(source)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	duplicate := duplicateInlay(source)
	if err := m.Db.Inlays.TxInsert(tx, duplicate); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to insert duplicate of inlay %d: %w", source.ID, err))
		return
	}

	if sourceProof != nil {
		proof := duplicateProof(sourceProof, duplicate.ID, isApproved)
		if err := m.Db.InlayProofs.TxInsert(tx, proof); err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to copy proof %d to duplicate inlay %d: %w", sourceProof.ID, duplicate.ID, err))
			return
		}

		if isApproved {
			// insertProof leaves the approval columns unset, so they are written
			// in a second step.
			if err := m.Db.InlayProofs.TxUpdate(tx, proof); err != nil {
				m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to copy approval to proof %d: %w", proof.ID, err))
				return
			}

			duplicate.ApprovedProofID = &proof.ID
			if err := m.Db.Inlays.TxUpdateFields(tx, duplicate); err != nil {
				m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to mark duplicate inlay %d approved: %w", duplicate.ID, err))
				return
			}
		}
	}

	_, err = m.Db.StoneDesigns.TxPlaceDuplicate(tx, project.ID, source.UUID, duplicate.UUID)
	if err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			m.WriteError(w, r, m.Err.Conflict, errors.New("the stone design was changed at the same moment; try again"))
			return
		}
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to place duplicate inlay %d on the stone: %w", duplicate.ID, err))
		return
	}

	if err := tx.Commit(); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	user := m.ContextGetUser(r)
	m.AutoWatchProject(project.ID, user)
	m.notifyDuplicateNeedsWork(project, user, duplicate)

	m.WriteJSON(w, r, http.StatusCreated, duplicate)
}

// proofToDuplicate picks the proof a copy inherits, and whether it inherits
// the approval too: the approved proof if there is one, otherwise a customized
// inlay's latest proof so the copy goes to the same internal review. A custom
// inlay without an approved proof inherits none; a designer proofs the copy
// from its description.
func (m InlayModule) proofToDuplicate(source *data.Inlay) (*data.InlayProof, bool, error) {
	if source.ApprovedProofID != nil {
		proof, found, err := m.Db.InlayProofs.GetByID(*source.ApprovedProofID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to load approved proof for inlay %d: %w", source.ID, err)
		}
		if found {
			return proof, true, nil
		}
	}

	if source.Type == data.InlayTypes.Catalog && source.IsCustomized {
		proof, found, err := m.Db.InlayProofs.GetLatestByInlayID(source.ID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to load latest proof for inlay %d: %w", source.ID, err)
		}
		if found {
			return proof, false, nil
		}
	}

	return nil, false, nil
}

func duplicateInlay(source *data.Inlay) *data.Inlay {
	duplicate := &data.Inlay{
		ProjectID:    source.ProjectID,
		Name:         source.Name + " (copy)",
		Type:         source.Type,
		IsCustomized: source.IsCustomized,
		PreviewURL:   source.PreviewURL,
		MirrorX:      source.MirrorX,
		MirrorY:      source.MirrorY,
	}

	if source.CatalogInfo != nil {
		duplicate.CatalogInfo = &data.InlayCatalogInfo{
			CatalogItemID:      source.CatalogInfo.CatalogItemID,
			CustomizationNotes: source.CatalogInfo.CustomizationNotes,
		}
	}

	if source.CustomInfo != nil {
		images := make([]data.InlayCustomReferenceImage, len(source.CustomInfo.ReferenceImages))
		for i, image := range source.CustomInfo.ReferenceImages {
			images[i] = data.InlayCustomReferenceImage{ImageURL: image.ImageURL}
		}
		duplicate.CustomInfo = &data.InlayCustomInfo{
			Description:     source.CustomInfo.Description,
			RequestedWidth:  source.CustomInfo.RequestedWidth,
			RequestedHeight: source.CustomInfo.RequestedHeight,
			ReferenceImages: images,
		}
	}

	return duplicate
}

// duplicateProof copies a proof as version 1 of the duplicate's history, either
// keeping its approval or as a fresh pending proof. The chat link stays behind
// with the original.
func duplicateProof(source *data.InlayProof, inlayID int, keepApproval bool) *data.InlayProof {
	proof := &data.InlayProof{
		InlayID:              inlayID,
		VersionNumber:        1,
		DesignAssetURL:       source.DesignAssetURL,
		Width:                source.Width,
		Height:               source.Height,
		PriceGroupID:         source.PriceGroupID,
		PriceAdjustmentType:  source.PriceAdjustmentType,
		PriceAdjustmentValue: source.PriceAdjustmentValue,
		ScaleFactor:          source.ScaleFactor,
		ColorOverrides:       source.ColorOverrides,
		ApprovalAuthority:    source.ApprovalAuthority,
		Status:               data.ProofStatuses.Pending,
	}

	if keepApproval {
		proof.Status = data.ProofStatuses.Approved
		proof.ApprovedAt = source.ApprovedAt
		proof.ApprovedByDealershipUserID = source.ApprovedByDealershipUserID
		proof.ApprovedByInternalUserID = source.ApprovedByInternalUserID
	}

	return proof
}

// notifyDuplicateNeedsWork sends the same notification adding the inlay from
// scratch would, when the copy is not ready and needs GlassAct to act on it.
func (m InlayModule) notifyDuplicateNeedsWork(project *data.Project, user data.AuthUser, duplicate *data.Inlay) {
	if inlayIsReady(duplicate) {
		return
	}

	if duplicate.Type == data.InlayTypes.Custom {
		m.NotifyInternal(
			project.ID,
			user,
			data.NotificationEventTypes.CustomInlaySubmitted,
			fmt.Sprintf("Custom inlay needs a proof: %s", duplicate.Name),
			fmt.Sprintf("A custom inlay %q was submitted and needs a designer to create a proof.", duplicate.Name),
			&duplicate.ID,
		)
		return
	}

	m.NotifyInternal(
		project.ID,
		user,
		data.NotificationEventTypes.InternalReviewRequired,
		fmt.Sprintf("Customized inlay needs review: %s", duplicate.Name),
		fmt.Sprintf("A customized inlay %q is ready for internal pricing review.", duplicate.Name),
		&duplicate.ID,
	)
}
