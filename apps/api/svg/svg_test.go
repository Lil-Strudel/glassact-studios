package svg

import (
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const svgMultiClass = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 200">
  <defs><style>.st0{fill:#7A8074;}.st1{fill:#A7A9AC;}</style></defs>
  <path class="st0" d="M0 0h10v10H0z"/>
  <g><polygon class="st0" points="0,0 5,5 0,5"/><rect class="st1" x="1" y="1" width="2" height="2"/></g>
</svg>`

const svgImplicitBlack = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st0{fill:#111111;}</style></defs>
  <path d="M0 0h1v1H0z"/>
  <rect class="st0" x="0" y="0" width="2" height="2"/>
</svg>`

const svgFillNone = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st0{fill:#222222;}.outline{fill:none;stroke:#000;}</style></defs>
  <path class="st0" d="M0 0h1v1H0z"/>
  <polyline class="outline" points="0,0 5,5"/>
</svg>`

const svgImage = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.st0{fill:#333333;}</style></defs>
  <image x="0" y="0" width="10" height="10" href="data:image/png;base64,AAAA"/>
  <path class="st0" d="M0 0h1v1H0z"/>
</svg>`

const svgNoViewBox = `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="480">
  <defs><style>.st0{fill:#abcdef;}</style></defs>
  <rect class="st0" x="0" y="0" width="2" height="2"/>
</svg>`

// The first piece (p0, back-most) is the grout group; remaining are glass.
func findByID(t *testing.T, data []byte, id string) *etree.Element {
	t.Helper()
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(data))
	for _, el := range doc.FindElements("//*") {
		if el.SelectAttrValue("id", "") == id {
			return el
		}
	}
	return nil
}

func ingestOK(t *testing.T, src string) ([]byte, *Manifest) {
	t.Helper()
	structureSVG, manifest, _, err := Ingest([]byte(src), nil, nil)
	require.NoError(t, err)
	require.NotNil(t, manifest)
	return structureSVG, manifest
}

func TestIngest_MultiClass_GroupsByColorAndGroutCollapse(t *testing.T) {
	structureSVG, manifest := ingestOK(t, svgMultiClass)

	assert.Equal(t, "0 0 100 200", manifest.ViewBox)

	// p0 + p1 are the back-most #7a8074 group -> grout region.
	assert.ElementsMatch(t, []string{"p0", "p1"}, manifest.GroutRegion.PieceIDs)
	assert.Equal(t, 2, manifest.GroutRegion.Count)

	// The remaining #a7a9ac group becomes a single glass region.
	require.Len(t, manifest.GlassRegions, 1)
	silver := manifest.GlassRegions["group-0"]
	assert.Equal(t, []string{"p2"}, silver.PieceIDs)
	require.NotNil(t, silver.SourceHex)
	assert.Equal(t, "#a7a9ac", *silver.SourceHex)
	require.NotNil(t, silver.SourceClass)
	assert.Equal(t, "st1", *silver.SourceClass)

	// structure carries stable ids and the group class on pieces.
	assert.NotNil(t, findByID(t, structureSVG, "p0"))
	p2 := findByID(t, structureSVG, "p2")
	require.NotNil(t, p2)
	assert.Equal(t, "group-0", p2.SelectAttrValue("class", ""))
	assert.Equal(t, groutGroupKey, findByID(t, structureSVG, "p0").SelectAttrValue("class", ""))
}

func TestIngest_ClasslessShapeJoinsGrout(t *testing.T) {
	_, manifest := ingestOK(t, svgImplicitBlack)

	// p0 is classless implicit-black (back-most) -> grout.
	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs)

	// The #111111 group becomes the single glass region.
	require.Len(t, manifest.GlassRegions, 1)
	assert.Equal(t, []string{"p1"}, manifest.GlassRegions["group-0"].PieceIDs)
}

func TestIngest_FillNoneShapeIsNotARegion(t *testing.T) {
	structureSVG, manifest := ingestOK(t, svgFillNone)

	// p0 (#222222, back-most) collapses to grout; the fill:none outline (p1) is
	// not part of any region but still has a stable id.
	assert.Equal(t, []string{"p0"}, manifest.GroutRegion.PieceIDs)
	assert.Len(t, manifest.GlassRegions, 0)
	assert.NotNil(t, findByID(t, structureSVG, "p1"))
}

func TestIngest_DerivesViewBoxFromWidthHeight(t *testing.T) {
	_, manifest := ingestOK(t, svgNoViewBox)
	assert.Equal(t, "0 0 640 480", manifest.ViewBox)
}

func TestIngest_EmbeddedRasterWarnsButSucceeds(t *testing.T) {
	_, manifest, warnings, err := Ingest([]byte(svgImage), nil, nil)
	require.NoError(t, err)
	require.NotNil(t, manifest)
	assert.NotEmpty(t, warnings)
	found := false
	for _, wmsg := range warnings {
		if strings.Contains(wmsg, "raster") {
			found = true
		}
	}
	assert.True(t, found, "expected a raster warning")
}

func TestIngest_NoSVGRootHardErrors(t *testing.T) {
	_, _, _, err := Ingest([]byte("not svg at all"), nil, nil)
	require.Error(t, err)
}

func TestIngest_BestGuessMatchesPalette(t *testing.T) {
	glass := []PaletteColor{{ID: 42, Hex: "#a7a9ac"}}
	grout := []PaletteColor{{ID: 7, Hex: "#7a8074"}}
	_, manifest, _, err := Ingest([]byte(svgMultiClass), glass, grout)
	require.NoError(t, err)

	require.NotNil(t, manifest.GroutRegion.GroutID)
	assert.Equal(t, 7, *manifest.GroutRegion.GroutID)

	region := manifest.GlassRegions["group-0"]
	require.NotNil(t, region.GlassColorID)
	assert.Equal(t, 42, *region.GlassColorID)
}

func bakedManifest(t *testing.T, src string) (*Manifest, []byte) {
	t.Helper()
	structureSVG, manifest := ingestOK(t, src)
	return manifest, structureSVG
}

func TestBake_GroupRecolorAppliesToAllPiecesInGroup(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgMultiClass)

	overrides := ColorOverrides{
		Groups: map[string]GlassColorRef{"group-0": {GlassColorID: 5}},
	}
	bbox := ContentBBox{X: 0, Y: 0, Width: 100, Height: 200}
	out, err := Bake(structureSVG, *manifest, bbox, 1, 2, overrides,
		map[int]string{5: "#ff0000"}, nil)
	require.NoError(t, err)

	p2 := findByID(t, out, "p2")
	require.NotNil(t, p2)
	assert.Contains(t, p2.SelectAttrValue("style", ""), "fill:#ff0000")
	assert.Equal(t, "5", p2.SelectAttrValue("data-glass-color-id", ""))
}

