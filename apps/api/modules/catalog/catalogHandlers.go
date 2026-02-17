package catalog

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type CatalogModule struct {
	*app.Application
}

func NewCatalogModule(app *app.Application) *CatalogModule {
	return &CatalogModule{app}
}

func (m *CatalogModule) HandleGetCatalog(w http.ResponseWriter, r *http.Request) {
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

	search := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")
	isActive := r.URL.Query().Get("is_active")

	items, err := m.Db.CatalogItems.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	filtered := filterCatalogItems(items, search, category, isActive)

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	if offset >= len(filtered) {
		filtered = []*data.CatalogItem{}
	} else {
		filtered = filtered[offset:end]
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"items":  filtered,
		"total":  len(items),
		"limit":  limit,
		"offset": offset,
	})
}

func (m *CatalogModule) HandlePostCatalog(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CatalogCode         string  `json:"catalog_code" validate:"required,min=1,max=255"`
		Name                string  `json:"name" validate:"required,min=1,max=255"`
		Description         *string `json:"description"`
		Category            string  `json:"category" validate:"required,min=1,max=255"`
		DefaultWidth        float64 `json:"default_width" validate:"required,gt=0"`
		DefaultHeight       float64 `json:"default_height" validate:"required,gt=0"`
		MinWidth            float64 `json:"min_width" validate:"required,gt=0"`
		MinHeight           float64 `json:"min_height" validate:"required,gt=0"`
		DefaultPriceGroupID int     `json:"default_price_group_id" validate:"required,gt=0"`
		SvgURL              string  `json:"svg_url" validate:"required,min=1"`
		IsActive            bool    `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.DefaultWidth < body.MinWidth || body.DefaultHeight < body.MinHeight {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("default dimensions must be >= minimum dimensions"))
		return
	}

	catalogItem := &data.CatalogItem{
		CatalogCode:         body.CatalogCode,
		Name:                body.Name,
		Description:         body.Description,
		Category:            body.Category,
		DefaultWidth:        body.DefaultWidth,
		DefaultHeight:       body.DefaultHeight,
		MinWidth:            body.MinWidth,
		MinHeight:           body.MinHeight,
		DefaultPriceGroupID: body.DefaultPriceGroupID,
		SvgURL:              body.SvgURL,
		IsActive:            body.IsActive,
	}

	err = m.Db.CatalogItems.Insert(catalogItem)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, catalogItem)
}

// HandleGetCatalogItem gets a single catalog item (public)
func (m *CatalogModule) HandleGetCatalogItem(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, item)
}

// HandlePatchCatalog updates a catalog item (internal admin only)
func (m *CatalogModule) HandlePatchCatalog(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	var body struct {
		Name                *string  `json:"name"`
		Description         *string  `json:"description"`
		Category            *string  `json:"category"`
		DefaultWidth        *float64 `json:"default_width"`
		DefaultHeight       *float64 `json:"default_height"`
		MinWidth            *float64 `json:"min_width"`
		MinHeight           *float64 `json:"min_height"`
		DefaultPriceGroupID *int     `json:"default_price_group_id"`
		SvgURL              *string  `json:"svg_url"`
		IsActive            *bool    `json:"is_active"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	// Apply updates
	if body.Name != nil {
		item.Name = *body.Name
	}
	if body.Description != nil {
		item.Description = body.Description
	}
	if body.Category != nil {
		item.Category = *body.Category
	}
	if body.DefaultWidth != nil {
		item.DefaultWidth = *body.DefaultWidth
	}
	if body.DefaultHeight != nil {
		item.DefaultHeight = *body.DefaultHeight
	}
	if body.MinWidth != nil {
		item.MinWidth = *body.MinWidth
	}
	if body.MinHeight != nil {
		item.MinHeight = *body.MinHeight
	}
	if body.DefaultPriceGroupID != nil {
		item.DefaultPriceGroupID = *body.DefaultPriceGroupID
	}
	if body.SvgURL != nil {
		item.SvgURL = *body.SvgURL
	}
	if body.IsActive != nil {
		item.IsActive = *body.IsActive
	}

	// Validate dimensions
	if item.DefaultWidth < item.MinWidth || item.DefaultHeight < item.MinHeight {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("default dimensions must be >= minimum dimensions"))
		return
	}

	err = m.Db.CatalogItems.Update(item)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, item)
}

