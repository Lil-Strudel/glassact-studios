package svg

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fitDoc(t *testing.T) *etree.Element {
	t.Helper()
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
	  <rect id="p0" x="10" y="10" width="80" height="80"/>
	</svg>`))
	return doc.SelectElement("svg")
}

func TestApplyFit_SetsViewBoxToInchesTimes300(t *testing.T) {
	root := fitDoc(t)
	applyFit(root, ContentBBox{X: 10, Y: 10, Width: 80, Height: 80}, 3, 3)
	assert.Equal(t, "0 0 900 900", root.SelectAttrValue("viewBox", ""))
	assert.Empty(t, root.SelectAttrValue("width", ""))
	assert.Empty(t, root.SelectAttrValue("height", ""))
}

func TestApplyFit_WrapsContentAndCentersWithinPadding(t *testing.T) {
	root := fitDoc(t)
	applyFit(root, ContentBBox{X: 10, Y: 10, Width: 80, Height: 80}, 3, 3)

	var wrapper *etree.Element
	for _, child := range root.ChildElements() {
		if child.Tag == "g" && child.SelectAttrValue("id", "") == "gac-fit" {
			wrapper = child
		}
	}
	require.NotNil(t, wrapper, "content should be wrapped in gac-fit")
	require.NotNil(t, wrapper.FindElement("./rect"), "original piece moved under wrapper")

	// Square bbox into a square viewBox: centered, so translate components equal.
	ft := computeFit(3, 3, ContentBBox{X: 10, Y: 10, Width: 80, Height: 80})
	pad := 0.04 * 900.0
	assert.InDelta(t, (900-2*pad)/80, ft.Scale, 1e-9)
	// Centered horizontally and vertically.
	assert.InDelta(t, ft.TX, ft.TY, 1e-9)
	// Scaled+translated bbox stays within the padded box.
	left := ft.TX + 10*ft.Scale
	right := ft.TX + 90*ft.Scale
	assert.GreaterOrEqual(t, left, pad-1e-6)
	assert.LessOrEqual(t, right, 900-pad+1e-6)
}

func TestApplyFit_Idempotent(t *testing.T) {
	root := fitDoc(t)
	bbox := ContentBBox{X: 10, Y: 10, Width: 80, Height: 80}

	applyFit(root, bbox, 3, 3)
	doc := etree.NewDocument()
	doc.SetRoot(root.Copy())
	first, err := doc.WriteToString()
	require.NoError(t, err)

	applyFit(root, bbox, 3, 3)
	doc2 := etree.NewDocument()
	doc2.SetRoot(root.Copy())
	second, err := doc2.WriteToString()
	require.NoError(t, err)

	assert.Equal(t, first, second, "re-fitting must be idempotent")

	// Exactly one gac-fit wrapper after two applies.
	count := 0
	for _, child := range root.ChildElements() {
		if child.SelectAttrValue("id", "") == "gac-fit" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}
