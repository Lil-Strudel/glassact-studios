package data

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func validLayout() StoneLayout {
	return StoneLayout{
		Groups: []StoneGroup{{ID: "g1", Name: "Doves"}},
		Elements: []StoneElement{
			{ID: "e1", Type: StoneElementTypes.Inlay, X: 10, Y: 8, InlayUUID: "7ddea090-6db0-4156-95cd-831bcbbf5cc0", GroupID: strPtr("g1")},
			{ID: "e2", Type: StoneElementTypes.Background, X: 18, Y: 12, Width: 30, BackgroundUUID: "22e3d0fe-a21f-410d-8748-e5baf62e9e37"},
			{ID: "e3", Type: StoneElementTypes.Text, X: 18, Y: 4, Text: "SMITH", Font: "roman", Size: 2, Align: "center", Finish: "engraved"},
		},
	}
}

func TestStoneLayoutValidate_AcceptsEachElementType(t *testing.T) {
	if err := validLayout().Validate(); err != nil {
		t.Fatalf("expected valid layout, got %v", err)
	}
}

func TestStoneLayoutValidate_AcceptsEachTextFinish(t *testing.T) {
	for _, finish := range []string{"engraved", "frosted", "black-fill"} {
		layout := validLayout()
		layout.Elements[2].Finish = finish
		if err := layout.Validate(); err != nil {
			t.Errorf("finish %q: expected valid layout, got %v", finish, err)
		}
	}
}

func TestStoneLayoutValidate_RejectsBadLayouts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*StoneLayout)
	}{
		{"unknown group", func(l *StoneLayout) { l.Elements[0].GroupID = strPtr("nope") }},
		{"duplicate element id", func(l *StoneLayout) { l.Elements[1].ID = "e1" }},
		{"inlay placed twice", func(l *StoneLayout) {
			dup := l.Elements[0]
			dup.ID = "e4"
			l.Elements = append(l.Elements, dup)
		}},
		{"NaN position", func(l *StoneLayout) { l.Elements[0].X = math.NaN() }},
		{"infinite rotation", func(l *StoneLayout) { l.Elements[0].Rotation = math.Inf(1) }},
		{"background without width", func(l *StoneLayout) { l.Elements[1].Width = 0 }},
		{"text too long", func(l *StoneLayout) { l.Elements[2].Text = strings.Repeat("a", MaxStoneTextLength+1) }},
		{"unknown font", func(l *StoneLayout) { l.Elements[2].Font = "comic" }},
		{"unknown finish", func(l *StoneLayout) { l.Elements[2].Finish = "gold-leaf" }},
		{"unknown type", func(l *StoneLayout) { l.Elements[2].Type = "sticker" }},
		{"bad inlay uuid", func(l *StoneLayout) { l.Elements[0].InlayUUID = "not-a-uuid" }},
		{"too many elements", func(l *StoneLayout) {
			l.Elements = make([]StoneElement, MaxStoneElements+1)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := validLayout()
			tt.mutate(&layout)
			if err := layout.Validate(); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestStoneDesign_InsertThenGetRoundTripsLayout(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	design := NewDefaultStoneDesign(project.ID)
	design.Layout = validLayout()
	if err := models.StoneDesigns.Insert(design); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	loaded, found, err := models.StoneDesigns.GetByProjectID(project.ID)
	if err != nil || !found {
		t.Fatalf("expected design, found=%v err=%v", found, err)
	}
	if len(loaded.Layout.Elements) != 3 || len(loaded.Layout.Groups) != 1 {
		t.Fatalf("layout did not round-trip: %+v", loaded.Layout)
	}
	if loaded.Layout.Elements[2].Text != "SMITH" || *loaded.Layout.Elements[0].GroupID != "g1" {
		t.Errorf("element fields lost: %+v", loaded.Layout.Elements)
	}
}

func TestStoneDesign_SecondInsertForProject_ReturnsEditConflict(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	if err := models.StoneDesigns.Insert(NewDefaultStoneDesign(project.ID)); err != nil {
		t.Fatalf("first insert failed: %v", err)
	}
	err := models.StoneDesigns.Insert(NewDefaultStoneDesign(project.ID))
	if !errors.Is(err, ErrEditConflict) {
		t.Fatalf("expected ErrEditConflict, got %v", err)
	}
}

func TestStoneDesign_UpdateWithStaleVersion_ReturnsEditConflict(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	design := NewDefaultStoneDesign(project.ID)
	if err := models.StoneDesigns.Insert(design); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	stale := *design
	design.StoneShape = StoneShapes.Heart
	if err := models.StoneDesigns.Update(design); err != nil {
		t.Fatalf("first update failed: %v", err)
	}
	if design.Version != 2 {
		t.Errorf("expected version 2, got %d", design.Version)
	}

	stale.StoneShape = StoneShapes.Cathedral
	if err := models.StoneDesigns.Update(&stale); !errors.Is(err, ErrEditConflict) {
		t.Fatalf("expected ErrEditConflict, got %v", err)
	}
}
