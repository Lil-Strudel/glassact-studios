package svg

import "math"

// matchThreshold is the maximum ΔE76 distance for a source hex to be considered
// a close-enough match to a palette color. Above this, no guess is made and the
// region is left unassigned (and surfaced as a warning).
const matchThreshold = 25.0

// PaletteColor is one candidate color (a glass color or a grout) to match
// against, identified by its database id.
type PaletteColor struct {
	ID  int
	Hex string
}

// labColor is a color in CIELAB space.
type labColor struct {
	L, A, B float64
}

// hexToLab converts a normalized "#rrggbb" hex string to CIELAB. ok is false for
// unparseable input.
func hexToLab(hex string) (labColor, bool) {
	h, ok := normalizeHex(hex)
	if !ok {
		return labColor{}, false
	}
	r := hexByte(h[1], h[2])
	g := hexByte(h[3], h[4])
	b := hexByte(h[5], h[6])
	return rgbToLab(r, g, b), true
}

func hexByte(hi, lo byte) int {
	return hexDigit(hi)*16 + hexDigit(lo)
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}

// rgbToLab converts 8-bit sRGB to CIELAB using the D65 white point.
func rgbToLab(r, g, b int) labColor {
	rl := srgbToLinear(float64(r) / 255.0)
	gl := srgbToLinear(float64(g) / 255.0)
	bl := srgbToLinear(float64(b) / 255.0)

	// Linear sRGB -> CIE XYZ (D65).
	x := rl*0.4124564 + gl*0.3575761 + bl*0.1804375
	y := rl*0.2126729 + gl*0.7151522 + bl*0.0721750
	z := rl*0.0193339 + gl*0.1191920 + bl*0.9503041

	// Normalize by the D65 reference white.
	const xn, yn, zn = 0.95047, 1.00000, 1.08883
	fx := labF(x / xn)
	fy := labF(y / yn)
	fz := labF(z / zn)

	return labColor{
		L: 116*fy - 16,
		A: 500 * (fx - fy),
		B: 200 * (fy - fz),
	}
}

func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func labF(t float64) float64 {
	const delta = 6.0 / 29.0
	if t > delta*delta*delta {
		return math.Cbrt(t)
	}
	return t/(3*delta*delta) + 4.0/29.0
}

// deltaE76 is the Euclidean distance between two Lab colors.
func deltaE76(a, b labColor) float64 {
	dl := a.L - b.L
	da := a.A - b.A
	db := a.B - b.B
	return math.Sqrt(dl*dl + da*da + db*db)
}

// matchNearest returns the id of the palette color closest to sourceHex by ΔE76,
// or nil if the source is unparseable, the palette is empty, or the minimum
// distance exceeds matchThreshold.
func matchNearest(sourceHex string, palette []PaletteColor) *int {
	src, ok := hexToLab(sourceHex)
	if !ok {
		return nil
	}

	bestDist := math.Inf(1)
	bestID := 0
	found := false
	for _, c := range palette {
		lab, ok := hexToLab(c.Hex)
		if !ok {
			continue
		}
		d := deltaE76(src, lab)
		if d < bestDist {
			bestDist = d
			bestID = c.ID
			found = true
		}
	}

	if !found || bestDist > matchThreshold {
		return nil
	}
	id := bestID
	return &id
}

// MatchGlass returns the best-guess glass_color_id for a source hex, or nil if
// no palette color is within the match threshold.
func MatchGlass(sourceHex string, palette []PaletteColor) *int {
	return matchNearest(sourceHex, palette)
}

// MatchGrout returns the best-guess grout_id for a source hex, or nil if no
// palette color is within the match threshold.
func MatchGrout(sourceHex string, palette []PaletteColor) *int {
	return matchNearest(sourceHex, palette)
}
