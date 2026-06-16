// Package svg ingests catalog inlay source SVGs into a stable-id canonical SVG
// plus a region manifest, and bakes a flat self-contained SVG from a set of
// color overrides. It is pure logic (no HTTP, no DB) so it can be unit-tested.
package svg

import (
	"regexp"
	"strconv"
	"strings"
)

// Fillable SVG shape elements. Each becomes a stable, recolorable piece.
var fillableTags = map[string]bool{
	"path":     true,
	"polygon":  true,
	"rect":     true,
	"circle":   true,
	"ellipse":  true,
	"polyline": true,
	"line":     true,
}

// A shape with no resolvable class fill renders as UA-default black, and is a
// real recolorable region.
const defaultFill = "#000000"

// GroutRegion is the single collapsed grout region: the back-most group plus
// classless implicit-black pieces. GroutID is nil until assigned in the editor.
type GroutRegion struct {
	GroutID  *int     `json:"grout_id"`
	PieceIDs []string `json:"piece_ids"`
	Count    int      `json:"count"`
}

// GlassRegion is one recolorable glass color group, keyed in the manifest by a
// stable group key (e.g. "group-0"). SourceClass/SourceHex are provenance from
// the source SVG (display + best-guess matching only).
type GlassRegion struct {
	GlassColorID *int     `json:"glass_color_id"`
	PieceIDs     []string `json:"piece_ids"`
	Count        int      `json:"count"`
	SourceClass  *string  `json:"source_class,omitempty"`
	SourceHex    *string  `json:"source_hex,omitempty"`
}

// Manifest is emitted by Ingest and stored on catalog_items.manifest. The
// customizer renders its UI from it. GlassRegions are keyed by stable group key.
type Manifest struct {
	ViewBox      string                 `json:"view_box"`
	GroutRegion  GroutRegion            `json:"grout_region"`
	GlassRegions map[string]GlassRegion `json:"glass_regions"`
}

type GlassColorRef struct {
	GlassColorID int `json:"glass_color_id"`
}

type GroutRef struct {
	GroutID int `json:"grout_id"`
}

// ColorOverrides is the durable changelist. Resolution order at bake is
// piece override -> group override -> manifest group default. Groups is keyed by
// the stable group key; Pieces is keyed by stable piece id (p0, p1, ...).
type ColorOverrides struct {
	Groups     map[string]GlassColorRef `json:"groups,omitempty"`
	Pieces     map[string]GlassColorRef `json:"pieces,omitempty"`
	Background *GroutRef                `json:"background,omitempty"`
}

// ContentBBox is the browser-measured content bounding box of the structure SVG,
// used to recompute the viewBox (300 units/inch) and fit+center artwork at bake.
type ContentBBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

var (
	cssRuleRe = regexp.MustCompile(`([^{}]+)\{([^}]*)\}`)
	cssFillRe = regexp.MustCompile(`(?i)fill\s*:\s*([^;]+)`)
	hex6Re    = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	hex3Re    = regexp.MustCompile(`^#[0-9a-fA-F]{3}$`)
	numRe     = regexp.MustCompile(`-?\d*\.?\d+`)
	// matches a `fill: ...;` declaration inside an inline style attribute
	inlineFillRe = regexp.MustCompile(`(?i)\s*fill\s*:[^;]*;?`)
)

// parseStyleFills maps each CSS class to its declared fill value (raw string).
// Generic on purpose: it does not assume the `.st` prefix (the glass palette
// file uses `.cls-`).
func parseStyleFills(css string) map[string]string {
	out := map[string]string{}
	for _, rule := range cssRuleRe.FindAllStringSubmatch(css, -1) {
		selectors, body := rule[1], rule[2]
		fm := cssFillRe.FindStringSubmatch(body)
		if fm == nil {
			continue
		}
		fill := strings.TrimSpace(fm[1])
		for _, sel := range strings.Split(selectors, ",") {
			sel = strings.TrimSpace(sel)
			if strings.HasPrefix(sel, ".") && !strings.ContainsAny(sel, " >+~") {
				out[strings.TrimPrefix(sel, ".")] = fill
			}
		}
	}
	return out
}

// normalizeHex lowercases and expands shorthand hex; returns ok=false for
// non-hex values like "none" or "url(#grad)".
func normalizeHex(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if hex6Re.MatchString(s) {
		return s, true
	}
	if hex3Re.MatchString(s) {
		return "#" + string([]byte{s[1], s[1], s[2], s[2], s[3], s[3]}), true
	}
	return "", false
}

// parseDim extracts the leading number from a dimension like "1200", "1200px"
// or "12.5in".
func parseDim(s string) float64 {
	m := numRe.FindString(s)
	if m == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(m, 64)
	return f
}

func formatNum(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// localName strips any namespace prefix from an element tag (e.g. "svg:path").
func localName(tag string) string {
	if i := strings.IndexByte(tag, ':'); i >= 0 {
		return tag[i+1:]
	}
	return tag
}

func isFillable(tag string) bool {
	return fillableTags[strings.ToLower(localName(tag))]
}