// HandleDeleteCatalog soft deletes a catalog item (internal admin only)
func (m *CatalogModule) HandleDeleteCatalog(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	// Soft delete
	item.IsActive = false
	err = m.Db.CatalogItems.Update(item)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

// HandleGetTags gets tags for a catalog item (public)
func (m *CatalogModule) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, item.Tags)
}

// HandlePostTag adds a tag to a catalog item (internal admin only)
func (m *CatalogModule) HandlePostTag(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	var body struct {
		Tag string `json:"tag" validate:"required,min=1,max=50"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	// Check for duplicate
	for _, existing := range item.Tags {
		if existing == body.Tag {
			m.WriteError(w, r, m.Err.BadRequest, errors.New("tag already exists"))
			return
		}
	}

	err = m.Db.CatalogItems.AddTag(item.ID, body.Tag)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	item.Tags = append(item.Tags, body.Tag)
	m.WriteJSON(w, r, http.StatusOK, item.Tags)
}

// HandleDeleteTag removes a tag from a catalog item (internal admin only)
func (m *CatalogModule) HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	tag := r.PathValue("tag")

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	err = m.Db.CatalogItems.RemoveTag(item.ID, tag)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	// Update item's tags
	filtered := make([]string, 0, len(item.Tags))
	for _, t := range item.Tags {
		if t != tag {
			filtered = append(filtered, t)
		}
	}

	m.WriteJSON(w, r, http.StatusOK, filtered)
}

// HandleBrowseCatalog returns active catalog items for dealership users (dealership)
func (m *CatalogModule) HandleBrowseCatalog(w http.ResponseWriter, r *http.Request) {
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

	search := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")
	tagsParam := r.URL.Query().Get("tags")

	var tags []string
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	items, err := m.Db.CatalogItems.GetAllActive()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	// Filter
	filtered := filterCatalogItemsWithTags(items, search, category, tags)

	// Paginate
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	if offset >= len(filtered) {
		filtered = []*data.CatalogItem{}
	} else {
		filtered = filtered[offset:end]
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"items":  filtered,
		"total":  len(items),
		"limit":  limit,
		"offset": offset,
	})
}

// HandleGetCategories returns all distinct categories (dealership)
func (m *CatalogModule) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := m.Db.CatalogItems.GetCategories()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, categories)
}

// HandleGetAllTags returns all distinct tags (dealership)
func (m *CatalogModule) HandleGetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := m.Db.CatalogItems.GetAllTags()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, tags)
}

// Helper functions

func filterCatalogItems(items []*data.CatalogItem, search, category, isActive string) []*data.CatalogItem {
	filtered := make([]*data.CatalogItem, 0, len(items))

	for _, item := range items {
		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), searchLower) &&
				!strings.Contains(strings.ToLower(item.CatalogCode), searchLower) {
				continue
			}
		}

		// Category filter
		if category != "" && item.Category != category {
			continue
		}

		// Active filter
		if isActive != "" {
			active := isActive == "true"
			if item.IsActive != active {
				continue
			}
		}

		filtered = append(filtered, item)
	}

	return filtered
}

func filterCatalogItemsWithTags(items []*data.CatalogItem, search, category string, tags []string) []*data.CatalogItem {
	filtered := make([]*data.CatalogItem, 0, len(items))

	for _, item := range items {
		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), searchLower) &&
				!strings.Contains(strings.ToLower(item.CatalogCode), searchLower) {
				continue
			}
		}

		// Category filter
		if category != "" && item.Category != category {
			continue
		}

		// Tags filter - ALL tags must match
		if len(tags) > 0 {
			hasAllTags := true
			for _, requiredTag := range tags {
				found := false
				for _, itemTag := range item.Tags {
					if itemTag == requiredTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		filtered = append(filtered, item)
	}

	return filtered
}
