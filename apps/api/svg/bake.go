package svg

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// cutListGlassGroup records the chosen group-level glass color for a manifest
// glass group (nil when left at the manifest default with no override).
type cutListGlassGroup struct {
	GroupKey     string `json:"group_key"`
	GlassColorID *int   `json:"glass_color_id,omitempty"`
	Count        int    `json:"count"`
}

type pieceCut struct {
	PieceID      string `json:"piece_id"`
	GlassColorID int    `json:"glass_color_id"`
}

type cutList struct {
	GlassGroups []cutListGlassGroup `json:"glass_groups"`
	Pieces      []pieceCut          `json:"pieces,omitempty"`
	GroutID     *int                `json:"grout_id,omitempty"`
}

// Bake produces a flat, fit, self-contained SVG from a structure SVG + manifest +
// content bbox + target dimensions + overrides. The artwork is fit and centered
// into a (width*300) x (height*300) viewBox. Colors resolve piece override ->
// group override -> manifest group default.
//
// The result stays re-editable: every piece keeps its id="pN" and group class,
// all <style> blocks are stripped, and only our ids/classes, the grout rect, the
// gac-fit wrapper, and the cutlist metadata remain.
func Bake(
	structureSVG []byte,
	manifest Manifest,
	bbox ContentBBox,
	width, height float64,
	overrides ColorOverrides,
	glassHexByID map[int]string,
	groutHexByID map[int]string,
) ([]byte, error) {
	doc, root, err := parseRoot(structureSVG)
	if err != nil {
		return nil, err
	}

	cl, err := recolor(root, manifest, overrides, glassHexByID, groutHexByID)
	if err != nil {
		return nil, err
	}

	stripStyles(root)
	applyFit(root, bbox, width, height)

	// Size the grout rect to the fit viewBox so it covers the whole canvas.
	if err := insertGrout(root, manifest, overrides, groutHexByID); err != nil {
		return nil, err
	}

	addCutListMetadata(root, cl)
	return doc.WriteToBytes()
}

// BakeConsumer renders a flat SVG for the consumer customizer. The stored
// structure SVG is already fit, so this path keeps the manifest viewBox and only
// applies scale_factor to the root width/height for display sizing — it never
// recomputes fit.
func BakeConsumer(
	structureSVG []byte,
	manifest Manifest,
	scaleFactor float64,
	overrides ColorOverrides,
	glassHexByID map[int]string,
	groutHexByID map[int]string,
) ([]byte, error) {
	doc, root, err := parseRoot(structureSVG)
	if err != nil {
		return nil, err
	}

	cl, err := recolor(root, manifest, overrides, glassHexByID, groutHexByID)
	if err != nil {
		return nil, err
	}

	if err := insertGrout(root, manifest, overrides, groutHexByID); err != nil {
		return nil, err
	}
	applyScale(root, manifest.ViewBox, scaleFactor)
	addCutListMetadata(root, cl)
	return doc.WriteToBytes()
}

func parseRoot(in []byte) (*etree.Document, *etree.Element, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(in); err != nil {
		return nil, nil, fmt.Errorf("parse structure svg: %w", err)
	}
	root := doc.SelectElement("svg")
	if root == nil {
		root = doc.Root()
	}
	if root == nil {
		return nil, nil, fmt.Errorf("structure svg has no root element")
	}
	return doc, root, nil
}

