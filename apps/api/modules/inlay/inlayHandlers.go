package inlay

import (
	"fmt"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type InlayModule struct {
	*app.Application
}

func NewInlayModule(app *app.Application) *InlayModule {
	return &InlayModule{
		app,
	}
}

func (m InlayModule) getProjectForInlayAccess(w http.ResponseWriter, r *http.Request, projectUUID string) (*data.Project, bool) {
	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, false
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return nil, false
		}
	}

	return project, true
}

func (m InlayModule) validateInlayOwnership(w http.ResponseWriter, r *http.Request, inlay *data.Inlay) (*data.Project, bool) {
	project, found, err := m.Db.Projects.GetByID(inlay.ProjectID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, false
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return nil, false
		}
	}

	return project, true
}

func (m InlayModule) HandleGetInlaysByProject(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectForInlayAccess(w, r, projectUUID)
	if !ok {
		return
	}

	inlays, err := m.Db.Inlays.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlays)
}

func (m InlayModule) HandleGetInlayByUUID(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
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

	if _, ok := m.validateInlayOwnership(w, r, inlay); !ok {
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlay)
}

func (m InlayModule) HandlePostCatalogInlay(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name               string `json:"name" validate:"required"`
		CatalogItemID      int    `json:"catalog_item_id" validate:"required,gt=0"`
		CustomizationNotes string `json:"customization_notes"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectForInlayAccess(w, r, projectUUID)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft or designing status, current status: %s", project.Status))
		return
	}

	catalogItem, found, err := m.Db.CatalogItems.GetByID(body.CatalogItemID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("catalog item with id %d not found", body.CatalogItemID))
		return
	}

	inlay := data.Inlay{
		ProjectID:  project.ID,
		Name:       body.Name,
		Type:       data.InlayTypes.Catalog,
		PreviewURL: catalogItem.SvgURL,
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID:      body.CatalogItemID,
			CustomizationNotes: body.CustomizationNotes,
		},
	}

	err = m.Db.Inlays.Insert(&inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, inlay)
}

func (m InlayModule) HandlePostCustomInlay(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name            string  `json:"name" validate:"required"`
		Description     string  `json:"description" validate:"required"`
		RequestedWidth  float64 `json:"requested_width"`
		RequestedHeight float64 `json:"requested_height"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectForInlayAccess(w, r, projectUUID)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft or designing status, current status: %s", project.Status))
		return
	}

	inlay := data.Inlay{
		ProjectID:  project.ID,
		Name:       body.Name,
		Type:       data.InlayTypes.Custom,
		PreviewURL: "",
		CustomInfo: &data.InlayCustomInfo{
			Description:     body.Description,
			RequestedWidth:  body.RequestedWidth,
			RequestedHeight: body.RequestedHeight,
		},
	}

	err = m.Db.Inlays.Insert(&inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, inlay)
}

func (m InlayModule) HandlePatchInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
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

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only modify inlays on projects in draft or designing status"))
		return
	}

	if body.Name != nil {
		inlay.Name = *body.Name
	}

	err = m.Db.Inlays.Update(inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlay)
}

func (m InlayModule) HandleDeleteInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
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

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only remove inlays from projects in draft or designing status"))
		return
	}

	err = m.Db.Inlays.Delete(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

var preOrderStatuses = map[data.ProjectStatus]bool{
	data.ProjectStatuses.Draft:           true,
	data.ProjectStatuses.Designing:       true,
	data.ProjectStatuses.PendingApproval: true,
	data.ProjectStatuses.Approved:        true,
}

func (m InlayModule) HandleExcludeInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Excluded bool `json:"excluded"`
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

	if !preOrderStatuses[project.Status] {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only change inlay inclusion on projects before ordering"))
		return
	}

	inlay.ExcludedFromOrder = body.Excluded

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	err = m.Db.Inlays.TxUpdateFields(tx, inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlay)
}
