package catalog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/svg"
)

type analyzeResponse struct {
	StructureSVG string       `json:"structure_svg"`
	Manifest     svg.Manifest `json:"manifest"`
	Warnings     []string     `json:"warnings"`
}

// HandleAnalyze ingests an uploaded source SVG into a working structure SVG plus
// a best-guess manifest (grout + glass groups with matched ids). It writes no DB
// row — the admin editor finalizes the manifest before POST/PUT.
func (m *CatalogModule) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SvgURL string `json:"svg_url" validate:"required,min=1"`
	}

	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if m.S3 == nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("file storage is not configured"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	key := strings.TrimPrefix(body.SvgURL, "/")
	raw, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, key)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to fetch source svg: %w", err))
		return
	}

	glassPalette, groutPalette, err := m.colorPalettes()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	structureSVG, manifest, warnings, err := svg.Ingest(raw, glassPalette, groutPalette)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("failed to analyze svg: %w", err))
		return
	}
	if warnings == nil {
		warnings = []string{}
	}

	m.WriteJSON(w, r, http.StatusOK, analyzeResponse{
		StructureSVG: string(structureSVG),
		Manifest:     *manifest,
		Warnings:     warnings,
	})
}

func (m *CatalogModule) colorPalettes() (glass []svg.PaletteColor, grout []svg.PaletteColor, err error) {
	glassColors, err := m.Db.GlassColors.GetAllActive()
	if err != nil {
		return nil, nil, err
	}
	grouts, err := m.Db.Grouts.GetAllActive()
	if err != nil {
		return nil, nil, err
	}

	glass = make([]svg.PaletteColor, len(glassColors))
	for i, gc := range glassColors {
		glass[i] = svg.PaletteColor{ID: gc.ID, Hex: gc.Hex}
	}
	grout = make([]svg.PaletteColor, len(grouts))
	for i, g := range grouts {
		grout[i] = svg.PaletteColor{ID: g.ID, Hex: g.Hex}
	}
	return glass, grout, nil
}
