package project

import (
	"fmt"
	"net/http"

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
