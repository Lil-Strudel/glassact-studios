package svg

import (
	"sort"
	"strings"

	"github.com/beevik/etree"
)

const illustratorNamespace = "http://ns.adobe.com/AdobeIllustrator/10.0/"

// artboardClipID is the <clipPath> holding the source artboard. The admin
// editor's bbox measurement (use-content-bbox.ts) reads it by this id.
const artboardClipID = "gac-artboard"

// Containers whose children are templates or resources, never drawn in place.
var nonRenderedContainers = map[string]bool{
	"defs":     true,
	"clippath": true,
	"mask":     true,
	"symbol":   true,
	"pattern":  true,
	"marker":   true,
}

func isNonRenderedContainer(tag string) bool {
	return nonRenderedContainers[strings.ToLower(localName(tag))]
}

// declaredProperty returns the value of a presentation property set on the
// element itself, in CSS precedence order: inline style, then a <style> class
// rule, then the presentation attribute. class is set when a class rule won.
func declaredProperty(el *etree.Element, prop string, rules []cssRule) (value string, class *string, ok bool) {
	if v, found := inlineValue(el.SelectAttrValue("style", ""), prop); found {
		return v, nil, true
	}
	if d, found := cascadeFor(el.SelectAttrValue("class", ""), rules)[prop]; found {
		c := d.class
		return d.value, &c, true
	}
	if attr := strings.TrimSpace(el.SelectAttrValue(prop, "")); attr != "" {
		return attr, nil, true
	}
	return "", nil, false
}

// removeHiddenElements deletes everything a renderer would not draw: any element
// with display:none (and its subtree), and any shape whose inherited visibility
// is hidden. Illustrator exports hidden layers, guides and reference photos this
// way, and left in place they become pieces, seed grout, inflate the measured
// bbox and reappear once Bake strips the CSS that hid them.
func removeHiddenElements(root *etree.Element, rules []cssRule) int {
	removed := 0
	var walk func(el *etree.Element)
	walk = func(el *etree.Element) {
		for _, child := range el.ChildElements() {
			if isHidden(child, rules) {
				el.RemoveChild(child)
				removed++
				continue
			}
			walk(child)
		}
	}
	walk(root)
	return removed
}

func isHidden(el *etree.Element, rules []cssRule) bool {
	if display, _, ok := declaredProperty(el, "display", rules); ok && strings.EqualFold(display, "none") {
		return true
	}
	if !isFillable(el.Tag) {
		return false
	}
	for node := el; node != nil; node = node.Parent() {
		visibility, _, ok := declaredProperty(node, "visibility", rules)
		if !ok {
			continue
		}
		visibility = strings.ToLower(visibility)
		return visibility == "hidden" || visibility == "collapse"
	}
	return false
}

// removeIllustratorData deletes Illustrator's private editing data: elements and
// attributes in its namespace, the <metadata> left empty by that, and the
// namespace declaration. With "Preserve Illustrator Editing Capabilities" on,
// the <i:aipgf> blob is a full copy of the .ai file and can be megabytes; no
// renderer reads it.
func removeIllustratorData(root *etree.Element) {
	var walk func(el *etree.Element)
	walk = func(el *etree.Element) {
		for _, child := range el.ChildElements() {
			if child.NamespaceURI() == illustratorNamespace {
				el.RemoveChild(child)
				continue
			}
			for _, attr := range append([]etree.Attr{}, child.Attr...) {
				if attr.NamespaceURI() == illustratorNamespace {
					child.RemoveAttr(attr.FullKey())
				}
			}
			hadChildren := len(child.ChildElements()) > 0
			walk(child)
			if hadChildren && isEmptyMetadata(child) {
				el.RemoveChild(child)
			}
		}
	}
	walk(root)

	for _, el := range append([]*etree.Element{root}, root.FindElements("//*")...) {
		for _, attr := range append([]etree.Attr{}, el.Attr...) {
			if attr.Space == "xmlns" && attr.Value == illustratorNamespace {
				el.RemoveAttr(attr.FullKey())
			}
		}
	}
}

func isEmptyMetadata(el *etree.Element) bool {
	return strings.EqualFold(localName(el.Tag), "metadata") &&
		len(el.ChildElements()) == 0 &&
		strings.TrimSpace(el.Text()) == ""
}

// clipToArtboard clips the artwork to the source viewBox. A browser crops
// anything drawn off the artboard, but Bake fits the browser-measured content
// bbox, which includes it: a stray point outside the artboard shrinks and
// off-centers the design. The clip rect is kept in the artwork's own coordinates
// (it moves into the gac-fit wrapper with everything else), so the editor can
// bound its measurement by it, and it keeps the baked output cropped like the
// source.
func clipToArtboard(root *etree.Element, viewBox string) {
	x, y, w, h, ok := parseViewBox(viewBox)
	if !ok {
		return
	}

	content := etree.NewElement("g")
	content.CreateAttr("clip-path", "url(#"+artboardClipID+")")
	for _, child := range root.ChildElements() {
		root.RemoveChild(child)
		content.AddChild(child)
	}

	clip := etree.NewElement("clipPath")
	clip.CreateAttr("id", artboardClipID)
	rect := clip.CreateElement("rect")
	rect.CreateAttr("x", formatNum(x))
	rect.CreateAttr("y", formatNum(y))
	rect.CreateAttr("width", formatNum(w))
	rect.CreateAttr("height", formatNum(h))

	root.AddChild(clip)
	root.AddChild(content)
}

// inlineClassStyles writes every element's class-rule declarations into its
// inline style and drops the class, so the document renders the same with no
// <style> at all. Ingest has to overwrite piece classes with group keys and Bake
// strips <style>, and either one alone loses whatever the source classes did
// beyond fill (strokes, fill:none outlines, opacity).
func inlineClassStyles(root *etree.Element, rules []cssRule) {
	var walk func(el *etree.Element)
	walk = func(el *etree.Element) {
		for _, child := range el.ChildElements() {
			inlineElementClassStyle(child, rules)
			walk(child)
		}
	}
	walk(root)
}

func inlineElementClassStyle(el *etree.Element, rules []cssRule) {
	classAttr := el.SelectAttr("class")
	if classAttr == nil {
		return
	}
	cascaded := cascadeFor(classAttr.Value, rules)
	el.RemoveAttr("class")
	if len(cascaded) == 0 {
		return
	}

	inline := parseDeclarations(el.SelectAttrValue("style", ""))
	for _, d := range inline {
		delete(cascaded, d.prop)
	}

	props := make([]string, 0, len(cascaded))
	for prop := range cascaded {
		props = append(props, prop)
	}
	sort.Strings(props)

	merged := make([]cssDecl, 0, len(props)+len(inline))
	for _, prop := range props {
		merged = append(merged, cssDecl{prop: prop, value: cascaded[prop].value})
	}
	merged = append(merged, inline...)
	el.CreateAttr("style", formatDeclarations(merged))
}
