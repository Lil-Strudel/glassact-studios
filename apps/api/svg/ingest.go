package svg

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beevik/etree"
)

// groutGroupKey is the DOM class applied to every piece classified as grout, so
// the browser editor/customizer can map grout pieces from the structure SVG.
const groutGroupKey = "grout"

// colorGroup is an intermediate, source-hex-keyed grouping built during ingest
// before stable group keys are assigned and grout collapse happens.
type colorGroup struct {
	sourceHex  string
	sourceCls  *string
	pieceIDs   []string
	firstOrder int // document order of the group's first piece, for stable keys
}

// Ingest parses a raw catalog source SVG into a structure SVG (with stable
// per-shape ids p0, p1, ... and a group class on each recolorable piece) plus a
// manifest grouped into a single grout region and N glass regions. It best-
// guesses a grout_id / glass_color_id for each region from the supplied
// palettes, leaving any region with no close match unassigned and noted in
// warnings.
//
// It only hard-errors on a genuinely unparseable SVG or a missing <svg> root.
// Embedded raster, gradients and missing fills are surfaced as warnings rather
// than rejected — the manifest editor handles fixing them.
func Ingest(raw []byte, glassPalette, groutPalette []PaletteColor) (structureSVG []byte, manifest *Manifest, warnings []string, err error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return nil, nil, nil, fmt.Errorf("parse svg: %w", err)
	}

	root := doc.SelectElement("svg")
	if root == nil {
		root = doc.Root()
	}
	if root == nil {
		return nil, nil, nil, fmt.Errorf("svg has no root element")
	}

	if len(doc.FindElements("//image")) > 0 {
		warnings = append(warnings, "source contains an embedded raster image; it cannot be recolored and was ignored")
	}
	if len(doc.FindElements("//linearGradient")) > 0 || len(doc.FindElements("//radialGradient")) > 0 {
		warnings = append(warnings, "source contains gradient fills; affected pieces fall back to a solid color")
	}

	classFills := parseStyleFills(collectStyleCSS(doc))

	// Group pieces by resolved source hex, preserving document order.
	groups := map[string]*colorGroup{}
	pieceIndex := 0

	walkFillable(root, func(el *etree.Element) {
		id := fmt.Sprintf("p%d", pieceIndex)
		order := pieceIndex
		pieceIndex++
		el.CreateAttr("id", id)

		hex, classPtr, paintable := resolveSourceFill(el, classFills)
		if !paintable {
			return // e.g. fill:none stroke outline — has an id but is not a region
		}

		g, ok := groups[hex]
		if !ok {
			g = &colorGroup{sourceHex: hex, sourceCls: classPtr, firstOrder: order}
			groups[hex] = g
		}
		g.pieceIDs = append(g.pieceIDs, id)
	})

	if pieceIndex == 0 {
		warnings = append(warnings, "no fillable shapes found in source")
	}
	if len(groups) == 0 {
		warnings = append(warnings, "no recolorable fills found in source")
	}

	// Identify the grout hexes: the implicit-black group plus the group that
	// owns p0 (the back-most shape in document order).
	groutHexes := map[string]bool{}
	if _, ok := groups[defaultFill]; ok {
		groutHexes[defaultFill] = true
	}
	for hex, g := range groups {
		for _, id := range g.pieceIDs {
			if id == "p0" {
				groutHexes[hex] = true
			}
		}
	}

	// Split into grout vs glass groups, ordered deterministically by the
	// document order of each group's first piece.
	ordered := make([]*colorGroup, 0, len(groups))
	for _, g := range groups {
		ordered = append(ordered, g)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].firstOrder < ordered[j].firstOrder
	})

	grout := GroutRegion{PieceIDs: []string{}}
	var groutSourceHex string
	glassRegions := map[string]GlassRegion{}
	byID := indexByID(root)
	groupIndex := 0

	for _, g := range ordered {
		if groutHexes[g.sourceHex] {
			grout.PieceIDs = append(grout.PieceIDs, g.pieceIDs...)
			grout.Count += len(g.pieceIDs)
			if groutSourceHex == "" {
				groutSourceHex = g.sourceHex
			}
			for _, id := range g.pieceIDs {
				if el := byID[id]; el != nil {
					el.CreateAttr("class", groutGroupKey)
				}
			}
			continue
		}

		key := fmt.Sprintf("group-%d", groupIndex)
		groupIndex++

		hex := g.sourceHex
		region := GlassRegion{
			PieceIDs:    append([]string{}, g.pieceIDs...),
			Count:       len(g.pieceIDs),
			SourceClass: g.sourceCls,
			SourceHex:   &hex,
		}
		region.GlassColorID = MatchGlass(hex, glassPalette)
		if region.GlassColorID == nil {
			warnings = append(warnings, fmt.Sprintf("glass group %s (%s) has no close color match", key, hex))
		}
		glassRegions[key] = region

		for _, id := range g.pieceIDs {
			if el := byID[id]; el != nil {
				el.CreateAttr("class", key)
			}
		}
	}

	if groutSourceHex != "" {
		grout.GroutID = MatchGrout(groutSourceHex, groutPalette)
		if grout.GroutID == nil {
			warnings = append(warnings, fmt.Sprintf("grout region (%s) has no close grout match", groutSourceHex))
		}
	}

	// Preserve the original viewBox; fit happens at bake.
	viewBox := ensureViewBox(root)

	manifest = &Manifest{
		ViewBox:      viewBox,
		GroutRegion:  grout,
		GlassRegions: glassRegions,
	}

	structureSVG, err = doc.WriteToBytes()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("serialize structure svg: %w", err)
	}
	return structureSVG, manifest, warnings, nil
}

