package data

import (
	"testing"
)

func TestGlassColor_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	gc := &GlassColor{
		Name:      "Cobalt Blue",
		Hex:       "#2b3278",
		Family:    stringPtr("blue"),
		SortOrder: 10,
		IsActive:  true,
	}

	if err := models.GlassColors.Insert(gc); err != nil {
		t.Fatalf("Failed to insert glass color: %v", err)
	}
	if gc.ID == 0 {
		t.Errorf("Expected non-zero ID")
	}
	if gc.UUID == "" {
		t.Errorf("Expected UUID to be set")
	}
}

func TestGlassColor_GetByIDAndUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &GlassColor{Name: "Red", Hex: "#910028", SortOrder: 20, IsActive: true}
	if err := models.GlassColors.Insert(original); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	byID, found, err := models.GlassColors.GetByID(original.ID)
	if err != nil || !found {
		t.Fatalf("GetByID failed: found=%v err=%v", found, err)
	}
	if byID.Hex != "#910028" {
		t.Errorf("Expected hex #910028, got %s", byID.Hex)
	}

	byUUID, found, err := models.GlassColors.GetByUUID(original.UUID)
	if err != nil || !found {
		t.Fatalf("GetByUUID failed: found=%v err=%v", found, err)
	}
	if byUUID.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, byUUID.ID)
	}
}

func TestGlassColor_GetAllActive_SortedAndFiltered(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	mustInsert := func(gc *GlassColor) {
		if err := models.GlassColors.Insert(gc); err != nil {
			t.Fatalf("Failed to insert %s: %v", gc.Name, err)
		}
	}
	mustInsert(&GlassColor{Name: "Second", Hex: "#222222", SortOrder: 20, IsActive: true})
	mustInsert(&GlassColor{Name: "First", Hex: "#111111", SortOrder: 10, IsActive: true})
	mustInsert(&GlassColor{Name: "Inactive", Hex: "#333333", SortOrder: 5, IsActive: false})

	active, err := models.GlassColors.GetAllActive()
	if err != nil {
		t.Fatalf("GetAllActive failed: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("Expected 2 active glass colors, got %d", len(active))
	}
	if active[0].Name != "First" || active[1].Name != "Second" {
		t.Errorf("Expected sort_order ordering [First, Second], got [%s, %s]", active[0].Name, active[1].Name)
	}
}

func TestGlassColor_VersionIncrement(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	gc := &GlassColor{Name: "Amber", Hex: "#eeb211", SortOrder: 1, IsActive: true}
	if err := models.GlassColors.Insert(gc); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	initial := gc.Version
	gc.Name = "Sunflower"
	if err := models.GlassColors.Update(gc); err != nil {
		t.Fatalf("Failed to update: %v", err)
	}
	if gc.Version <= initial {
		t.Errorf("Expected version to increment from %d, got %d", initial, gc.Version)
	}
}
