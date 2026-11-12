package proof

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func twoPixelPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{B: 255, A: 255})
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestMirrorDesignAsset_PNG_FlipsPixelsLeftToRight(t *testing.T) {
	out, contentType, err := mirrorDesignAsset(twoPixelPNG(t), ".png", true, false)
	require.NoError(t, err)
	assert.Equal(t, "image/png", contentType)

	img, err := png.Decode(bytes.NewReader(out))
	require.NoError(t, err)
	r, _, b, _ := img.At(0, 0).RGBA()
	assert.Zero(t, r, "left pixel should now be the blue one")
	assert.NotZero(t, b)
}

func TestMirrorDesignAsset_SVG_WrapsInMirrorTransform(t *testing.T) {
	out, contentType, err := mirrorDesignAsset([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect/></svg>`), ".SVG", true, false)
	require.NoError(t, err)
	assert.Equal(t, "image/svg+xml", contentType)
	assert.Contains(t, string(out), `scale(-1 1)`)
}

func TestMirrorDesignAsset_PDF_ReportsUnmirrorable(t *testing.T) {
	_, _, err := mirrorDesignAsset([]byte("%PDF-1.7"), ".pdf", true, false)
	assert.ErrorIs(t, err, errUnmirrorableAsset)
}

func TestMirrorDirection_NamesEachCombination(t *testing.T) {
	assert.Equal(t, "left-right", mirrorDirection(true, false))
	assert.Equal(t, "top-bottom", mirrorDirection(false, true))
	assert.Equal(t, "left-right-and-top-bottom", mirrorDirection(true, true))
}
