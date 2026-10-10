package svg

import (
	"sort"
	"strings"

	"github.com/beevik/etree"
)

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
