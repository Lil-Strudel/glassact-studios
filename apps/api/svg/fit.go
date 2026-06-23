package svg

import (
	"fmt"

	"github.com/beevik/etree"
)

// unitsPerInch is the canonical SVG coordinate density of a baked design.
const unitsPerInch = 300.0

// fitTransform describes the translate+scale that fits a content bounding box,
// centered with padding, into a width x height target viewBox.
type fitTransform struct {
	W, H  float64 // target viewBox dimensions (inches * unitsPerInch)
	Scale float64
	TX    float64
	TY    float64
}

// computeFit returns the transform that fits bbox into a width x height inch
// target (converted to units), centered with a 4% padding of the shorter side.
func computeFit(width, height float64, bbox ContentBBox) fitTransform {
	w := width * unitsPerInch
	h := height * unitsPerInch
	pad := 0.04 * minFloat(w, h)

	scale := 1.0
	if bbox.Width > 0 && bbox.Height > 0 {
		scale = minFloat((w-2*pad)/bbox.Width, (h-2*pad)/bbox.Height)
	}

	tx := (w-bbox.Width*scale)/2 - bbox.X*scale
	ty := (h-bbox.Height*scale)/2 - bbox.Y*scale

	return fitTransform{W: w, H: h, Scale: scale, TX: tx, TY: ty}
}

// applyFit fits the artwork into a width x height inch viewBox. It is idempotent:
// an existing <g id="gac-fit"> wrapper is unwrapped first so re-fitting from a
// previously baked SVG produces the same result. The transform lives only on the
// wrapper — path coordinate data is never mutated.
func applyFit(root *etree.Element, bbox ContentBBox, width, height float64) {
	unwrapFit(root)

	ft := computeFit(width, height, bbox)

	wrapper := etree.NewElement("g")
	wrapper.CreateAttr("id", "gac-fit")
	wrapper.CreateAttr("transform", fmt.Sprintf("translate(%s %s) scale(%s)",
		formatNum(ft.TX), formatNum(ft.TY), formatNum(ft.Scale)))

	// Move all current element children into the wrapper, preserving order.
	for _, child := range root.ChildElements() {
		root.RemoveChild(child)
		wrapper.AddChild(child)
	}
	root.AddChild(wrapper)

	root.CreateAttr("viewBox", fmt.Sprintf("0 0 %s %s", formatNum(ft.W), formatNum(ft.H)))
	root.RemoveAttr("width")
	root.RemoveAttr("height")
}

// unwrapFit moves the children of any <g id="gac-fit"> wrapper back up to the
// root and removes the wrapper, restoring the pre-fit structure.
func unwrapFit(root *etree.Element) {
	for _, child := range root.ChildElements() {
		if child.Tag == "g" && child.SelectAttrValue("id", "") == "gac-fit" {
			idx := child.Index()
			inner := child.ChildElements()
			for i, grandchild := range inner {
				child.RemoveChild(grandchild)
				root.InsertChildAt(idx+i, grandchild)
			}
			root.RemoveChild(child)
		}
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
