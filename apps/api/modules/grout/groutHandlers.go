package grout

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
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

// HandleGetGroutsAdmin returns the full grout palette (including inactive grouts)
// as a paginated list for the admin management screen.
func (m *GroutModule) HandleGetGroutsAdmin(w http.ResponseWriter, r *http.Request) {
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

	items, err := m.Db.Grouts.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	if offset >= len(items) {
		items = []*data.Grout{}
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

func (m *GroutModule) HandlePostGrout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string `json:"name" validate:"required,min=1,max=255"`
		Hex       string `json:"hex" validate:"required,hexcolor"`
		SortOrder int    `json:"sort_order"`
		IsActive  bool   `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	grout := &data.Grout{
		Name:      body.Name,
		Hex:       body.Hex,
		SortOrder: body.SortOrder,
		IsActive:  body.IsActive,
	}

	err = m.Db.Grouts.Insert(grout)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, grout)
}

func (m *GroutModule) HandlePatchGrout(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
		return
	}

	var body struct {
		Name      *string `json:"name"`
		Hex       *string `json:"hex" validate:"omitempty,hexcolor"`
		SortOrder *int    `json:"sort_order"`
		IsActive  *bool   `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
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

	if body.Name != nil {
		grout.Name = *body.Name
	}
	if body.Hex != nil {
		grout.Hex = *body.Hex
	}
	if body.SortOrder != nil {
		grout.SortOrder = *body.SortOrder
	}
	if body.IsActive != nil {
		grout.IsActive = *body.IsActive
	}

	err = m.Db.Grouts.Update(grout)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, grout)
}

// HandleDeleteGrout deactivates a grout rather than hard-deleting it, so existing
// customizer selections and order snapshots keep resolving.
func (m *GroutModule) HandleDeleteGrout(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
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

	grout.IsActive = false
	err = m.Db.Grouts.Update(grout)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
