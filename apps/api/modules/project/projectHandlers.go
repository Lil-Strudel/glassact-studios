package project

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
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
	projects, err := m.Db.Projects.GetAll()
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

	m.WriteJSON(w, r, http.StatusOK, project)
}

func (m ProjectModule) HandlePostProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string             `json:"name" validate:"required"`
		Status       data.ProjectStatus `json:"status" validate:"required"`
		Approved     bool               `json:"approved"`
		DealershipID int                `json:"dealership_id" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project := data.Project{
		Name:         body.Name,
		Status:       body.Status,
		Approved:     body.Approved,
		DealershipID: body.DealershipID,
	}

	err = m.Db.Projects.Insert(&project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

func (m ProjectModule) HandlePostProjectWithInlays(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string             `json:"name" validate:"required"`
		Status       data.ProjectStatus `json:"status" validate:"required"`
		Approved     bool               `json:"approved"`
		DealershipID int                `json:"dealership_id" validate:"required"`
		Inlays       []struct {
			Name       string         `json:"name" validate:"required"`
			PreviewURL string         `json:"preview_url" validate:"required"`
			PriceGroup int            `json:"price_group" validate:"required"`
			Type       data.InlayType `json:"type" validate:"required"`

			CatalogInfo *struct {
				CatalogItemID int `json:"catalog_item_id" validate:"required"`
			} `json:"catalog_info" validate:"required_if=Type catalog,excluded_unless=Type catalog"`

			CustomInfo *struct {
				Description string  `json:"description" validate:"required"`
				Width       float64 `json:"width" validate:"required"`
				Height      float64 `json:"height" validate:"required"`
			} `json:"custom_info" validate:"required_if=Type custom,excluded_unless=Type custom"`
		} `json:"inlays" validate:"required,dive"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
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
		Status:       body.Status,
		Approved:     body.Approved,
		DealershipID: body.DealershipID,
	}

	err = m.Db.Projects.TxInsert(tx, &project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	for _, body := range body.Inlays {
		inlay := data.Inlay{
			ProjectID:  project.ID,
			Name:       body.Name,
			PreviewURL: body.PreviewURL,
			PriceGroup: body.PriceGroup,
			Type:       body.Type,
		}

		if body.CatalogInfo != nil {
			inlay.CatalogInfo = &data.InlayCatalogInfo{
				CatalogItemID: body.CatalogInfo.CatalogItemID,
			}
		}

		if body.CustomInfo != nil {
			inlay.CustomInfo = &data.InlayCustomInfo{
				Description: body.CustomInfo.Description,
				Width:       body.CustomInfo.Width,
				Height:      body.CustomInfo.Height,
			}
		}

		err = m.Db.Inlays.TxInsert(tx, &inlay)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, nil)
}
