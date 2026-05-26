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

// Region is one recolorable color group of a design, keyed in the manifest by
// the design's original source hex. Class is the source CSS class that produced
// the color (nil for implicit-black, classless shapes).
type Region struct {
	Class    *string  `json:"class"`
	PieceIDs []string `json:"piece_ids"`
	Count    int      `json:"count"`
}

// Manifest is emitted by Ingest and stored on catalog_items.manifest. The
// customizer renders its UI from it.
type Manifest struct {
	ViewBox string            `json:"view_box"`
	Regions map[string]Region `json:"regions"`
}

type GlassColorRef struct {
	GlassColorID int `json:"glass_color_id"`
}

type GroutRef struct {
	GroutID int `json:"grout_id"`
}

// ColorOverrides is the durable changelist. Resolution order at bake is
// piece override -> region mapping -> original source hex.
type ColorOverrides struct {
	Regions    map[string]GlassColorRef `json:"regions,omitempty"`
	Pieces     map[string]GlassColorRef `json:"pieces,omitempty"`
	Background *GroutRef                `json:"background,omitempty"`
}

// Quarantine signals a source SVG the customizer can't handle. Ingest returns
// one instead of a manifest; the catalog item is flagged, not rejected.
type Quarantine struct {
	Reason string `json:"reason"`
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
