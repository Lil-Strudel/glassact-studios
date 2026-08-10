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

// sourceFill is a shape's fill as a renderer would resolve it.
type sourceFill struct {
	// key groups pieces that share a fill: the normalized hex, or the raw paint
	// string for paints that resolve to no color (url(#grad), currentColor).
	key string
	// hex is the normalized "#rrggbb" fill, empty for an unresolvable paint.
	hex string
	// class is the <style> class the fill came from, kept for provenance.
	class *string
	// implicit is true when no fill was declared anywhere up the tree, so the
	// shape renders UA-default black.
	implicit bool
}

// colorGroup is an intermediate, fill-keyed grouping built during ingest before
// stable group keys are assigned and the grout region is split off.
type colorGroup struct {
	key        string
	sourceHex  string
	sourceCls  *string
	implicit   bool
	pieceIDs   []string
	firstOrder int // document order of the group's first piece, for stable keys
}

// Ingest parses a raw catalog source SVG into a structure SVG (with stable
// per-shape ids p0, p1, ... and a group class on each recolorable piece) plus a
// manifest grouped into a single grout region and N glass regions.
//
// Grout is seeded from the back-most paintable shape: in document order that is
// the shape every other one is drawn on top of, i.e. the backing plate or the
// border ring. Its whole fill group becomes the grout region. Nothing is
// classified as grout for merely being black — a black glass detail sitting on a
// dark backing has to stay glass — and the manifest editor is the escape hatch
// when the seed guesses wrong.
//
// It best-guesses a grout_id / glass_color_id for each region from the supplied
// palettes, leaving any region with no declared fill or no close match unassigned
// and noted in warnings.
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

	// Group pieces by resolved fill, preserving document order.
	groups := map[string]*colorGroup{}
	groupOf := map[string]*colorGroup{}
	backMostID := ""
	pieceIndex := 0

	walkFillable(root, func(el *etree.Element) {
		id := fmt.Sprintf("p%d", pieceIndex)
		order := pieceIndex
		pieceIndex++
		el.CreateAttr("id", id)

		fill, paintable := resolveSourceFill(el, classFills)
		if !paintable {
			return // e.g. fill:none stroke outline — has an id but is not a region
		}
		if backMostID == "" {
			backMostID = id
		}

		g, ok := groups[fill.key]
		if !ok {
			g = &colorGroup{
				key:        fill.key,
				sourceHex:  fill.hex,
				sourceCls:  fill.class,
				implicit:   fill.implicit,
				firstOrder: order,
			}
			groups[fill.key] = g
		}
		if g.sourceCls == nil {
			g.sourceCls = fill.class
		}
		// A group only counts as undeclared if every piece in it is.
		if !fill.implicit {
			g.implicit = false
		}
		g.pieceIDs = append(g.pieceIDs, id)
		groupOf[id] = g
	})

	if pieceIndex == 0 {
		warnings = append(warnings, "no fillable shapes found in source")
	}
	if len(groups) == 0 {
		warnings = append(warnings, "no recolorable fills found in source")
	}

	// The back-most paintable shape is the backing plate, so its fill group is
	// the grout region.
	groutKey := ""
	if g := groupOf[backMostID]; g != nil {
		groutKey = g.key
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
	glassRegions := map[string]GlassRegion{}
	byID := indexByID(root)
	groupIndex := 0

	for _, g := range ordered {
		if g.key == groutKey {
			grout.PieceIDs = append(grout.PieceIDs, g.pieceIDs...)
			grout.Count = len(grout.PieceIDs)
			for _, id := range g.pieceIDs {
				if el := byID[id]; el != nil {
					el.CreateAttr("class", groutGroupKey)
				}
			}

			if g.sourceHex != "" {
				grout.GroutID = MatchGrout(g.sourceHex, groutPalette)
			}
			if grout.GroutID == nil {
				warnings = append(warnings, fmt.Sprintf("grout region (%s) has no close grout match", g.key))
			}
			// The one case the seed cannot decide: mortar lines drawn on top share
			// the backing's fill and are grout, a black eye shares it and is not.
			if grout.Count > 1 {
				warnings = append(warnings, fmt.Sprintf(
					"grout region seeded with %d pieces sharing %s — check none of them are glass details",
					grout.Count, g.key))
			}
			continue
		}

		key := fmt.Sprintf("group-%d", groupIndex)
		groupIndex++

		region := GlassRegion{
			PieceIDs:    append([]string{}, g.pieceIDs...),
			Count:       len(g.pieceIDs),
			SourceClass: g.sourceCls,
		}
		switch {
		case g.implicit:
			// The source is asking for a color here, not offering one: leave it
			// unassigned so the admin picks rather than defaulting it to black.
			warnings = append(warnings, fmt.Sprintf("glass group %s has no declared fill in the source — pick a color", key))
		case g.sourceHex == "":
			warnings = append(warnings, fmt.Sprintf("glass group %s uses an unsupported paint (%s) — pick a color", key, g.key))
		default:
			hex := g.sourceHex
			region.SourceHex = &hex
			region.GlassColorID = MatchGlass(hex, glassPalette)
			if region.GlassColorID == nil {
				warnings = append(warnings, fmt.Sprintf("glass group %s (%s) has no close color match", key, hex))
			}
		}
		glassRegions[key] = region

		for _, id := range g.pieceIDs {
			if el := byID[id]; el != nil {
				el.CreateAttr("class", key)
			}
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

// resolveSourceFill determines the fill a renderer would paint a shape with. It
// applies CSS precedence at each level (inline style, then a <style> class rule,
// then the fill presentation attribute) and walks up to the root for inherited
// fills. Returns paintable=false when the winning declaration is "none", e.g. a
// stroke-only outline.
func resolveSourceFill(el *etree.Element, classFills map[string]string) (sourceFill, bool) {
	for node := el; node != nil; node = node.Parent() {
		value, class, ok := declaredFill(node, classFills)
		if !ok {
			continue
		}
		if strings.EqualFold(value, "none") {
			return sourceFill{}, false
		}
		if hex, isColor := normalizeColor(value); isColor {
			return sourceFill{key: hex, hex: hex, class: class}, true
		}
		// A paint we cannot resolve to a color (url(#grad), currentColor). Keep
		// those pieces together under the raw value so the editor can assign them.
		return sourceFill{key: value, class: class}, true
	}
	// Nothing declared anywhere: renders UA-default black, and is recolorable.
	return sourceFill{key: defaultFill, implicit: true}, true
}

// declaredFill returns the fill declaration that wins on a single element, in CSS
// precedence order: inline style, then a <style> class rule, then the fill
// presentation attribute. Among several matching classes a color beats "none",
// so a shape carrying both a paint class and an outline class stays paintable.
func declaredFill(el *etree.Element, classFills map[string]string) (value string, class *string, ok bool) {
	if style := el.SelectAttrValue("style", ""); style != "" {
		if m := cssFillRe.FindStringSubmatch(style); m != nil {
			return strings.TrimSpace(m[1]), nil, true
		}
	}

	sawFillNone := false
	for _, c := range strings.Fields(el.SelectAttrValue("class", "")) {
		fv, found := classFills[c]
		if !found {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(fv), "none") {
			sawFillNone = true
			continue
		}
		cc := c
		return strings.TrimSpace(fv), &cc, true
	}
	if sawFillNone {
		return "none", nil, true
	}

	if attr := strings.TrimSpace(el.SelectAttrValue("fill", "")); attr != "" {
		return attr, nil, true
	}
	return "", nil, false
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
