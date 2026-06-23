package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/svg"
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

// catalogWriteRequest is the finalized create/update body. The manifest is
// already perfected in the admin editor (every glass/grout id assigned); the
// server bakes the structure SVG into the baked asset and stores it.
type catalogWriteRequest struct {
	CatalogCode         string          `json:"catalog_code" validate:"required,min=1,max=255"`
	Name                string          `json:"name" validate:"required,min=1,max=255"`
	Description         *string         `json:"description"`
	Category            string          `json:"category" validate:"required,min=1,max=255"`
	DefaultWidth        float64         `json:"default_width" validate:"required,gt=0"`
	DefaultHeight       float64         `json:"default_height" validate:"required,gt=0"`
	MinWidth            float64         `json:"min_width" validate:"required,gt=0"`
	MinHeight           float64         `json:"min_height" validate:"required,gt=0"`
	DefaultPriceGroupID int             `json:"default_price_group_id" validate:"required,gt=0"`
	SvgURL              string          `json:"svg_url" validate:"required,min=1"`
	Manifest            svg.Manifest    `json:"manifest"`
	ContentBBox         svg.ContentBBox `json:"content_bbox"`
	IsActive            bool            `json:"is_active"`
	Tags                []string        `json:"tags"`
}

func (m *CatalogModule) HandlePostCatalog(w http.ResponseWriter, r *http.Request) {
	var body catalogWriteRequest

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.DefaultWidth < body.MinWidth || body.DefaultHeight < body.MinHeight {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("default dimensions must be >= minimum dimensions"))
		return
	}

	if err := validateManifestAssigned(body.Manifest); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
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

	if err := m.bakeAndStore(r.Context(), catalogItem, body.Manifest, body.ContentBBox); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	err = m.Db.CatalogItems.Insert(catalogItem)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	for _, tag := range body.Tags {
		if err := m.Db.CatalogItems.AddTag(catalogItem.ID, tag); err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		catalogItem.Tags = append(catalogItem.Tags, tag)
	}

	m.WriteJSON(w, r, http.StatusCreated, catalogItem)
}

// HandlePutCatalog fully updates a catalog item from a finalized write request.
// The structure SVG is re-baked when the manifest or target dimensions change,
// otherwise the existing baked svg_url is preserved.
func (m *CatalogModule) HandlePutCatalog(w http.ResponseWriter, r *http.Request) {
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

	var body catalogWriteRequest
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.DefaultWidth < body.MinWidth || body.DefaultHeight < body.MinHeight {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("default dimensions must be >= minimum dimensions"))
		return
	}

	if err := validateManifestAssigned(body.Manifest); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	dimsChanged := body.DefaultWidth != item.DefaultWidth || body.DefaultHeight != item.DefaultHeight
	manifestChanged := !manifestEqual(item.Manifest, body.Manifest)

	item.CatalogCode = body.CatalogCode
	item.Name = body.Name
	item.Description = body.Description
	item.Category = body.Category
	item.DefaultWidth = body.DefaultWidth
	item.DefaultHeight = body.DefaultHeight
	item.MinWidth = body.MinWidth
	item.MinHeight = body.MinHeight
	item.DefaultPriceGroupID = body.DefaultPriceGroupID
	item.IsActive = body.IsActive

	if dimsChanged || manifestChanged {
		item.SvgURL = body.SvgURL // re-bake from the supplied working structure svg
		if err := m.bakeAndStore(r.Context(), item, body.Manifest, body.ContentBBox); err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	} else {
		manifestMap, err := toMap(body.Manifest)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to encode manifest: %w", err))
			return
		}
		item.Manifest = manifestMap
	}

	if err := m.Db.CatalogItems.Update(item); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, item)
}

