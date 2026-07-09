package glasscolor

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
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

// HandleGetGlassColorsAdmin returns the full glass palette (including inactive
// colors) as a paginated list for the admin management screen.
func (m *GlassColorModule) HandleGetGlassColorsAdmin(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	items, err := m.Db.GlassColors.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	if offset >= len(items) {
		items = []*data.GlassColor{}
	} else {
		items = items[offset:end]
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"items":  items,
		"total":  len(items),
		"limit":  limit,
		"offset": offset,
	})
}

func (m *GlassColorModule) HandlePostGlassColor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string  `json:"name" validate:"required,min=1,max=255"`
		Hex       string  `json:"hex" validate:"required,hexcolor"`
		Family    *string `json:"family"`
		SortOrder int     `json:"sort_order"`
		IsActive  bool    `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	glassColor := &data.GlassColor{
		Name:      body.Name,
		Hex:       body.Hex,
		Family:    body.Family,
		SortOrder: body.SortOrder,
		IsActive:  body.IsActive,
	}

	err = m.Db.GlassColors.Insert(glassColor)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, glassColor)
}

func (m *GlassColorModule) HandlePatchGlassColor(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
		return
	}

	var body struct {
		Name      *string `json:"name"`
		Hex       *string `json:"hex" validate:"omitempty,hexcolor"`
		Family    *string `json:"family"`
		SortOrder *int    `json:"sort_order"`
		IsActive  *bool   `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
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

	if body.Name != nil {
		glassColor.Name = *body.Name
	}
	if body.Hex != nil {
		glassColor.Hex = *body.Hex
	}
	if body.Family != nil {
		glassColor.Family = body.Family
	}
	if body.SortOrder != nil {
		glassColor.SortOrder = *body.SortOrder
	}
	if body.IsActive != nil {
		glassColor.IsActive = *body.IsActive
	}

	err = m.Db.GlassColors.Update(glassColor)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, glassColor)
}

// HandleDeleteGlassColor deactivates a glass color rather than hard-deleting it,
// so existing customizer selections and order snapshots keep resolving.
func (m *GlassColorModule) HandleDeleteGlassColor(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
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

	glassColor.IsActive = false
	err = m.Db.GlassColors.Update(glassColor)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