func TestBake_PieceOverrideWinsOverGroup(t *testing.T) {
	// A design with two glass pieces in one group: split them with a piece override.
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
	  <defs><style>.bg{fill:#101010;}.g{fill:#a7a9ac;}</style></defs>
	  <rect class="bg" x="0" y="0" width="10" height="10"/>
	  <path class="g" d="M0 0h1v1H0z"/>
	  <path class="g" d="M2 2h1v1H2z"/>
	</svg>`
	manifest, structureSVG := bakedManifest(t, src)
	require.Len(t, manifest.GlassRegions, 1)

	overrides := ColorOverrides{
		Groups: map[string]GlassColorRef{"group-0": {GlassColorID: 5}},
		Pieces: map[string]GlassColorRef{"p2": {GlassColorID: 7}},
	}
	bbox := ContentBBox{X: 0, Y: 0, Width: 10, Height: 10}
	out, err := Bake(structureSVG, *manifest, bbox, 1, 1, overrides,
		map[int]string{5: "#00ff00", 7: "#0000ff"}, nil)
	require.NoError(t, err)

	assert.Contains(t, findByID(t, out, "p1").SelectAttrValue("style", ""), "fill:#00ff00")
	assert.Contains(t, findByID(t, out, "p2").SelectAttrValue("style", ""), "fill:#0000ff")
}

func TestBake_GroutRectInsertedAndSizedToFitViewBox(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgMultiClass)
	manifest.GroutRegion.GroutID = intPtr(3)

	bbox := ContentBBox{X: 0, Y: 0, Width: 100, Height: 200}
	out, err := Bake(structureSVG, *manifest, bbox, 3, 3, ColorOverrides{},
		nil, map[int]string{3: "#cccccc"})
	require.NoError(t, err)

	grout := findByID(t, out, "glassact-grout")
	require.NotNil(t, grout)
	assert.Equal(t, "rect", grout.Tag)
	assert.Contains(t, grout.SelectAttrValue("style", ""), "fill:#cccccc")
	// 3in x 3in -> 900 x 900 viewBox.
	assert.Equal(t, "900", grout.SelectAttrValue("width", ""))
	assert.Equal(t, "900", grout.SelectAttrValue("height", ""))
}

func TestBake_StripsStyleAndSetsFitViewBox(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgMultiClass)
	manifest.GroutRegion.GroutID = intPtr(3)

	bbox := ContentBBox{X: 0, Y: 0, Width: 100, Height: 200}
	out, err := Bake(structureSVG, *manifest, bbox, 3, 3, ColorOverrides{},
		map[int]string{}, map[int]string{3: "#cccccc"})
	require.NoError(t, err)

	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(out))
	assert.Empty(t, doc.FindElements("//style"), "style blocks must be stripped")
	root := doc.SelectElement("svg")
	assert.Equal(t, "0 0 900 900", root.SelectAttrValue("viewBox", ""))
	require.NotNil(t, findByID(t, out, "p0"), "pieces stay re-editable")
}

func TestBake_EmbedsCutListMetadata(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgMultiClass)
	manifest.GroutRegion.GroutID = intPtr(3)

	overrides := ColorOverrides{Groups: map[string]GlassColorRef{"group-0": {GlassColorID: 5}}}
	bbox := ContentBBox{X: 0, Y: 0, Width: 100, Height: 200}
	out, err := Bake(structureSVG, *manifest, bbox, 1, 2, overrides,
		map[int]string{5: "#ff0000"}, map[int]string{3: "#cccccc"})
	require.NoError(t, err)

	meta := findByID(t, out, "glassact-cutlist")
	require.NotNil(t, meta)
	assert.Equal(t, "metadata", meta.Tag)
	body := meta.Text()
	assert.Contains(t, body, "\"group_key\":\"group-0\"")
	assert.Contains(t, body, "\"grout_id\":3")
}

func TestBake_UnknownGlassColorIDErrors(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgMultiClass)

	overrides := ColorOverrides{Groups: map[string]GlassColorRef{"group-0": {GlassColorID: 99}}}
	bbox := ContentBBox{X: 0, Y: 0, Width: 100, Height: 200}
	_, err := Bake(structureSVG, *manifest, bbox, 1, 2, overrides, map[int]string{}, nil)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "99"))
}

func TestBakeConsumer_KeepsManifestViewBoxAndScales(t *testing.T) {
	manifest, structureSVG := bakedManifest(t, svgMultiClass)

	out, err := BakeConsumer(structureSVG, *manifest, 2.0, ColorOverrides{}, nil, nil)
	require.NoError(t, err)

	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(out))
	root := doc.SelectElement("svg")
	assert.Equal(t, "200", root.SelectAttrValue("width", ""))
	assert.Equal(t, "400", root.SelectAttrValue("height", ""))
}

func intPtr(v int) *int { return &v }
