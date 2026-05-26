package svg

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// Ingest parses a raw catalog source SVG into a canonical SVG (with stable
// per-shape ids p0, p1, ... injected) and a region manifest grouped by source
// color. If the SVG can't be customized (embedded raster, gradient/pattern
// fills, or nothing recolorable) it returns a non-nil Quarantine and no
// manifest; the caller flags the catalog item rather than rejecting it.
func Ingest(raw []byte) (canonical []byte, manifest *Manifest, quarantine *Quarantine, err error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return nil, nil, nil, fmt.Errorf("parse svg: %w", err)
	}

	root := doc.SelectElement("svg")
	if root == nil {
		root = doc.Root()
	}
	if root == nil {
		return nil, nil, &Quarantine{Reason: "no <svg> root element"}, nil
	}

	if len(doc.FindElements("//image")) > 0 {
		return nil, nil, &Quarantine{Reason: "embedded raster image"}, nil
	}
	if len(doc.FindElements("//linearGradient")) > 0 || len(doc.FindElements("//radialGradient")) > 0 {
		return nil, nil, &Quarantine{Reason: "gradient fill"}, nil
	}

	classFills := parseStyleFills(collectStyleCSS(doc))
	for _, v := range classFills {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "url(") {
			return nil, nil, &Quarantine{Reason: "gradient or pattern fill"}, nil
		}
	}

	regions := map[string]*Region{}
	pieceIndex := 0
	sawFillable := false

	walkFillable(root, func(el *etree.Element) {
		sawFillable = true
		id := fmt.Sprintf("p%d", pieceIndex)
		pieceIndex++
		el.CreateAttr("id", id)

		hex, classPtr, paintable := resolveSourceFill(el, classFills)
		if !paintable {
			return // e.g. fill:none stroke outline — assigned an id but not a region
		}

		region, ok := regions[hex]
		if !ok {
			region = &Region{Class: classPtr, PieceIDs: []string{}}
			regions[hex] = region
		}
		region.PieceIDs = append(region.PieceIDs, id)
		region.Count++
	})

	if !sawFillable || len(regions) == 0 {
		return nil, nil, &Quarantine{Reason: "no recolorable fills"}, nil
	}

	viewBox := ensureViewBox(root)
	// Make the canonical SVG scale to fill its container in the customizer.
	root.CreateAttr("width", "100%")
	root.CreateAttr("height", "100%")
	if root.SelectAttrValue("preserveAspectRatio", "") == "" {
		root.CreateAttr("preserveAspectRatio", "xMidYMid meet")
	}

	manifest = &Manifest{ViewBox: viewBox, Regions: map[string]Region{}}
	for hex, region := range regions {
		manifest.Regions[hex] = *region
	}

	canonical, err = doc.WriteToBytes()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("serialize canonical svg: %w", err)
	}
	return canonical, manifest, nil, nil
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
