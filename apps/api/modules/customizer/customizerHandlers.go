package customizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/svg"
)

type CustomizerModule struct {
	*app.Application
}

func NewCustomizerModule(app *app.Application) *CustomizerModule {
	return &CustomizerModule{app}
}

type bakeRequest struct {
	ScaleFactor    float64                `json:"scale_factor"`
	Width          float64                `json:"width" validate:"required,gt=0"`
	Height         float64                `json:"height" validate:"required,gt=0"`
	ColorOverrides map[string]interface{} `json:"color_overrides"`
}

type bakeResponse struct {
	DesignAssetURL string                 `json:"design_asset_url"`
	ColorOverrides map[string]interface{} `json:"color_overrides"`
	ScaleFactor    float64                `json:"scale_factor"`
	Width          float64                `json:"width"`
	Height         float64                `json:"height"`
}

// HandleBake renders a flat, self-contained SVG from a catalog item's canonical
// SVG + the supplied color overrides, uploads it to S3, and returns the asset
// URL. It creates no DB row — the future ordering flow persists these artifacts
// onto an inlay_proof.
func (m *CustomizerModule) HandleBake(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := m.Validate.Var(uuid, "required,uuid4"); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	item, found, err := m.Db.CatalogItems.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	var body bakeRequest
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var manifest svg.Manifest
	if err := remarshal(item.Manifest, &manifest); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to decode manifest: %w", err))
		return
	}
	if len(manifest.GlassRegions) == 0 {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("catalog item has no recolorable regions"))
		return
	}

	var overrides svg.ColorOverrides
	if err := remarshal(body.ColorOverrides, &overrides); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("invalid color_overrides: %w", err))
		return
	}

	glassHexByID, groutHexByID, err := m.colorMaps()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	key := strings.TrimPrefix(item.SvgURL, "/")
	structureSVG, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, key)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to fetch baked svg: %w", err))
		return
	}

	scaleFactor := body.ScaleFactor
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}

	baked, err := svg.BakeConsumer(structureSVG, manifest, scaleFactor, overrides, glassHexByID, groutHexByID)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	result, err := upload.UploadFileToS3(
		ctx, m.S3, m.Cfg,
		bytes.NewReader(baked),
		item.CatalogCode+"-baked.svg",
		int64(len(baked)),
		"image/svg+xml",
		"baked",
	)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to upload baked svg: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, bakeResponse{
		DesignAssetURL: result.URL,
		ColorOverrides: body.ColorOverrides,
		ScaleFactor:    scaleFactor,
		Width:          body.Width,
		Height:         body.Height,
	})
}

func (m *CustomizerModule) colorMaps() (glass map[int]string, grout map[int]string, err error) {
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

// remarshal converts between a generic JSONB map and a typed struct.
func remarshal(in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
