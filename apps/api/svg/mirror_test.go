package svg

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mirrorInput = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="10 20 100 50">
  <metadata id="glassact-cutlist">{}</metadata>
  <rect id="p0" x="10" y="20" width="30" height="50"/>
</svg>`

func mirroredRoot(t *testing.T, mirrorX, mirrorY bool) *etree.Element {
	t.Helper()
	out, err := Mirror([]byte(mirrorInput), mirrorX, mirrorY)
	require.NoError(t, err)
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(out))
	return doc.SelectElement("svg")
}

func mirrorWrapper(root *etree.Element) *etree.Element {
	for _, child := range root.ChildElements() {
		if child.Tag == "g" && child.SelectAttrValue("id", "") == mirrorWrapperID {
			return child
		}
	}
	return nil
}

func TestMirror_LeftToRight_FlipsAcrossViewBoxCenter(t *testing.T) {
	root := mirroredRoot(t, true, false)
	wrapper := mirrorWrapper(root)
	require.NotNil(t, wrapper)
	assert.Equal(t, "translate(120 0) scale(-1 1)", wrapper.SelectAttrValue("transform", ""))
	assert.NotNil(t, wrapper.FindElement("./rect"), "artwork moved under the wrapper")
	assert.Equal(t, "10 20 100 50", root.SelectAttrValue("viewBox", ""), "file keeps its size")
}

func TestMirror_TopToBottom_FlipsAcrossViewBoxCenter(t *testing.T) {
	wrapper := mirrorWrapper(mirroredRoot(t, false, true))
	require.NotNil(t, wrapper)
	assert.Equal(t, "translate(0 90) scale(1 -1)", wrapper.SelectAttrValue("transform", ""))
}

func TestMirror_Both_FlipsBothAxes(t *testing.T) {
	wrapper := mirrorWrapper(mirroredRoot(t, true, true))
	require.NotNil(t, wrapper)
	assert.Equal(t, "translate(120 90) scale(-1 -1)", wrapper.SelectAttrValue("transform", ""))
}

func TestMirror_KeepsMetadataAtRoot(t *testing.T) {
	root := mirroredRoot(t, true, false)
	assert.NotNil(t, root.FindElement("./metadata[@id='glassact-cutlist']"))
}

func TestMirror_Neither_ReturnsInputUnchanged(t *testing.T) {
	out, err := Mirror([]byte(mirrorInput), false, false)
	require.NoError(t, err)
	assert.Equal(t, mirrorInput, string(out))
}

func TestMirror_NoViewBox_UsesWidthAndHeight(t *testing.T) {
	out, err := Mirror([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="20"><rect/></svg>`), true, false)
	require.NoError(t, err)
	assert.Contains(t, string(out), `transform="translate(40 0) scale(-1 1)"`)
}

func TestMirror_NoDimensions_ReturnsError(t *testing.T) {
	_, err := Mirror([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`), true, false)
	assert.Error(t, err)
}
