package svg

import (
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Illustrator keeps hidden layers in the export behind a display:none class
// (F-TMP-0014's grey background, G-AIR-0002R's reference photo frame).
const svgHiddenLayer = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st7{fill:#999;}.st8{display:none;}.st0{fill:#5f4d40;}.st1{fill:#c1cdca;}</style></defs>
  <g id="Layer_6" class="st8"><rect class="st7" x="-5" y="-5" width="20" height="20"/></g>
  <path class="st0" d="M0 0h10v10H0z"/>
  <path class="st1" d="M1 1h2v2H1z"/>
</svg>`

// G-CNS-0003 hides individual shapes rather than a layer, with display:none
// shared across their fill classes.
const svgShapeHiddenByOwnClass = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st0{fill:#be8f8d;}.st0,.st1{display:none;}.st1{fill:#dfb9be;}.st3{fill:#7a8074;}.st5{fill:#c1cdca;}</style></defs>
  <polygon class="st0" points="-9,0 -5,-4 -9,-4"/>
  <path class="st1" d="M-8 1h2v2h-2z"/>
  <path class="st3" d="M0 0h10v10H0z"/>
  <path class="st5" d="M1 1h2v2H1z"/>
</svg>`

const svgHiddenRaster = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st5{display:none;}.st0{fill:#ae6219;}</style></defs>
  <g class="st5"><image width="10" height="10" href="data:image/png;base64,AAAA"/></g>
  <path class="st0" d="M0 0h10v10H0z"/>
</svg>`

const svgVisibilityHidden = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <path fill="#5f4d40" d="M0 0h10v10H0z"/>
  <g visibility="hidden">
    <path fill="#ff0000" d="M1 1h2v2H1z"/>
    <path fill="#c1cdca" visibility="visible" d="M4 1h2v2H4z"/>
  </g>
</svg>`

// E-BOR-0008: the border ring is an 8px-stroked ellipse with fill:none, drawn
// last, on top of every glass piece.
const svgStrokeOnlyOutline = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 60">
  <defs><style>.st0{fill:none;stroke:#000;stroke-miterlimit:10;stroke-width:8px;}.st1{fill:#4d4941;}</style></defs>
  <path d="M0 0h100v60H0z"/>
  <path class="st1" d="M10 10h20v20H10z"/>
  <ellipse class="st0" cx="50" cy="30" rx="45" ry="25"/>
</svg>`

// E-CRN-0006 splits an outline's declarations across two rules.
const svgRulesSplitAcrossBlocks = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>
    .st0 { fill: #59595c; }
    .st1 { stroke: #221f1f; }
    .st1, .st2 { fill: none; stroke-miterlimit: 10; }
    .st2 { stroke: #babcbe; }
  </style></defs>
  <path class="st0" d="M0 0h10v10H0z"/>
  <path class="st1" d="M1 1h2v2H1z"/>
  <path class="st2" d="M4 1h2v2H4z"/>
</svg>`

const svgShapeInClipPath = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><clipPath id="clip"><rect x="0" y="0" width="5" height="5"/></clipPath></defs>
  <path fill="#5f4d40" d="M0 0h10v10H0z"/>
  <path fill="#c1cdca" clip-path="url(#clip)" d="M1 1h2v2H1z"/>
</svg>`

func cascadeSVG(css string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>` + css + `.bg{fill:#5f4d40;}</style></defs>
  <path class="bg" d="M0 0h10v10H0z"/>
  <path class="a b" d="M1 1h2v2H1z"/>
</svg>`
}

func styleOf(t *testing.T, data []byte, id string) map[string]string {
	t.Helper()
	el := findByID(t, data, id)
	require.NotNil(t, el, "piece %s missing", id)
	out := map[string]string{}
	for _, d := range parseDeclarations(el.SelectAttrValue("style", "")) {
		out[d.prop] = d.value
	}
	return out
}

func TestIngest_HiddenLayerIsRemoved(t *testing.T) {
	structureSVG, manifest, warnings, err := Ingest([]byte(svgHiddenLayer), nil, nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs, "grout seeds from the first visible shape")
	require.Len(t, manifest.GlassRegions, 1)
	assert.Equal(t, []string{"p1"}, manifest.GlassRegions["group-0"].PieceIDs)
	assert.NotContains(t, string(structureSVG), "Layer_6")
	assert.NotContains(t, string(structureSVG), "#999")
	assert.True(t, containsWarning(warnings, "hidden"), "the admin is told hidden content was dropped")
}

func TestIngest_ShapeHiddenByOwnClassIsRemoved(t *testing.T) {
	structureSVG, manifest := ingestOK(t, svgShapeHiddenByOwnClass)

	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs)
	require.Len(t, manifest.GlassRegions, 1)
	assert.Equal(t, []string{"p1"}, manifest.GlassRegions["group-0"].PieceIDs)
	assert.Equal(t, "#7a8074", styleOf(t, structureSVG, "p0")["fill"])
	assert.NotContains(t, string(structureSVG), "polygon")
}

func TestIngest_HiddenRasterIsRemovedWithoutWarning(t *testing.T) {
	structureSVG, _, warnings, err := Ingest([]byte(svgHiddenRaster), nil, nil)
	require.NoError(t, err)

	assert.NotContains(t, string(structureSVG), "<image")
	assert.False(t, containsWarning(warnings, "raster"), "a hidden raster is not part of the design")
}

func TestIngest_InheritedVisibilityHiddenRemovesShape(t *testing.T) {
	structureSVG, manifest := ingestOK(t, svgVisibilityHidden)

	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs)
	require.Len(t, manifest.GlassRegions, 1)
	region := manifest.GlassRegions["group-0"]
	require.NotNil(t, region.SourceHex)
	assert.Equal(t, "#c1cdca", *region.SourceHex, "a child that overrides visibility stays")
	assert.NotContains(t, string(structureSVG), "#ff0000")
}

func TestBake_StrokeOnlyShapeKeepsStrokeAndNoFill(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgStrokeOnlyOutline)

	out, err := Bake(structureSVG, *manifest, ContentBBox{Width: 100, Height: 60}, 1, 0.6,
		ColorOverrides{}, map[int]string{}, map[int]string{})
	require.NoError(t, err)

	style := styleOf(t, out, "p2")
	assert.Equal(t, "none", style["fill"], "an outline must not fall back to UA-default black")
	assert.Equal(t, "#000", style["stroke"])
	assert.Equal(t, "8px", style["stroke-width"])
}

