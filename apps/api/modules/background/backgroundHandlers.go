package background

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/svg"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type BackgroundModule struct {
	*app.Application
}

func NewBackgroundModule(app *app.Application) *BackgroundModule {
	return &BackgroundModule{app}
}

type backgroundWriteBody struct {
	Code                  string           `json:"code" validate:"required,min=1,max=255"`
	Name                  string           `json:"name" validate:"required,min=1,max=255"`
	Description           *string          `json:"description"`
	Tags                  []string         `json:"tags" validate:"max=50,dive,max=100"`
	IsActive              bool             `json:"is_active"`
	SvgURL                string           `json:"svg_url" validate:"required,min=1"`
	ContentBBox           *svg.ContentBBox `json:"content_bbox"`
	RelatedCatalogItemIDs []int            `json:"related_catalog_item_ids" validate:"max=100,dive,gt=0"`
}

func (m BackgroundModule) getBackground(w http.ResponseWriter, r *http.Request) (*data.SandblastBackground, bool) {
	backgroundUUID := r.PathValue("uuid")
	if err := m.Validate.Var(backgroundUUID, "required,uuid4"); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return nil, false
	}

	background, found, err := m.Db.SandblastBackgrounds.GetByUUID(backgroundUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, false
	}
	return background, true
}

func (m BackgroundModule) writeSaveError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, data.ErrEditConflict) {
		m.WriteError(w, r, m.Err.Conflict, errors.New("the background was changed elsewhere; reload to see the latest version"))
		return
	}
	m.WriteError(w, r, m.Err.ServerError, err)
}

func (m BackgroundModule) applyWriteBody(w http.ResponseWriter, r *http.Request, background *data.SandblastBackground, body backgroundWriteBody) bool {
	code := strings.TrimSpace(body.Code)
	if err := m.ensureCodeIsFree(code, background.ID); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return false
	}

	related, err := m.relatedItemsFromIDs(body.RelatedCatalogItemIDs, background.RelatedItems)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return false
	}

	if body.SvgURL != background.SvgURL {
		if body.ContentBBox == nil {
			m.WriteError(w, r, m.Err.BadRequest, errors.New("content_bbox is required with a new svg"))
			return false
		}
		storedURL, err := m.storeMask(r.Context(), code, body.SvgURL, *body.ContentBBox)
		if err != nil {
			m.WriteError(w, r, m.Err.BadRequest, err)
			return false
		}
		background.SvgURL = storedURL
	}

	background.Code = code
	background.Name = strings.TrimSpace(body.Name)
	background.Description = body.Description
	background.Tags = normalizeTags(body.Tags)
	background.IsActive = body.IsActive
	background.RelatedItems = related
	return true
}

func (m BackgroundModule) HandleGetBackgrounds(w http.ResponseWriter, r *http.Request) {
	backgrounds, err := m.Db.SandblastBackgrounds.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	m.WriteJSON(w, r, http.StatusOK, backgrounds)
}

func (m BackgroundModule) HandleBrowseBackgrounds(w http.ResponseWriter, r *http.Request) {
	backgrounds, err := m.Db.SandblastBackgrounds.GetAllActive()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	offered := make([]*data.SandblastBackground, len(backgrounds))
	for i, background := range backgrounds {
		offered[i] = background.WithoutRetiredItems()
	}
	m.WriteJSON(w, r, http.StatusOK, offered)
}

func (m BackgroundModule) HandleGetBackground(w http.ResponseWriter, r *http.Request) {
	background, ok := m.getBackground(w, r)
	if !ok {
		return
	}
	m.WriteJSON(w, r, http.StatusOK, background)
}

// HandlePrepareBackground returns the cleaned mask for an uploaded SVG without
// storing anything, so the admin form can preview it and measure its content.
func (m BackgroundModule) HandlePrepareBackground(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SvgURL string `json:"svg_url" validate:"required,min=1"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if m.S3 == nil {
		m.WriteError(w, r, m.Err.ServerError, errors.New("file storage is not configured"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	raw, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, strings.TrimPrefix(body.SvgURL, "/"))
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to fetch uploaded svg: %w", err))
		return
	}

	mask, warnings, err := svg.PrepareMask(raw)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("failed to read svg: %w", err))
		return
	}
	if warnings == nil {
		warnings = []string{}
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]any{
		"mask_svg": string(mask),
		"warnings": warnings,
	})
}

func (m BackgroundModule) HandlePostBackground(w http.ResponseWriter, r *http.Request) {
	var body backgroundWriteBody
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	background := &data.SandblastBackground{}
	if !m.applyWriteBody(w, r, background, body) {
		return
	}

	if err := m.Db.SandblastBackgrounds.Insert(background); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	saved, _, err := m.Db.SandblastBackgrounds.GetByUUID(background.UUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	m.WriteJSON(w, r, http.StatusCreated, saved)
}

func (m BackgroundModule) HandlePutBackground(w http.ResponseWriter, r *http.Request) {
	var body struct {
		backgroundWriteBody
		Version int `json:"version" validate:"required,gt=0"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	background, ok := m.getBackground(w, r)
	if !ok {
		return
	}

	if !m.applyWriteBody(w, r, background, body.backgroundWriteBody) {
		return
	}
	background.Version = body.Version

	if err := m.Db.SandblastBackgrounds.Update(background); err != nil {
		m.writeSaveError(w, r, err)
		return
	}

	saved, _, err := m.Db.SandblastBackgrounds.GetByUUID(background.UUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	m.WriteJSON(w, r, http.StatusOK, saved)
}

func (m BackgroundModule) HandlePatchBackground(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IsActive *bool `json:"is_active" validate:"required"`
		Version  int   `json:"version" validate:"required,gt=0"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	background, ok := m.getBackground(w, r)
	if !ok {
		return
	}

	background.IsActive = *body.IsActive
	background.Version = body.Version
	if err := m.Db.SandblastBackgrounds.UpdateFields(background); err != nil {
		m.writeSaveError(w, r, err)
		return
	}
	m.WriteJSON(w, r, http.StatusOK, background)
}

// HandleGetBackgroundSVG streams the mask same-origin, so the stone designer
// can inline it without S3 CORS issues.
func (m BackgroundModule) HandleGetBackgroundSVG(w http.ResponseWriter, r *http.Request) {
	background, ok := m.getBackground(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	svgBytes, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, strings.TrimPrefix(background.SvgURL, "/"))
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svgBytes)
}

// HandleGetBackgroundSandblast hands out the mask as a download named by its
// code, so the file is recognisable in the dealership's sandblasting software.
func (m BackgroundModule) HandleGetBackgroundSandblast(w http.ResponseWriter, r *http.Request) {
	background, ok := m.getBackground(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	filename := upload.SanitizeFilenamePart(background.Code, "background") + ".svg"
	url, err := upload.GenerateSignedDownloadURL(ctx, m.S3, m.Cfg, strings.TrimPrefix(background.SvgURL, "/"), filename, 15*time.Minute)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]string{"url": url})
}
