package grout

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type GroutModule struct {
	*app.Application
}

func NewGroutModule(app *app.Application) *GroutModule {
	return &GroutModule{app}
}

// HandleGetGrouts returns the active grout (background) palette, sorted, as a
// flat array for the customizer's background picker.
func (m *GroutModule) HandleGetGrouts(w http.ResponseWriter, r *http.Request) {
	grouts, err := m.Db.Grouts.GetAllActive()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, grouts)
}

func (m *GroutModule) HandleGetGrout(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	grout, found, err := m.Db.Grouts.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, grout)
}
