package pricegroup

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type PriceGroupModule struct {
	*app.Application
}

func NewPriceGroupModule(app *app.Application) *PriceGroupModule {
	return &PriceGroupModule{app}
}

func (m *PriceGroupModule) HandleGetPriceGroups(w http.ResponseWriter, r *http.Request) {
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

	items, err := m.Db.PriceGroups.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	if offset >= len(items) {
		items = []*data.PriceGroup{}
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

func (m *PriceGroupModule) HandlePostPriceGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name           string  `json:"name" validate:"required,min=1,max=255"`
		BasePriceCents int     `json:"base_price_cents" validate:"required,gt=0"`
		Description    *string `json:"description"`
		IsActive       bool    `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	priceGroup := &data.PriceGroup{
		Name:           body.Name,
		BasePriceCents: body.BasePriceCents,
		Description:    body.Description,
		IsActive:       body.IsActive,
	}

	err = m.Db.PriceGroups.Insert(priceGroup)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, priceGroup)
}

func (m *PriceGroupModule) HandleGetPriceGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
		return
	}

	priceGroup, found, err := m.Db.PriceGroups.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, priceGroup)
}

func (m *PriceGroupModule) HandlePatchPriceGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
		return
	}

	var body struct {
		Name           *string `json:"name"`
		BasePriceCents *int    `json:"base_price_cents"`
		Description    *string `json:"description"`
		IsActive       *bool   `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	priceGroup, found, err := m.Db.PriceGroups.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if body.Name != nil {
		priceGroup.Name = *body.Name
	}
	if body.BasePriceCents != nil {
		priceGroup.BasePriceCents = *body.BasePriceCents
	}
	if body.Description != nil {
		priceGroup.Description = body.Description
	}
	if body.IsActive != nil {
		priceGroup.IsActive = *body.IsActive
	}

	err = m.Db.PriceGroups.Update(priceGroup)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, priceGroup)
}

func (m *PriceGroupModule) HandleDeletePriceGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("uuid required"))
		return
	}

	priceGroup, found, err := m.Db.PriceGroups.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	priceGroup.IsActive = false
	err = m.Db.PriceGroups.Update(priceGroup)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
