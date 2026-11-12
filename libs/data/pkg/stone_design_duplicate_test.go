package data

import "testing"

const (
	sourceInlayUUID = "7ddea090-6db0-4156-95cd-831bcbbf5cc0"
	copyInlayUUID   = "0b6f3a63-1f2e-4f8e-9b0e-2b8c3f6c9a11"
)

func TestStoneDesign_TxPlaceDuplicate_PlacesCopyNextToSource(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	design := NewDefaultStoneDesign(project.ID)
	design.Layout = validLayout()
	if err := models.StoneDesigns.Insert(design); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	tx, err := models.STDB.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	placed, err := models.StoneDesigns.TxPlaceDuplicate(tx, project.ID, sourceInlayUUID, copyInlayUUID)
	if err != nil {
		tx.Rollback()
		t.Fatalf("place failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if !placed {
		t.Fatal("expected the copy to be placed")
	}

	loaded, _, err := models.StoneDesigns.GetByProjectID(project.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	var copyElement *StoneElement
	for i := range loaded.Layout.Elements {
		if loaded.Layout.Elements[i].InlayUUID == copyInlayUUID {
			copyElement = &loaded.Layout.Elements[i]
		}
	}
	if copyElement == nil {
		t.Fatalf("copy not on the stone: %+v", loaded.Layout.Elements)
	}
	if copyElement.X != 10+DuplicateOffsetInches || copyElement.Y != 8+DuplicateOffsetInches {
		t.Errorf("copy at (%v, %v), expected offset from source at (10, 8)", copyElement.X, copyElement.Y)
	}
	if copyElement.GroupID != nil {
		t.Errorf("copy should be ungrouped, got group %q", *copyElement.GroupID)
	}
}

func TestStoneDesign_TxPlaceDuplicate_SourceNotOnStone_ChangesNothing(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	design := NewDefaultStoneDesign(project.ID)
	if err := models.StoneDesigns.Insert(design); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	tx, err := models.STDB.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	defer tx.Rollback()

	placed, err := models.StoneDesigns.TxPlaceDuplicate(tx, project.ID, sourceInlayUUID, copyInlayUUID)
	if err != nil {
		t.Fatalf("place failed: %v", err)
	}
	if placed {
		t.Error("expected nothing placed when the source is not on the stone")
	}
}

func TestStoneDesign_TxPlaceDuplicate_NoDesign_ChangesNothing(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	tx, err := models.STDB.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	defer tx.Rollback()

	placed, err := models.StoneDesigns.TxPlaceDuplicate(tx, project.ID, sourceInlayUUID, copyInlayUUID)
	if err != nil {
		t.Fatalf("place failed: %v", err)
	}
	if placed {
		t.Error("expected nothing placed when the project has no design")
	}
}