// recolor applies glass colors to pieces and returns the cutlist. Resolution
// order per piece is piece override -> group override -> manifest group default.
func recolor(
	root *etree.Element,
	manifest Manifest,
	overrides ColorOverrides,
	glassHexByID map[int]string,
	groutHexByID map[int]string,
) (cutList, error) {
	byID := indexByID(root)
	cl := cutList{}

	// Resolve the grout id (override wins over manifest default) for the cutlist.
	groutID := manifest.GroutRegion.GroutID
	if overrides.Background != nil {
		id := overrides.Background.GroutID
		groutID = &id
	}
	cl.GroutID = groutID

	// Stable iteration over glass groups for deterministic output.
	groupKeys := make([]string, 0, len(manifest.GlassRegions))
	for key := range manifest.GlassRegions {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	for _, key := range groupKeys {
		region := manifest.GlassRegions[key]

		// Group default: manifest value, overridden by an explicit group override.
		groupGlassID := region.GlassColorID
		if go_, ok := overrides.Groups[key]; ok {
			id := go_.GlassColorID
			groupGlassID = &id
		}
		cl.GlassGroups = append(cl.GlassGroups, cutListGlassGroup{
			GroupKey:     key,
			GlassColorID: groupGlassID,
			Count:        region.Count,
		})

		for _, pieceID := range region.PieceIDs {
			var glassID *int
			isPieceOverride := false
			if po, ok := overrides.Pieces[pieceID]; ok {
				id := po.GlassColorID
				glassID = &id
				isPieceOverride = true
			} else if groupGlassID != nil {
				glassID = groupGlassID
			}
			if glassID == nil {
				continue // no color assigned — keep the original source color
			}
			hex, ok := glassHexByID[*glassID]
			if !ok {
				return cutList{}, fmt.Errorf("unknown glass_color_id %d", *glassID)
			}
			el := byID[pieceID]
			if el == nil {
				continue
			}
			setInlineFill(el, hex)
			el.CreateAttr("data-glass-color-id", strconv.Itoa(*glassID))
			if isPieceOverride {
				cl.Pieces = append(cl.Pieces, pieceCut{PieceID: pieceID, GlassColorID: *glassID})
			}
		}
	}

	return cl, nil
}

// insertGrout colors the grout pieces when a grout id is resolved
// (override wins over manifest default).
func insertGrout(
	root *etree.Element,
	manifest Manifest,
	overrides ColorOverrides,
	groutHexByID map[int]string,
) error {
	groutID := manifest.GroutRegion.GroutID
	if overrides.Background != nil {
		id := overrides.Background.GroutID
		groutID = &id
	}
	if groutID == nil {
		return nil
	}

	hex, ok := groutHexByID[*groutID]
	if !ok {
		return fmt.Errorf("unknown grout_id %d", *groutID)
	}

	byID := indexByID(root)
	for _, pieceID := range groutPieceIDs(manifest) {
		if el := byID[pieceID]; el != nil {
			setInlineFill(el, hex)
			el.CreateAttr("data-grout-id", strconv.Itoa(*groutID))
		}
	}

	return nil
}

// groutPieceIDs returns all piece IDs in the manifest's single grout region.
func groutPieceIDs(manifest Manifest) []string {
	return manifest.GroutRegion.PieceIDs
}

// stripStyles removes all <style> elements so the baked SVG's color comes only
// from inline fills, keeping it self-contained.
func stripStyles(root *etree.Element) {
	doc := root
	for _, style := range doc.FindElements("//style") {
		if parent := style.Parent(); parent != nil {
			parent.RemoveChild(style)
		}
	}
	// Remove now-empty <defs> wrappers left behind.
	for _, defs := range doc.FindElements("//defs") {
		if len(defs.ChildElements()) == 0 {
			if parent := defs.Parent(); parent != nil {
				parent.RemoveChild(defs)
			}
		}
	}
}

func indexByID(root *etree.Element) map[string]*etree.Element {
	byID := map[string]*etree.Element{}
	var walk func(*etree.Element)
	walk = func(el *etree.Element) {
		for _, child := range el.ChildElements() {
			if id := child.SelectAttrValue("id", ""); id != "" {
				byID[id] = child
			}
			walk(child)
		}
	}
	walk(root)
	return byID
}

// setInlineFill replaces any existing inline fill declaration and sets the new
// one, preserving other inline style declarations.
func setInlineFill(el *etree.Element, hex string) {
	style := el.SelectAttrValue("style", "")
	style = inlineFillRe.ReplaceAllString(style, "")
	style = strings.Trim(strings.TrimSpace(style), ";")
	if style != "" {
		style += ";"
	}
	style += "fill:" + hex
	el.CreateAttr("style", style)
}

func applyScale(root *etree.Element, viewBox string, scaleFactor float64) {
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	_, _, w, h, ok := parseViewBox(viewBox)
	if !ok {
		return
	}
	root.CreateAttr("width", formatNum(w*scaleFactor))
	root.CreateAttr("height", formatNum(h*scaleFactor))
}

func addCutListMetadata(root *etree.Element, cl cutList) {
	data, err := json.Marshal(cl)
	if err != nil {
		return
	}
	meta := etree.NewElement("metadata")
	meta.CreateAttr("id", "glassact-cutlist")
	meta.SetText(string(data))
	root.InsertChildAt(0, meta)
}

func parseViewBox(viewBox string) (x, y, w, h float64, ok bool) {
	fields := strings.Fields(strings.ReplaceAll(viewBox, ",", " "))
	if len(fields) != 4 {
		return 0, 0, 0, 0, false
	}
	x = parseDim(fields[0])
	y = parseDim(fields[1])
	w = parseDim(fields[2])
	h = parseDim(fields[3])
	if w <= 0 || h <= 0 {
		return 0, 0, 0, 0, false
	}
	return x, y, w, h, true
}
