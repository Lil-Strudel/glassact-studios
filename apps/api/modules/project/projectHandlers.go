package project

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type ProjectModule struct {
	*app.Application
}

func NewProjectModule(app *app.Application) *ProjectModule {
	return &ProjectModule{
		app,
	}
}

func (m ProjectModule) HandleGetProjects(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var projects []*data.Project
	var err error

	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
		projects, err = m.Db.Projects.GetByDealershipID(*dealershipID)
	} else {
		projects, err = m.Db.Projects.GetAll()
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, projects)
}

func (m ProjectModule) HandleGetProjectByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(uuid)
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

	m.WriteJSON(w, r, http.StatusOK, project)
}

func (m ProjectModule) HandlePostProject(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var body struct {
		Name string `json:"name" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if !user.IsDealership() {
		m.WriteError(w, r, m.Err.Forbidden, fmt.Errorf("only dealership users can create projects"))
		return
	}

	dealershipID := user.GetDealershipID()
	if dealershipID == nil {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	project := data.Project{
		Name:         body.Name,
		Status:       data.ProjectStatuses.Draft,
		DealershipID: *dealershipID,
	}

	err = m.Db.Projects.Insert(&project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, project)
}

type projectWithInlaysResponse struct {
	data.Project
	Inlays []*data.Inlay `json:"inlays"`
}

func (m ProjectModule) HandlePostProjectWithInlays(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var body struct {
		Name   string `json:"name" validate:"required"`
		Inlays []struct {
			Name       string         `json:"name" validate:"required"`
			PreviewURL string         `json:"preview_url"`
			Type       data.InlayType `json:"type" validate:"required"`

			CatalogInfo *struct {
				CatalogItemID      int    `json:"catalog_item_id" validate:"required"`
				CustomizationNotes string `json:"customization_notes"`
			} `json:"catalog_info" validate:"required_if=Type catalog"`

			CustomInfo *struct {
				Description     string  `json:"description" validate:"required"`
				RequestedWidth  float64 `json:"requested_width"`
				RequestedHeight float64 `json:"requested_height"`
			} `json:"custom_info" validate:"required_if=Type custom"`
		} `json:"inlays" validate:"required,min=1,dive"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if !user.IsDealership() {
		m.WriteError(w, r, m.Err.Forbidden, fmt.Errorf("only dealership users can create projects"))
		return
	}

	dealershipID := user.GetDealershipID()
	if dealershipID == nil {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	project := data.Project{
		Name:         body.Name,
		Status:       data.ProjectStatuses.Draft,
		DealershipID: *dealershipID,
	}

	err = m.Db.Projects.TxInsert(tx, &project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	inlays := make([]*data.Inlay, 0, len(body.Inlays))

	for _, inlayBody := range body.Inlays {
		previewURL := inlayBody.PreviewURL

		if inlayBody.Type == data.InlayTypes.Catalog && inlayBody.CatalogInfo != nil {
			catalogItem, found, lookupErr := m.Db.CatalogItems.GetByID(inlayBody.CatalogInfo.CatalogItemID)
			if lookupErr != nil {
				m.WriteError(w, r, m.Err.ServerError, lookupErr)
				return
			}
			if !found {
				m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("catalog item with id %d not found", inlayBody.CatalogInfo.CatalogItemID))
				return
			}
			previewURL = catalogItem.SvgURL
		}

		inlay := data.Inlay{
			ProjectID:  project.ID,
			Name:       inlayBody.Name,
			PreviewURL: previewURL,
			Type:       inlayBody.Type,
		}

		if inlayBody.CatalogInfo != nil {
			inlay.CatalogInfo = &data.InlayCatalogInfo{
				CatalogItemID:      inlayBody.CatalogInfo.CatalogItemID,
				CustomizationNotes: inlayBody.CatalogInfo.CustomizationNotes,
			}
		}

		if inlayBody.CustomInfo != nil {
			inlay.CustomInfo = &data.InlayCustomInfo{
				Description:     inlayBody.CustomInfo.Description,
				RequestedWidth:  inlayBody.CustomInfo.RequestedWidth,
				RequestedHeight: inlayBody.CustomInfo.RequestedHeight,
			}
		}

		err = m.Db.Inlays.TxInsert(tx, &inlay)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}

		inlays = append(inlays, &inlay)
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	response := projectWithInlaysResponse{
		Project: project,
		Inlays:  inlays,
	}

	m.WriteJSON(w, r, http.StatusCreated, response)
}

func (m ProjectModule) HandlePatchProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name *string `json:"name"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(uuid)
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

	if body.Name != nil {
		project.Name = *body.Name
	}

	err = m.Db.Projects.Update(project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

var cancellableStatuses = map[data.ProjectStatus]bool{
	data.ProjectStatuses.Draft:           true,
	data.ProjectStatuses.Designing:       true,
	data.ProjectStatuses.PendingApproval: true,
	data.ProjectStatuses.Approved:        true,
}

func (m ProjectModule) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(uuid)
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

	if !cancellableStatuses[project.Status] {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("cannot cancel a project in %s status", project.Status))
		return
	}

	project.Status = data.ProjectStatuses.Cancelled

	err = m.Db.Projects.Update(project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

func (m ProjectModule) HandleSubmitProject(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
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

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("project must be in draft status to submit, currently: %s", project.Status))
		return
	}

	includedCount, err := m.Db.Inlays.CountIncludedByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if includedCount == 0 {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("project must have at least one included inlay to submit"))
		return
	}

	project.Status = data.ProjectStatuses.Designing

	err = m.Db.Projects.Update(project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update project status: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

func (m ProjectModule) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
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

	if project.Status != data.ProjectStatuses.Approved {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("project must be in approved status to place order, currently: %s", project.Status))
		return
	}

	allInlays, err := m.Db.Inlays.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	// Filter to only included inlays
	var inlays []*data.Inlay
	for _, inlayItem := range allInlays {
		if !inlayItem.ExcludedFromOrder {
			inlays = append(inlays, inlayItem)
		}
	}

	if len(inlays) == 0 {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("project has no included inlays"))
		return
	}

	for _, inlayItem := range inlays {
		if inlayItem.ApprovedProofID == nil {
			m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("inlay %q has no approved proof", inlayItem.Name))
			return
		}
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	orderedStep := "ordered"
	for _, inlayItem := range inlays {
		approvedProof, proofFound, proofErr := m.Db.InlayProofs.GetByID(*inlayItem.ApprovedProofID)
		if proofErr != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get approved proof for inlay %q: %w", inlayItem.Name, proofErr))
			return
		}

		if !proofFound {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("approved proof not found for inlay %q", inlayItem.Name))
			return
		}

		priceGroupID := 0
		if approvedProof.PriceGroupID != nil {
			priceGroupID = *approvedProof.PriceGroupID
		}

		priceCents := 0
		if approvedProof.PriceCents != nil {
			priceCents = *approvedProof.PriceCents
		}

		snapshot := data.OrderSnapshot{
			ProjectID:    project.ID,
			InlayID:      inlayItem.ID,
			ProofID:      approvedProof.ID,
			PriceGroupID: priceGroupID,
			PriceCents:   priceCents,
			Width:        approvedProof.Width,
			Height:       approvedProof.Height,
		}

		snapshotErr := m.Db.OrderSnapshots.TxInsert(tx, &snapshot)
		if snapshotErr != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create order snapshot for inlay %q: %w", inlayItem.Name, snapshotErr))
			return
		}

		inlayItem.ManufacturingStep = &orderedStep
		inlayErr := m.Db.Inlays.TxUpdateFields(tx, inlayItem)
		if inlayErr != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update inlay manufacturing step: %w", inlayErr))
			return
		}
	}

	now := time.Now()
	userID := user.GetID()
	project.Status = data.ProjectStatuses.Ordered
	project.OrderedAt = &now
	project.OrderedBy = &userID

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

	m.WriteJSON(w, r, http.StatusOK, project)
}
