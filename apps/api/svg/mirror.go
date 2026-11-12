package svg

import (
	"errors"
	"fmt"

	"github.com/beevik/etree"
)

const mirrorWrapperID = "gac-mirror"

// nonRenderedTags stay at the root when the artwork is mirrored, so metadata
// readers find the cutlist where they always have.
var nonRenderedTags = map[string]bool{
	"defs": true, "metadata": true, "style": true, "title": true, "desc": true,
}

// Mirror flips the artwork left-to-right (mirrorX) and/or top-to-bottom
// (mirrorY) within its own viewBox, so the file keeps its size and position.
// Like applyFit, the transform lives on a wrapper and path data is untouched.
func Mirror(in []byte, mirrorX, mirrorY bool) ([]byte, error) {
	if !mirrorX && !mirrorY {
		return in, nil
	}

	doc, root, err := parseRoot(in)
	if err != nil {
		return nil, err
	}

	minX, minY, w, h, err := artworkBox(root)
	if err != nil {
		return nil, err
	}

	translateX, scaleX := 0.0, 1.0
	if mirrorX {
		translateX, scaleX = 2*minX+w, -1
	}
	translateY, scaleY := 0.0, 1.0
	if mirrorY {
		translateY, scaleY = 2*minY+h, -1
	}

	wrapper := etree.NewElement("g")
	wrapper.CreateAttr("id", mirrorWrapperID)
	wrapper.CreateAttr("transform", fmt.Sprintf("translate(%s %s) scale(%s %s)",
		formatNum(translateX), formatNum(translateY), formatNum(scaleX), formatNum(scaleY)))

	for _, child := range root.ChildElements() {
		if nonRenderedTags[localName(child.Tag)] {
			continue
		}
		root.RemoveChild(child)
		wrapper.AddChild(child)
	}
	root.AddChild(wrapper)

	return doc.WriteToBytes()
}

// artworkBox is the coordinate box the artwork is drawn in: the viewBox, or
// the width/height when an uploaded file has no viewBox.
func artworkBox(root *etree.Element) (minX, minY, w, h float64, err error) {
	if x, y, vw, vh, ok := parseViewBox(root.SelectAttrValue("viewBox", "")); ok {
		return x, y, vw, vh, nil
	}

	w = parseDim(root.SelectAttrValue("width", ""))
	h = parseDim(root.SelectAttrValue("height", ""))
	if w <= 0 || h <= 0 {
		return 0, 0, 0, 0, errors.New("svg has neither a viewBox nor a width and height to mirror within")
	}
	return 0, 0, w, h, nil
}
