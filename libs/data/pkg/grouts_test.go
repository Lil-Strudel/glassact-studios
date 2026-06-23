package data

import (
	"testing"
)

func TestGrout_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	g := &Grout{Name: "Black Granite", Hex: "#1c1c1c", SortOrder: 10, IsActive: true}
	if err := models.Grouts.Insert(g); err != nil {
		t.Fatalf("Failed to insert grout: %v", err)
	}
	if g.ID == 0 || g.UUID == "" {
		t.Errorf("Expected ID and UUID to be set, got ID=%d UUID=%q", g.ID, g.UUID)
	}
}

func TestGrout_GetAllActive_SortedAndFiltered(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	mustInsert := func(g *Grout) {
		if err := models.Grouts.Insert(g); err != nil {
			t.Fatalf("Failed to insert %s: %v", g.Name, err)
		}
	}
	mustInsert(&Grout{Name: "Grey", Hex: "#8a8d8f", SortOrder: 20, IsActive: true})
	mustInsert(&Grout{Name: "Charcoal", Hex: "#3b3e40", SortOrder: 10, IsActive: true})
	mustInsert(&Grout{Name: "Retired", Hex: "#999999", SortOrder: 1, IsActive: false})

	active, err := models.Grouts.GetAllActive()
	if err != nil {
		t.Fatalf("GetAllActive failed: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("Expected 2 active grouts, got %d", len(active))
	}
	if active[0].Name != "Charcoal" || active[1].Name != "Grey" {
		t.Errorf("Expected ordering [Charcoal, Grey], got [%s, %s]", active[0].Name, active[1].Name)
	}
}

func TestGrout_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &Grout{Name: "Mahogany Granite", Hex: "#4a1f1a", SortOrder: 5, IsActive: true}
	if err := models.Grouts.Insert(original); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.Grouts.GetByUUID(original.UUID)
	if err != nil || !found {
		t.Fatalf("GetByUUID failed: found=%v err=%v", found, err)
	}
	if retrieved.Hex != "#4a1f1a" {
		t.Errorf("Expected hex #4a1f1a, got %s", retrieved.Hex)
	}
}
