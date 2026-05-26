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
  <image x="0" y="0" width="10" height="10" href="data:image/png;base64,AAAA"/>
</svg>`

const svgGradient = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs>
    <linearGradient id="g"><stop offset="0" stop-color="#000"/></linearGradient>
    <style>.st0{fill:#333333;}</style>
  </defs>
  <path class="st0" d="M0 0h1v1H0z"/>
</svg>`

const svgNoFills = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <defs><style>.s{fill:none;}</style></defs>
  <path class="s" d="M0 0h1v1H0z"/>
</svg>`

const svgNoViewBox = `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="480">
  <defs><style>.st0{fill:#abcdef;}</style></defs>
  <rect class="st0" x="0" y="0" width="2" height="2"/>
</svg>`

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
	canonical, manifest, q, err := Ingest([]byte(src))
	require.NoError(t, err)
	require.Nil(t, q, "expected no quarantine")
	require.NotNil(t, manifest)
	return canonical, manifest
}

func TestIngest_MultiClass_GroupsByColorAndAssignsStableIDs(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	assert.Equal(t, "0 0 100 200", manifest.ViewBox)
	require.Len(t, manifest.Regions, 2)

	grey := manifest.Regions["#7a8074"]
	assert.Equal(t, []string{"p0", "p1"}, grey.PieceIDs)
	assert.Equal(t, 2, grey.Count)
	require.NotNil(t, grey.Class)
	assert.Equal(t, "st0", *grey.Class)

	silver := manifest.Regions["#a7a9ac"]
	assert.Equal(t, []string{"p2"}, silver.PieceIDs)

	// canonical carries the injected ids and scales to its container
	require.NotNil(t, findByID(t, canonical, "p0"))
	require.NotNil(t, findByID(t, canonical, "p2"))
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(canonical))
	root := doc.SelectElement("svg")
	assert.Equal(t, "100%", root.SelectAttrValue("width", ""))
	assert.Equal(t, "100%", root.SelectAttrValue("height", ""))
}

func TestIngest_ClasslessShapeIsImplicitBlackRegion(t *testing.T) {
	_, manifest := ingestOK(t, svgImplicitBlack)
	require.Len(t, manifest.Regions, 2)

	black := manifest.Regions["#000000"]
	assert.Equal(t, []string{"p0"}, black.PieceIDs)
	assert.Nil(t, black.Class, "implicit-black region has no source class")

	assert.Equal(t, []string{"p1"}, manifest.Regions["#111111"].PieceIDs)
}

func TestIngest_FillNoneShapeIsNotARegion(t *testing.T) {
	canonical, manifest := ingestOK(t, svgFillNone)
	require.Len(t, manifest.Regions, 1)
	assert.Equal(t, []string{"p0"}, manifest.Regions["#222222"].PieceIDs)
	// the outline shape still received a stable id, it is just not recolorable
	assert.NotNil(t, findByID(t, canonical, "p1"))
}

func TestIngest_DerivesViewBoxFromWidthHeight(t *testing.T) {
	_, manifest := ingestOK(t, svgNoViewBox)
	assert.Equal(t, "0 0 640 480", manifest.ViewBox)
}

func TestIngest_Quarantine(t *testing.T) {
	cases := []struct {
		name   string
		src    string
		reason string
	}{
		{"embedded image", svgImage, "embedded raster image"},
		{"gradient", svgGradient, "gradient fill"},
		{"no recolorable fills", svgNoFills, "no recolorable fills"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, manifest, q, err := Ingest([]byte(tc.src))
			require.NoError(t, err)
			assert.Nil(t, manifest)
			require.NotNil(t, q)
			assert.Equal(t, tc.reason, q.Reason)
		})
	}
}

func TestBake_RegionRecolorAppliesToAllPiecesInGroup(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	overrides := ColorOverrides{
		Regions: map[string]GlassColorRef{"#7a8074": {GlassColorID: 5}},
	}
	out, err := Bake(canonical, *manifest, 1.0, overrides, map[int]string{5: "#ff0000"}, nil)
	require.NoError(t, err)

	for _, id := range []string{"p0", "p1"} {
		el := findByID(t, out, id)
		require.NotNil(t, el, id)
		assert.Contains(t, el.SelectAttrValue("style", ""), "fill:#ff0000")
		assert.Equal(t, "5", el.SelectAttrValue("data-glass-color-id", ""))
	}
	// untouched region keeps its source color (no inline fill)
	p2 := findByID(t, out, "p2")
	require.NotNil(t, p2)
	assert.NotContains(t, p2.SelectAttrValue("style", ""), "fill:")
}

func TestBake_PieceOverrideWinsOverRegion(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	overrides := ColorOverrides{
		Regions: map[string]GlassColorRef{"#7a8074": {GlassColorID: 5}},
		Pieces:  map[string]GlassColorRef{"p1": {GlassColorID: 7}},
	}
	out, err := Bake(canonical, *manifest, 1.0, overrides, map[int]string{5: "#ff0000", 7: "#00ff00"}, nil)
	require.NoError(t, err)

	assert.Contains(t, findByID(t, out, "p0").SelectAttrValue("style", ""), "fill:#ff0000")
	assert.Contains(t, findByID(t, out, "p1").SelectAttrValue("style", ""), "fill:#00ff00")
}

func TestBake_GroutRectInsertedBehind(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	overrides := ColorOverrides{Background: &GroutRef{GroutID: 3}}
	out, err := Bake(canonical, *manifest, 1.0, overrides, nil, map[int]string{3: "#cccccc"})
	require.NoError(t, err)

	grout := findByID(t, out, "glassact-grout")
	require.NotNil(t, grout)
	assert.Equal(t, "rect", grout.Tag)
	assert.Contains(t, grout.SelectAttrValue("style", ""), "fill:#cccccc")
	assert.Equal(t, "100", grout.SelectAttrValue("width", ""))
	assert.Equal(t, "200", grout.SelectAttrValue("height", ""))
}

func TestBake_AppliesScaleToRootDimensions(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	out, err := Bake(canonical, *manifest, 2.0, ColorOverrides{}, nil, nil)
	require.NoError(t, err)

	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(out))
	root := doc.SelectElement("svg")
	assert.Equal(t, "200", root.SelectAttrValue("width", ""))
	assert.Equal(t, "400", root.SelectAttrValue("height", ""))
}

func TestBake_EmbedsCutListMetadata(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	overrides := ColorOverrides{
		Regions:    map[string]GlassColorRef{"#7a8074": {GlassColorID: 5}},
		Background: &GroutRef{GroutID: 3},
	}
	out, err := Bake(canonical, *manifest, 1.0, overrides, map[int]string{5: "#ff0000"}, map[int]string{3: "#cccccc"})
	require.NoError(t, err)

	meta := findByID(t, out, "glassact-cutlist")
	require.NotNil(t, meta)
	assert.Equal(t, "metadata", meta.Tag)
	body := meta.Text()
	assert.Contains(t, body, "\"source_hex\":\"#7a8074\"")
	assert.Contains(t, body, "\"grout_id\":3")
}

func TestBake_UnknownGlassColorIDErrors(t *testing.T) {
	canonical, manifest := ingestOK(t, svgMultiClass)

	overrides := ColorOverrides{Regions: map[string]GlassColorRef{"#7a8074": {GlassColorID: 99}}}
	_, err := Bake(canonical, *manifest, 1.0, overrides, map[int]string{}, nil)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "99"))
}
