package data

import "testing"

func TestInlay_MirrorFlags_RoundTripThroughInsertAndUpdate(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:   project.ID,
		Name:        "Bunny",
		Type:        InlayTypes.Catalog,
		MirrorX:     true,
		CatalogInfo: &InlayCatalogInfo{CatalogItemID: catalogItem.ID},
	}
	if err := models.Inlays.Insert(inlay); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	inserted, _, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !inserted.MirrorX || inserted.MirrorY {
		t.Fatalf("expected mirror_x only after insert, got x=%v y=%v", inserted.MirrorX, inserted.MirrorY)
	}

	inserted.MirrorX = false
	inserted.MirrorY = true
	if err := models.Inlays.Update(inserted); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, _, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if updated.MirrorX || !updated.MirrorY {
		t.Errorf("expected mirror_y only after update, got x=%v y=%v", updated.MirrorX, updated.MirrorY)
	}
}
