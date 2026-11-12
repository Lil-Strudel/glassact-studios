package svg

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

const maskPaint = "#000000"

// maskPaddingRatio is the margin FitMask leaves around the content, as a share
// of its longer side, so strokes on the edge are not shaved off.
const maskPaddingRatio = 0.02

// PrepareMask turns an uploaded sandblast background into a plain black mask:
// the same cleanup ingest does (Illustrator data, hidden layers, class styles)
// with every paint flattened to black, raster images dropped and the artwork
// clipped to its artboard. It is deterministic, so the admin editor can
// measure the prepared text and the server can rebuild the identical document
// from the raw upload at save time.
func PrepareMask(raw []byte) (mask []byte, warnings []string, err error) {
	doc, _, warnings, err := prepareMaskDoc(raw)
	if err != nil {
		return nil, nil, err
	}
	mask, err = doc.WriteToBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("serialize mask svg: %w", err)
	}
	return mask, warnings, nil
}

// FitMask prepares the mask and trims its viewBox to the browser-measured
// content box, so a background placed at a width fills that width rather than
// carrying the artboard's empty margins.
func FitMask(raw []byte, bbox ContentBBox) ([]byte, error) {
	if bbox.Width <= 0 || bbox.Height <= 0 {
		return nil, fmt.Errorf("content box must have a positive size")
	}

	doc, root, _, err := prepareMaskDoc(raw)
	if err != nil {
		return nil, err
	}

	pad := maskPaddingRatio * max(bbox.Width, bbox.Height)
	root.CreateAttr("viewBox", fmt.Sprintf("%s %s %s %s",
		formatNum(bbox.X-pad), formatNum(bbox.Y-pad),
		formatNum(bbox.Width+2*pad), formatNum(bbox.Height+2*pad)))
	root.RemoveAttr("width")
	root.RemoveAttr("height")

	out, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("serialize fitted mask svg: %w", err)
	}
	return out, nil
}

func prepareMaskDoc(raw []byte) (*etree.Document, *etree.Element, []string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return nil, nil, nil, fmt.Errorf("parse svg: %w", err)
	}
	root := doc.SelectElement("svg")
	if root == nil {
		return nil, nil, nil, fmt.Errorf("svg has no root element")
	}

	var warnings []string
	rules, unsupportedSelectors := parseClassRules(collectStyleCSS(doc))
	if len(unsupportedSelectors) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"source CSS uses selectors other than plain classes (%s); their styles were ignored",
			strings.Join(unsupportedSelectors, ", ")))
	}

	removeIllustratorData(root)
	if removed := removeHiddenElements(root, rules); removed > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"%d hidden element(s) (hidden layers, guides or reference images) were removed", removed))
	}

	if images := doc.FindElements("//image"); len(images) > 0 {
		for _, image := range images {
			image.Parent().RemoveChild(image)
		}
		warnings = append(warnings, "embedded raster images cannot be sandblasted and were removed")
	}

	inlineClassStyles(root, rules)
	stripStyles(root)
	flattenPaint(root)

	shapes := 0
	walkFillable(root, func(*etree.Element) { shapes++ })
	if shapes == 0 {
		warnings = append(warnings, "no shapes found in source")
	}

	clipToArtboard(root, ensureViewBox(root))
	return doc, root, warnings, nil
}

// flattenPaint paints every fill and stroke black, leaving "none" alone so
// outlines stay outlines. A gradient reference becomes solid black too: a
// sandblaster only knows blasted and not blasted.
func flattenPaint(root *etree.Element) {
	for _, el := range root.FindElements("//*") {
		for _, prop := range []string{"fill", "stroke"} {
			if value := el.SelectAttrValue(prop, ""); value != "" && !strings.EqualFold(value, "none") {
				el.CreateAttr(prop, maskPaint)
			}
		}
		if style := el.SelectAttr("style"); style != nil {
			decls := parseDeclarations(style.Value)
			for i, d := range decls {
				if (d.prop == "fill" || d.prop == "stroke") && !strings.EqualFold(d.value, "none") {
					decls[i].value = maskPaint
				}
			}
			el.CreateAttr("style", formatDeclarations(decls))
		}
	}
}
