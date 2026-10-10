package svg

import (
	"regexp"
	"strings"
)

var (
	cssCommentRe    = regexp.MustCompile(`(?s)/\*.*?\*/`)
	classSelectorRe = regexp.MustCompile(`^\.(-?[_a-zA-Z][-_a-zA-Z0-9]*)$`)
)

type cssDecl struct {
	prop  string
	value string
}

// cssRule is one source rule narrowed to the class selectors it applies to.
type cssRule struct {
	classes []string
	decls   []cssDecl
}

// classDecl is a property's winning value from the class cascade, with the class
// whose rule supplied it.
type classDecl struct {
	value string
	class string
}

// parseClassRules parses <style> CSS into rules in source order. Only bare class
// selectors are understood, which is all Illustrator emits; any other selector is
// returned so ingest can say it was ignored rather than silently render wrong.
func parseClassRules(css string) (rules []cssRule, unsupportedSelectors []string) {
	css = cssCommentRe.ReplaceAllString(css, "")
	for _, match := range cssRuleRe.FindAllStringSubmatch(css, -1) {
		decls := parseDeclarations(match[2])
		if len(decls) == 0 {
			continue
		}
		var classes []string
		for _, sel := range strings.Split(match[1], ",") {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			if m := classSelectorRe.FindStringSubmatch(sel); m != nil {
				classes = append(classes, m[1])
			} else {
				unsupportedSelectors = append(unsupportedSelectors, sel)
			}
		}
		if len(classes) > 0 {
			rules = append(rules, cssRule{classes: classes, decls: decls})
		}
	}
	return rules, unsupportedSelectors
}

// cascadeFor resolves the class-rule declarations that apply to an element.
// Class selectors share one specificity, so the later rule in source order wins
// regardless of the order the classes are listed in the attribute.
func cascadeFor(classAttr string, rules []cssRule) map[string]classDecl {
	classes := strings.Fields(classAttr)
	if len(classes) == 0 {
		return nil
	}
	out := map[string]classDecl{}
	for _, rule := range rules {
		matched := matchingClass(rule, classes)
		if matched == "" {
			continue
		}
		for _, d := range rule.decls {
			out[d.prop] = classDecl{value: d.value, class: matched}
		}
	}
	return out
}

func matchingClass(rule cssRule, classes []string) string {
	for _, rc := range rule.classes {
		for _, c := range classes {
			if rc == c {
				return c
			}
		}
	}
	return ""
}

// parseDeclarations splits a declaration block or inline style attribute into
// property/value pairs, lowercasing property names. A trailing !important is
// dropped: no source we ingest relies on it.
func parseDeclarations(block string) []cssDecl {
	var decls []cssDecl
	for _, part := range strings.Split(block, ";") {
		prop, value, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		prop = strings.ToLower(strings.TrimSpace(prop))
		value = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(value), "!important"))
		if prop == "" || value == "" {
			continue
		}
		decls = append(decls, cssDecl{prop: prop, value: value})
	}
	return decls
}

func formatDeclarations(decls []cssDecl) string {
	parts := make([]string, len(decls))
	for i, d := range decls {
		parts[i] = d.prop + ":" + d.value
	}
	return strings.Join(parts, ";")
}

// inlineValue returns a property's value from an element's inline style. The last
// declaration wins, as in CSS.
func inlineValue(styleAttr, prop string) (string, bool) {
	value, found := "", false
	for _, d := range parseDeclarations(styleAttr) {
		if d.prop == prop {
			value, found = d.value, true
		}
	}
	return value, found
}
