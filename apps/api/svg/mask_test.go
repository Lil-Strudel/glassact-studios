package svg

import (
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const illustratorBackground = `<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:i="http://ns.adobe.com/AdobeIllustrator/10.0/" viewBox="0 0 200 100" width="200" height="100">
  <style>.gold{fill:#c9a227}.line{fill:none;stroke:#336699;stroke-width:2}</style>
  <rect class="gold" x="20" y="10" width="160" height="80"/>
  <path class="line" d="M30 50 L170 50"/>
  <circle cx="100" cy="50" r="10" fill="url(#shine)" i:knockout="Off"/>
  <g style="display:none"><rect x="0" y="0" width="5" height="5" fill="red"/></g>
  <image href="photo.png" x="0" y="0" width="10" height="10"/>
</svg>`

func parseMask(t *testing.T, mask []byte) *etree.Element {
	t.Helper()
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(mask))
	root := doc.SelectElement("svg")
	require.NotNil(t, root)
	return root
}

func TestPrepareMask_FlattensEveryPaintToBlackButKeepsOutlinesHollow(t *testing.T) {
	mask, _, err := PrepareMask([]byte(illustratorBackground))
	require.NoError(t, err)

	root := parseMask(t, mask)
	assert.Empty(t, root.FindElements("//style"), "class styles are inlined, not kept")

	rect := root.FindElement("//rect[@x='20']")
	require.NotNil(t, rect)
	fill, _ := inlineValue(rect.SelectAttrValue("style", ""), "fill")
	assert.Equal(t, "#000000", fill)

	line := root.FindElement("//path")
	require.NotNil(t, line)
	lineFill, _ := inlineValue(line.SelectAttrValue("style", ""), "fill")
	lineStroke, _ := inlineValue(line.SelectAttrValue("style", ""), "stroke")
	assert.Equal(t, "none", lineFill)
	assert.Equal(t, "#000000", lineStroke)

	circle := root.FindElement("//circle")
	require.NotNil(t, circle)
	assert.Equal(t, "#000000", circle.SelectAttrValue("fill", ""), "gradient paint becomes solid")
}

func TestPrepareMask_DropsHiddenLayersRastersAndIllustratorData(t *testing.T) {
	mask, warnings, err := PrepareMask([]byte(illustratorBackground))
	require.NoError(t, err)

	text := string(mask)
	assert.NotContains(t, text, "photo.png")
	assert.NotContains(t, text, "knockout")
	assert.NotContains(t, text, `fill="red"`)
	assert.Contains(t, strings.Join(warnings, "\n"), "raster")
	assert.Contains(t, strings.Join(warnings, "\n"), "hidden")
}

func TestPrepareMask_IsDeterministic(t *testing.T) {
	first, _, err := PrepareMask([]byte(illustratorBackground))
	require.NoError(t, err)
	second, _, err := PrepareMask([]byte(illustratorBackground))
	require.NoError(t, err)
	assert.Equal(t, string(first), string(second))
}

func TestPrepareMask_RejectsNonSVG(t *testing.T) {
	_, _, err := PrepareMask([]byte("<html><body/></html>"))
	assert.Error(t, err)
}

func TestFitMask_TrimsViewBoxToPaddedContent(t *testing.T) {
	fitted, err := FitMask([]byte(illustratorBackground), ContentBBox{X: 20, Y: 10, Width: 160, Height: 80})
	require.NoError(t, err)

	root := parseMask(t, fitted)
	assert.Equal(t, "16.8 6.8 166.4 86.4", root.SelectAttrValue("viewBox", ""))
	assert.Empty(t, root.SelectAttrValue("width", ""))
	assert.Empty(t, root.SelectAttrValue("height", ""))
}

func TestFitMask_RejectsEmptyContentBox(t *testing.T) {
	_, err := FitMask([]byte(illustratorBackground), ContentBBox{})
	assert.Error(t, err)
}
