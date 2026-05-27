package svg

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// cutListEntry records, per source region, the chosen group-level glass color
// (nil when the region was left at its original color). Individual piece
// overrides are listed separately so production can see divergence.
type cutListEntry struct {
	SourceHex    string `json:"source_hex"`
	GlassColorID *int   `json:"glass_color_id,omitempty"`
	Count        int    `json:"count"`
}

type pieceCut struct {
	PieceID      string `json:"piece_id"`
	GlassColorID int    `json:"glass_color_id"`
}

type cutList struct {
	Regions []cutListEntry `json:"regions"`
	Pieces  []pieceCut     `json:"pieces,omitempty"`
	GroutID *int           `json:"grout_id,omitempty"`
}

// Bake produces a flat, self-contained SVG from a canonical SVG + manifest +
// overrides. Colors are resolved piece override -> region mapping -> source.
//
// Recolored shapes get an inline `style="fill:#hex"`. Inline style is used
// (not a `fill` attribute) deliberately: a CSS rule from the `<style>` block
// outranks a presentation attribute, but an inline style outranks the class
// rule — so this is what actually overrides the original color.
func Bake(
	canonical []byte,
	manifest Manifest,
	scaleFactor float64,
	overrides ColorOverrides,
	glassHexByID map[int]string,
	groutHexByID map[int]string,
) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(canonical); err != nil {
		return nil, fmt.Errorf("parse canonical svg: %w", err)
	}
	root := doc.SelectElement("svg")
	if root == nil {
		root = doc.Root()
	}
	if root == nil {
		return nil, fmt.Errorf("canonical svg has no root element")
	}

	byID := indexByID(root)

	cl := cutList{}
	if overrides.Background != nil {
		id := overrides.Background.GroutID
		cl.GroutID = &id
	}

	// Stable iteration for deterministic output.
	sourceHexes := make([]string, 0, len(manifest.Regions))
	for hex := range manifest.Regions {
		sourceHexes = append(sourceHexes, hex)
	}
	sort.Strings(sourceHexes)

	for _, sourceHex := range sourceHexes {
		region := manifest.Regions[sourceHex]

		var regionGlassID *int
		if ro, ok := overrides.Regions[sourceHex]; ok {
			id := ro.GlassColorID
			regionGlassID = &id
		}
		cl.Regions = append(cl.Regions, cutListEntry{
			SourceHex:    sourceHex,
			GlassColorID: regionGlassID,
			Count:        region.Count,
		})

		for _, pieceID := range region.PieceIDs {
			var glassID *int
			isPieceOverride := false
			if po, ok := overrides.Pieces[pieceID]; ok {
				id := po.GlassColorID
				glassID = &id
				isPieceOverride = true
			} else if regionGlassID != nil {
				glassID = regionGlassID
			}
			if glassID == nil {
				continue // unchanged — keep the original source color
			}
			hex, ok := glassHexByID[*glassID]
			if !ok {
				return nil, fmt.Errorf("unknown glass_color_id %d", *glassID)
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

	if overrides.Background != nil {
		hex, ok := groutHexByID[overrides.Background.GroutID]
		if !ok {
			return nil, fmt.Errorf("unknown grout_id %d", overrides.Background.GroutID)
		}
		// Color all grout shapes (back-most region + implicit-black pieces) with the grout hex.
		for _, pieceID := range groutPieceIDs(manifest) {
			if el := byID[pieceID]; el != nil {
				setInlineFill(el, hex)
			}
		}
		if rect := groutRect(manifest.ViewBox, hex, overrides.Background.GroutID); rect != nil {
			root.InsertChildAt(0, rect)
		}
	}

	applyScale(root, manifest.ViewBox, scaleFactor)
	addCutListMetadata(root, cl)

	return doc.WriteToBytes()
}

// groutPieceIDs returns all piece IDs belonging to grout regions:
//  1. The region containing "p0" (first/back-most shape in document order).
//  2. The defaultFill ("#000000") region (classless implicit-black shapes like eyes).
func groutPieceIDs(manifest Manifest) []string {
	groutHexes := map[string]bool{defaultFill: true}
	for hex, region := range manifest.Regions {
		for _, id := range region.PieceIDs {
			if id == "p0" {
				groutHexes[hex] = true
				break
			}
		}
	}
	var ids []string
	for hex, region := range manifest.Regions {
		if groutHexes[hex] {
			ids = append(ids, region.PieceIDs...)
		}
	}
	return ids
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

func groutRect(viewBox, hex string, groutID int) *etree.Element {
	x, y, w, h, ok := parseViewBox(viewBox)
	if !ok {
		return nil
	}
	rect := etree.NewElement("rect")
	rect.CreateAttr("id", "glassact-grout")
	rect.CreateAttr("data-grout-id", strconv.Itoa(groutID))
	rect.CreateAttr("x", formatNum(x))
	rect.CreateAttr("y", formatNum(y))
	rect.CreateAttr("width", formatNum(w))
	rect.CreateAttr("height", formatNum(h))
	rect.CreateAttr("style", "fill:"+hex)
	return rect
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