func TestIngest_ClassRulesMergeAcrossRules(t *testing.T) {
	structureSVG, manifest := ingestOK(t, svgRulesSplitAcrossBlocks)

	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs)
	assert.Empty(t, manifest.GlassRegions, "both outlines are fill:none")
	assert.Equal(t, map[string]string{"fill": "none", "stroke": "#221f1f", "stroke-miterlimit": "10"},
		styleOf(t, structureSVG, "p1"))
	assert.Equal(t, map[string]string{"fill": "none", "stroke": "#babcbe", "stroke-miterlimit": "10"},
		styleOf(t, structureSVG, "p2"))
}

func TestIngest_LaterClassRuleWinsInCascade(t *testing.T) {
	_, fillNoneLast := ingestOK(t, cascadeSVG(`.a{fill:#ff0000;}.b{fill:none;}`))
	assert.Empty(t, fillNoneLast.GlassRegions, "fill:none comes later, so the shape is an outline")

	_, colorLast := ingestOK(t, cascadeSVG(`.b{fill:none;}.a{fill:#ff0000;}`))
	require.Len(t, colorLast.GlassRegions, 1)
	region := colorLast.GlassRegions["group-0"]
	require.NotNil(t, region.SourceClass)
	assert.Equal(t, "a", *region.SourceClass)
}

func TestIngest_InlineStyleBeatsClassRuleWhenInlined(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st0{fill:#5f4d40;stroke:#000;}</style></defs>
  <path class="st0" style="stroke:#fff" d="M0 0h10v10H0z"/>
</svg>`
	structureSVG, _ := ingestOK(t, src)
	assert.Equal(t, map[string]string{"fill": "#5f4d40", "stroke": "#fff"}, styleOf(t, structureSVG, "p0"))
}

func TestIngest_StructureSVGIsSelfContained(t *testing.T) {
	for name, src := range map[string]string{
		"multi class":        svgMultiClass,
		"hidden layer":       svgHiddenLayer,
		"stroke-only":        svgStrokeOnlyOutline,
		"split across rules": svgRulesSplitAcrossBlocks,
	} {
		t.Run(name, func(t *testing.T) {
			structureSVG, _ := ingestOK(t, src)

			doc := etree.NewDocument()
			require.NoError(t, doc.ReadFromBytes(structureSVG))
			assert.Empty(t, doc.FindElements("//style"))
			for _, el := range doc.FindElements("//*[@class]") {
				class := el.SelectAttrValue("class", "")
				assert.True(t, class == groutGroupKey || strings.HasPrefix(class, "group-"),
					"source class %q survived on <%s>", class, el.Tag)
			}
		})
	}
}

func TestBake_UnmatchedGroupKeepsSourceColor(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgHiddenLayer)
	require.Nil(t, manifest.GlassRegions["group-0"].GlassColorID)

	out, err := Bake(structureSVG, *manifest, ContentBBox{Width: 10, Height: 10}, 1, 1,
		ColorOverrides{}, map[int]string{}, map[int]string{})
	require.NoError(t, err)

	assert.Equal(t, "#c1cdca", styleOf(t, out, "p1")["fill"])
}

func TestIngest_ShapesInsideDefsAreNotPieces(t *testing.T) {
	structureSVG, manifest := ingestOK(t, svgShapeInClipPath)

	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs)
	require.Len(t, manifest.GlassRegions, 1)
	assert.Equal(t, []string{"p1"}, manifest.GlassRegions["group-0"].PieceIDs)
	assert.Contains(t, string(structureSVG), "<clipPath", "clip paths stay referenceable")
}

func containsWarning(warnings []string, substr string) bool {
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}