// resolveSourceFill determines a shape's source color. Returns paintable=false
// for shapes whose class explicitly sets fill:none (stroke-only outlines).
func resolveSourceFill(el *etree.Element, classFills map[string]string) (hex string, class *string, paintable bool) {
	classAttr := strings.TrimSpace(el.SelectAttrValue("class", ""))
	sawFillNone := false

	if classAttr != "" {
		for _, c := range strings.Fields(classAttr) {
			fv, ok := classFills[c]
			if !ok {
				continue
			}
			if h, isHex := normalizeHex(fv); isHex {
				cc := c
				return h, &cc, true
			}
			if strings.EqualFold(strings.TrimSpace(fv), "none") {
				sawFillNone = true
			}
		}
	}

	if sawFillNone {
		return "", nil, false
	}
	// No class, or a class with no fill rule: UA-default black, recolorable.
	return defaultFill, nil, true
}

func walkFillable(el *etree.Element, visit func(*etree.Element)) {
	for _, child := range el.ChildElements() {
		if isFillable(child.Tag) {
			visit(child)
		}
		walkFillable(child, visit)
	}
}

func collectStyleCSS(doc *etree.Document) string {
	var b strings.Builder
	for _, style := range doc.FindElements("//style") {
		b.WriteString(style.Text())
		b.WriteByte('\n')
	}
	return b.String()
}

// ensureViewBox returns the root viewBox, deriving one from width/height when
// absent, and falling back to a unit box if neither is available.
func ensureViewBox(root *etree.Element) string {
	if vb := strings.TrimSpace(root.SelectAttrValue("viewBox", "")); vb != "" {
		return vb
	}
	w := parseDim(root.SelectAttrValue("width", ""))
	h := parseDim(root.SelectAttrValue("height", ""))
	if w > 0 && h > 0 {
		vb := fmt.Sprintf("0 0 %s %s", formatNum(w), formatNum(h))
		root.CreateAttr("viewBox", vb)
		return vb
	}
	vb := "0 0 100 100"
	root.CreateAttr("viewBox", vb)
	return vb
}
