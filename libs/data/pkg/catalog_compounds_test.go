package data

import (
	"errors"
	"testing"
)

func createCompoundMemberItem(t *testing.T, models Models, priceGroupID int, code, category string) *CatalogItem {
	t.Helper()
	item := &CatalogItem{
		CatalogCode:         code,
		Name:                "Item " + code,
		Category:            category,
		DefaultWidth:        3,
		DefaultHeight:       4,
		MinWidth:            3,
		MinHeight:           4,
		DefaultPriceGroupID: priceGroupID,
		SvgURL:              "/file/catalog-items/" + code + ".svg",
		IsActive:            true,
	}
	if err := models.CatalogItems.Insert(item); err != nil {
		t.Fatalf("failed to create catalog item %s: %v", code, err)
	}
	return item
}

func TestCatalogCompound_InsertLoadsMembersInOrderWithItems(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)
	left := createCompoundMemberItem(t, models, priceGroup.ID, "A-DOV-0001L", "A-ANIMALS")
	right := createCompoundMemberItem(t, models, priceGroup.ID, "A-DOV-0001R", "A-ANIMALS")

	compound := &CatalogCompound{
		Name:     "Dove Pair",
		IsActive: true,
		Members: []CatalogCompoundMember{
			{CatalogItemID: left.ID, CenterX: -2, CenterY: 0},
			{CatalogItemID: right.ID, CenterX: 2, CenterY: 0, RotationDeg: 15},
		},
	}
	if err := models.CatalogCompounds.Insert(compound); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	loaded, found, err := models.CatalogCompounds.GetByUUID(compound.UUID)
	if err != nil || !found {
		t.Fatalf("expected compound, found=%v err=%v", found, err)
	}
	if len(loaded.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(loaded.Members))
	}
	if loaded.Members[0].CatalogItem.CatalogCode != "A-DOV-0001L" || loaded.Members[1].RotationDeg != 15 {
		t.Errorf("members out of order or missing data: %+v", loaded.Members)
	}
	if !loaded.IsOfferable() {
		t.Error("expected compound of active items to be offerable")
	}
}

func TestCatalogCompound_WithRetiredMember_IsNotOfferable(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)
	kept := createCompoundMemberItem(t, models, priceGroup.ID, "A-DOV-0002", "A-ANIMALS")
	retired := createCompoundMemberItem(t, models, priceGroup.ID, "A-DOV-0003", "A-ANIMALS")

	compound := &CatalogCompound{
		Name:     "Mixed",
		IsActive: true,
		Members: []CatalogCompoundMember{
			{CatalogItemID: kept.ID},
			{CatalogItemID: retired.ID, CenterX: 3},
		},
	}
	if err := models.CatalogCompounds.Insert(compound); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	retired.IsActive = false
	if err := models.CatalogItems.Update(retired); err != nil {
		t.Fatalf("failed to retire item: %v", err)
	}

	loaded, _, err := models.CatalogCompounds.GetByUUID(compound.UUID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.IsOfferable() {
		t.Error("expected compound with a retired member to be unofferable")
	}
}

func TestCatalogCompound_UpdateReplacesMembersAndRejectsStaleVersion(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)
	a := createCompoundMemberItem(t, models, priceGroup.ID, "B-TRE-0001", "B-OUTDOORS")
	b := createCompoundMemberItem(t, models, priceGroup.ID, "B-TRE-0002", "B-OUTDOORS")
	c := createCompoundMemberItem(t, models, priceGroup.ID, "B-TRE-0003", "B-OUTDOORS")

	compound := &CatalogCompound{
		Name:     "Trees",
		IsActive: true,
		Members:  []CatalogCompoundMember{{CatalogItemID: a.ID}, {CatalogItemID: b.ID, CenterX: 4}},
	}
	if err := models.CatalogCompounds.Insert(compound); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	stale := *compound

	compound.Members = []CatalogCompoundMember{{CatalogItemID: c.ID}, {CatalogItemID: a.ID, CenterY: 2}, {CatalogItemID: b.ID}}
	if err := models.CatalogCompounds.Update(compound); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	loaded, _, err := models.CatalogCompounds.GetByUUID(compound.UUID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Members) != 3 || loaded.Members[0].CatalogItemID != c.ID {
		t.Fatalf("members not replaced: %+v", loaded.Members)
	}

	stale.Name = "Stale"
	if err := models.CatalogCompounds.Update(&stale); !errors.Is(err, ErrEditConflict) {
		t.Fatalf("expected ErrEditConflict, got %v", err)
	}
}
