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
