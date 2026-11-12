package data

import (
	"errors"
	"slices"
	"testing"
)

func relatedIDs(background *SandblastBackground) []int {
	ids := make([]int, len(background.RelatedItems))
	for i, item := range background.RelatedItems {
		ids[i] = item.ID
	}
	return ids
}

func TestSandblastBackground_InsertLoadsTagsAndRelatedItemsInOrder(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)
	small := createCompoundMemberItem(t, models, priceGroup.ID, "E-HRT-0002", "E-ORNAMENTS")
	large := createCompoundMemberItem(t, models, priceGroup.ID, "E-HRT-0001", "E-ORNAMENTS")

	background := &SandblastBackground{
		Code:         "SB-HRT-0001",
		Name:         "Heart Frame",
		SvgURL:       "/file/sandblast-backgrounds/heart.svg",
		Tags:         []string{"heart", "frame"},
		IsActive:     true,
		RelatedItems: []RelatedCatalogItem{{ID: small.ID}, {ID: large.ID}},
	}
	if err := models.SandblastBackgrounds.Insert(background); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	loaded, found, err := models.SandblastBackgrounds.GetByUUID(background.UUID)
	if err != nil || !found {
		t.Fatalf("expected background, found=%v err=%v", found, err)
	}
	if !slices.Equal(loaded.Tags, []string{"heart", "frame"}) {
		t.Errorf("expected tags to round-trip, got %v", loaded.Tags)
	}
	if !slices.Equal(relatedIDs(loaded), []int{small.ID, large.ID}) {
		t.Errorf("expected related items in the order given, got %v", relatedIDs(loaded))
	}
	if loaded.RelatedItems[0].CatalogCode != "E-HRT-0002" || loaded.RelatedItems[0].DefaultWidth != 3 {
		t.Errorf("related item is missing catalog data: %+v", loaded.RelatedItems[0])
	}
}

func TestSandblastBackground_UpdateReplacesRelatedItemsAndRejectsStaleVersion(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)
	a := createCompoundMemberItem(t, models, priceGroup.ID, "D-ROS-0001", "D-FLOWERS")
	b := createCompoundMemberItem(t, models, priceGroup.ID, "D-ROS-0002", "D-FLOWERS")
	c := createCompoundMemberItem(t, models, priceGroup.ID, "D-ROS-0003", "D-FLOWERS")

	background := &SandblastBackground{
		Code:         "SB-FL-0001",
		Name:         "Rose Border",
		SvgURL:       "/file/sandblast-backgrounds/rose.svg",
		IsActive:     true,
		RelatedItems: []RelatedCatalogItem{{ID: a.ID}, {ID: b.ID}},
	}
	if err := models.SandblastBackgrounds.Insert(background); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	stale := *background

	background.Name = "Rose Border Wide"
	background.RelatedItems = []RelatedCatalogItem{{ID: c.ID}, {ID: a.ID}}
	if err := models.SandblastBackgrounds.Update(background); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	loaded, _, err := models.SandblastBackgrounds.GetByUUID(background.UUID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Name != "Rose Border Wide" || !slices.Equal(relatedIDs(loaded), []int{c.ID, a.ID}) {
		t.Fatalf("background not updated: name=%q related=%v", loaded.Name, relatedIDs(loaded))
	}

	stale.Name = "Stale"
	if err := models.SandblastBackgrounds.Update(&stale); !errors.Is(err, ErrEditConflict) {
		t.Fatalf("expected ErrEditConflict, got %v", err)
	}
}

func TestSandblastBackground_UpdateFieldsKeepsRelatedItems(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)
	item := createCompoundMemberItem(t, models, priceGroup.ID, "F-CRS-0001", "F-FAITH")

	background := &SandblastBackground{
		Code:         "SB-CRS-0001",
		Name:         "Cross Rays",
		SvgURL:       "/file/sandblast-backgrounds/cross.svg",
		IsActive:     true,
		RelatedItems: []RelatedCatalogItem{{ID: item.ID}},
	}
	if err := models.SandblastBackgrounds.Insert(background); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	background.IsActive = false
	background.RelatedItems = nil
	if err := models.SandblastBackgrounds.UpdateFields(background); err != nil {
		t.Fatalf("update fields failed: %v", err)
	}

	loaded, _, err := models.SandblastBackgrounds.GetByUUID(background.UUID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.IsActive || len(loaded.RelatedItems) != 1 {
		t.Errorf("expected inactive background with its related item kept, got active=%v related=%v", loaded.IsActive, relatedIDs(loaded))
	}
}

func TestSandblastBackground_GetAllActiveHidesRetiredBackgrounds(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	for _, b := range []*SandblastBackground{
		{Code: "SB-BOR-0002", Name: "Shown", SvgURL: "/a.svg", IsActive: true},
		{Code: "SB-BOR-0001", Name: "Retired", SvgURL: "/b.svg", IsActive: false},
	} {
		if err := models.SandblastBackgrounds.Insert(b); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	active, err := models.SandblastBackgrounds.GetAllActive()
	if err != nil {
		t.Fatalf("get active failed: %v", err)
	}
	if len(active) != 1 || active[0].Code != "SB-BOR-0002" {
		t.Errorf("expected only the active background, got %+v", active)
	}

	all, err := models.SandblastBackgrounds.GetAll()
	if err != nil {
		t.Fatalf("get all failed: %v", err)
	}
	if len(all) != 2 || all[0].Code != "SB-BOR-0001" {
		t.Errorf("expected both backgrounds ordered by code, got %+v", all)
	}
}

func TestSandblastBackground_WithoutRetiredItemsDropsDisabledCatalogItems(t *testing.T) {
	background := &SandblastBackground{
		RelatedItems: []RelatedCatalogItem{{ID: 1, IsActive: true}, {ID: 2, IsActive: false}, {ID: 3, IsActive: true}},
	}

	offered := background.WithoutRetiredItems()

	if !slices.Equal(relatedIDs(offered), []int{1, 3}) {
		t.Errorf("expected only active items, got %v", relatedIDs(offered))
	}
	if len(background.RelatedItems) != 3 {
		t.Error("expected the original background to be left untouched")
	}
}