// bakeAndStore fetches the working structure SVG referenced by item.SvgURL,
// bakes it (fit + colored from the manifest), uploads the baked asset, and
// mutates the item to point at the baked URL with the stored manifest.
//
// Without S3 configured (e.g. tests) it stores the manifest only and leaves the
// svg_url as-is, so item creation still succeeds.
func (m *CatalogModule) bakeAndStore(ctx context.Context, item *data.CatalogItem, manifest svg.Manifest, bbox svg.ContentBBox) error {
	manifestMap, err := toMap(manifest)
	if err != nil {
		return fmt.Errorf("failed to encode manifest: %w", err)
	}
	item.Manifest = manifestMap

	if m.S3 == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	key := strings.TrimPrefix(item.SvgURL, "/")
	structureSVG, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, key)
	if err != nil {
		return fmt.Errorf("failed to fetch structure svg for bake: %w", err)
	}

	glassHexByID, groutHexByID, err := m.colorMaps()
	if err != nil {
		return err
	}

	baked, err := svg.Bake(structureSVG, manifest, bbox, item.DefaultWidth, item.DefaultHeight, svg.ColorOverrides{}, glassHexByID, groutHexByID)
	if err != nil {
		return fmt.Errorf("failed to bake catalog svg: %w", err)
	}

	result, err := upload.UploadFileToS3(
		ctx, m.S3, m.Cfg,
		bytes.NewReader(baked),
		item.CatalogCode+".svg",
		int64(len(baked)),
		"image/svg+xml",
		"catalog-items",
	)
	if err != nil {
		return fmt.Errorf("failed to upload baked svg: %w", err)
	}

	item.SvgURL = result.URL
	return nil
}

func (m *CatalogModule) colorMaps() (glass map[int]string, grout map[int]string, err error) {
	glassColors, err := m.Db.GlassColors.GetAllActive()
	if err != nil {
		return nil, nil, err
	}
	grouts, err := m.Db.Grouts.GetAllActive()
	if err != nil {
		return nil, nil, err
	}

	glass = make(map[int]string, len(glassColors))
	for _, gc := range glassColors {
		glass[gc.ID] = gc.Hex
	}
	grout = make(map[int]string, len(grouts))
	for _, g := range grouts {
		grout[g.ID] = g.Hex
	}
	return glass, grout, nil
}

// validateManifestAssigned rejects a manifest that still has unassigned grout or
// glass color ids — the bake step requires every region to resolve to a color.
func validateManifestAssigned(manifest svg.Manifest) error {
	if manifest.GroutRegion.GroutID == nil {
		return errors.New("manifest grout region has no assigned grout")
	}
	for key, region := range manifest.GlassRegions {
		if region.GlassColorID == nil {
			return fmt.Errorf("manifest glass group %s has no assigned glass color", key)
		}
	}
	return nil
}

func manifestEqual(stored map[string]interface{}, incoming svg.Manifest) bool {
	storedBytes, err := json.Marshal(stored)
	if err != nil {
		return false
	}
	var storedManifest svg.Manifest
	if err := json.Unmarshal(storedBytes, &storedManifest); err != nil {
		return false
	}
	incomingBytes, err := json.Marshal(incoming)
	if err != nil {
		return false
	}
	normalized, err := json.Marshal(storedManifest)
	if err != nil {
		return false
	}
	return bytes.Equal(normalized, incomingBytes)
}

func toMap(v any) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// HandleGetCatalogSVG streams a catalog item's canonical SVG bytes same-origin,
// so the customizer can fetch it as text without S3 CORS issues.
func (m *CatalogModule) HandleGetCatalogSVG(w http.ResponseWriter, r *http.Request) {
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

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	key := strings.TrimPrefix(item.SvgURL, "/")
	data, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, key)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

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

	item.IsActive = false
	err = m.Db.CatalogItems.Update(item)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

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

	filtered := make([]string, 0, len(item.Tags))
	for _, t := range item.Tags {
		if t != tag {
			filtered = append(filtered, t)
		}
	}

	m.WriteJSON(w, r, http.StatusOK, filtered)
}

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

	filtered := filterCatalogItemsWithTags(items, search, category, tags)

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

func (m *CatalogModule) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := m.Db.CatalogItems.GetCategories()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, categories)
}

func (m *CatalogModule) HandleGetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := m.Db.CatalogItems.GetAllTags()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, tags)
}

func filterCatalogItems(items []*data.CatalogItem, search, category, isActive string) []*data.CatalogItem {
	filtered := make([]*data.CatalogItem, 0, len(items))

	for _, item := range items {
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), searchLower) &&
				!strings.Contains(strings.ToLower(item.CatalogCode), searchLower) {
				continue
			}
		}

		if category != "" && item.Category != category {
			continue
		}

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
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), searchLower) &&
				!strings.Contains(strings.ToLower(item.CatalogCode), searchLower) {
				continue
			}
		}

		if category != "" && item.Category != category {
			continue
		}

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
