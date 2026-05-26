package glasscolor

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type GlassColorModule struct {
	*app.Application
}

func NewGlassColorModule(app *app.Application) *GlassColorModule {
	return &GlassColorModule{app}
}

// HandleGetGlassColors returns the active glass palette, sorted, as a flat array
// for the customizer's color picker.
func (m *GlassColorModule) HandleGetGlassColors(w http.ResponseWriter, r *http.Request) {
	glassColors, err := m.Db.GlassColors.GetAllActive()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, glassColors)
}

func (m *GlassColorModule) HandleGetGlassColor(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	glassColor, found, err := m.Db.GlassColors.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, glassColor)
}
