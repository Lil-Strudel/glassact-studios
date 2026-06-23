package svg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchNearest_PicksClosestColor(t *testing.T) {
	palette := []PaletteColor{
		{ID: 1, Hex: "#ff0000"},
		{ID: 2, Hex: "#00ff00"},
		{ID: 3, Hex: "#0000ff"},
	}

	// A slightly-off red should match id 1.
	got := MatchGlass("#fe0203", palette)
	require.NotNil(t, got)
	assert.Equal(t, 1, *got)
}

func TestMatchNearest_ExactMatch(t *testing.T) {
	palette := []PaletteColor{{ID: 9, Hex: "#7a8074"}, {ID: 10, Hex: "#a7a9ac"}}
	got := MatchGrout("#a7a9ac", palette)
	require.NotNil(t, got)
	assert.Equal(t, 10, *got)
}

func TestMatchNearest_NilPastThreshold(t *testing.T) {
	// Black palette, white source — far beyond the ΔE threshold.
	palette := []PaletteColor{{ID: 1, Hex: "#000000"}}
	got := MatchGlass("#ffffff", palette)
	assert.Nil(t, got)
}

func TestMatchNearest_EmptyPaletteOrBadHex(t *testing.T) {
	assert.Nil(t, MatchGlass("#abcdef", nil))
	assert.Nil(t, MatchGlass("not-a-hex", []PaletteColor{{ID: 1, Hex: "#abcdef"}}))
}

func TestDeltaE76_ZeroForIdentical(t *testing.T) {
	a, ok := hexToLab("#123456")
	require.True(t, ok)
	assert.InDelta(t, 0, deltaE76(a, a), 1e-9)
}
