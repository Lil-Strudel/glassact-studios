package data

import (
	"fmt"
	"slices"
	"testing"
)

func TestCatalogItem_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	catalogItem := &CatalogItem{
		CatalogCode:         "CAT-001",
		Name:                "Test Item",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(catalogItem)
	if err != nil {
		t.Fatalf("Failed to insert catalog item: %v", err)
	}

	if catalogItem.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", catalogItem.ID)
	}
	if catalogItem.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if catalogItem.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestCatalogItem_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	// Insert test item
	original := &CatalogItem{
		CatalogCode:         "CAT-002",
		Name:                "Test Item 2",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Retrieve by ID
	retrieved, found, err := models.CatalogItems.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Item not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.CatalogCode != original.CatalogCode {
		t.Errorf("Expected code %s, got %s", original.CatalogCode, retrieved.CatalogCode)
	}
	if retrieved.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, retrieved.Name)
	}
}

func TestCatalogItem_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	original := &CatalogItem{
		CatalogCode:         "CAT-003",
		Name:                "Test Item 3",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.CatalogItems.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Item not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestCatalogItem_GetByCatalogCode(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	original := &CatalogItem{
		CatalogCode:         "UNIQUE-CODE",
		Name:                "Test Item",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.CatalogItems.GetByCatalogCode(original.CatalogCode)
	if err != nil {
		t.Fatalf("Failed to get by catalog code: %v", err)
	}
	if !found {
		t.Errorf("Item not found by catalog code")
	}
	if retrieved.CatalogCode != original.CatalogCode {
		t.Errorf("Expected code %s, got %s", original.CatalogCode, retrieved.CatalogCode)
	}
}

func TestCatalogItem_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	original := &CatalogItem{
		CatalogCode:         "CAT-004",
		Name:                "Original Name",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.Name = "Updated Name"
	original.IsActive = false

	err = models.CatalogItems.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.CatalogItems.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Item not found after update")
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.IsActive != false {
		t.Errorf("Expected IsActive to be false")
	}
}

func TestCatalogItem_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	item := &CatalogItem{
		CatalogCode:         "CAT-005",
		Name:                "Item to Delete",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(item)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.CatalogItems.Delete(item.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.CatalogItems.GetByID(item.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected item to be deleted")
	}
}

func TestCatalogItem_Tags(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	item := &CatalogItem{
		CatalogCode:         "CAT-006",
		Name:                "Item with Tags",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(item)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Add tags
	err = models.CatalogItems.AddTag(item.ID, "premium")
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	err = models.CatalogItems.AddTag(item.ID, "glass-cut")
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	// Get tags
	tags, err := models.CatalogItems.GetTags(item.ID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	// Remove tag
	err = models.CatalogItems.RemoveTag(item.ID, "premium")
	if err != nil {
		t.Fatalf("Failed to remove tag: %v", err)
	}

	tags, err = models.CatalogItems.GetTags(item.ID)
	if err != nil {
		t.Fatalf("Failed to get tags after removal: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag after removal, got %d", len(tags))
	}
}

func TestCatalogItem_GetByTag(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	item1 := &CatalogItem{
		CatalogCode:         "CAT-007",
		Name:                "Tagged Item 1",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	item2 := &CatalogItem{
		CatalogCode:         "CAT-008",
		Name:                "Tagged Item 2",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(item1)
	if err != nil {
		t.Fatalf("Failed to insert item1: %v", err)
	}
	err = models.CatalogItems.Insert(item2)
	if err != nil {
		t.Fatalf("Failed to insert item2: %v", err)
	}

	// Tag both items
	err = models.CatalogItems.AddTag(item1.ID, "premium")
	if err != nil {
		t.Fatalf("Failed to add tag to item1: %v", err)
	}
	err = models.CatalogItems.AddTag(item2.ID, "premium")
	if err != nil {
		t.Fatalf("Failed to add tag to item2: %v", err)
	}

	// Get by tag
	items, err := models.CatalogItems.GetByTag("premium")
	if err != nil {
		t.Fatalf("Failed to get by tag: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items with premium tag, got %d", len(items))
	}
}

func TestCatalogItem_GetByCategory(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	item1 := &CatalogItem{
		CatalogCode:         "CAT-009",
		Name:                "Glass Item",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	item2 := &CatalogItem{
		CatalogCode:         "CAT-010",
		Name:                "Metal Item",
		Category:            "Metal",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(item1)
	if err != nil {
		t.Fatalf("Failed to insert item1: %v", err)
	}
	err = models.CatalogItems.Insert(item2)
	if err != nil {
		t.Fatalf("Failed to insert item2: %v", err)
	}

	// Get by category
	glassItems, err := models.CatalogItems.GetByCategory("Glass")
	if err != nil {
		t.Fatalf("Failed to get by category: %v", err)
	}
	if len(glassItems) < 1 {
		t.Errorf("Expected at least 1 glass item, got %d", len(glassItems))
	}
}

func TestCatalogItem_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	item1 := &CatalogItem{
		CatalogCode:         "CAT-011",
		Name:                "All Test 1",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	item2 := &CatalogItem{
		CatalogCode:         "CAT-012",
		Name:                "All Test 2",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}

	err := models.CatalogItems.Insert(item1)
	if err != nil {
		t.Fatalf("Failed to insert item1: %v", err)
	}
	err = models.CatalogItems.Insert(item2)
	if err != nil {
		t.Fatalf("Failed to insert item2: %v", err)
	}

	items, err := models.CatalogItems.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(items) < 2 {
		t.Errorf("Expected at least 2 items, got %d", len(items))
	}
}

// newTestCatalogItem builds a valid catalog item so ordering/ranking tests can
// stay focused on the fields they actually exercise.
func newTestCatalogItem(code, name, category string, priceGroupID int) *CatalogItem {
	return &CatalogItem{
		CatalogCode:         code,
		Name:                name,
		Category:            category,
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: priceGroupID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}
}

func TestCatalogItem_GetCategories_ReturnsSortedCategories(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	// Inserted deliberately out of order.
	for i, category := range []string{"D-FLOWERS", "A-ANIMALS", "SB-BACKGROUNDS", "B-OUTDOORS"} {
		item := newTestCatalogItem(fmt.Sprintf("SORT-CAT-%d", i), "Item", category, priceGroup.ID)
		if err := models.CatalogItems.Insert(item); err != nil {
			t.Fatalf("Failed to insert item: %v", err)
		}
	}

	categories, err := models.CatalogItems.GetCategories()
	if err != nil {
		t.Fatalf("Failed to get categories: %v", err)
	}

	want := []string{"A-ANIMALS", "B-OUTDOORS", "D-FLOWERS", "SB-BACKGROUNDS"}
	if !slices.Equal(categories, want) {
		t.Errorf("Expected categories %v, got %v", want, categories)
	}
}

func TestCatalogItem_GetAllActive_OrdersByDisplayOrderThenName(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	// Names chosen so alphabetical order differs from insertion order.
	zebra := newTestCatalogItem("ORD-1", "Zebra", "A-ANIMALS", priceGroup.ID)
	apple := newTestCatalogItem("ORD-2", "Apple", "A-ANIMALS", priceGroup.ID)
	mango := newTestCatalogItem("ORD-3", "Mango", "A-ANIMALS", priceGroup.ID)

	for _, item := range []*CatalogItem{zebra, apple, mango} {
		if err := models.CatalogItems.Insert(item); err != nil {
			t.Fatalf("Failed to insert %s: %v", item.Name, err)
		}
	}

	// Rank Zebra first; Apple and Mango stay unranked and should follow by name.
	if err := models.CatalogItems.SetDisplayOrder([]string{zebra.UUID}); err != nil {
		t.Fatalf("Failed to set display order: %v", err)
	}

	items, err := models.CatalogItems.GetAllActive()
	if err != nil {
		t.Fatalf("Failed to get active items: %v", err)
	}

	got := make([]string, len(items))
	for i, item := range items {
		got[i] = item.Name
	}

	want := []string{"Zebra", "Apple", "Mango"}
	if !slices.Equal(got, want) {
		t.Errorf("Expected order %v, got %v", want, got)
	}
}

func TestCatalogItem_SetDisplayOrder_ReplacesPreviousRanking(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	first := newTestCatalogItem("RANK-1", "First", "A-ANIMALS", priceGroup.ID)
	second := newTestCatalogItem("RANK-2", "Second", "A-ANIMALS", priceGroup.ID)

	for _, item := range []*CatalogItem{first, second} {
		if err := models.CatalogItems.Insert(item); err != nil {
			t.Fatalf("Failed to insert %s: %v", item.Name, err)
		}
	}

	if err := models.CatalogItems.SetDisplayOrder([]string{first.UUID, second.UUID}); err != nil {
		t.Fatalf("Failed to set initial display order: %v", err)
	}

	ranked, err := models.CatalogItems.GetRanked()
	if err != nil {
		t.Fatalf("Failed to get ranked: %v", err)
	}
	if len(ranked) != 2 {
		t.Fatalf("Expected 2 ranked items, got %d", len(ranked))
	}
	if ranked[0].DisplayOrder == nil || *ranked[0].DisplayOrder != 1 {
		t.Errorf("Expected first item at position 1, got %v", ranked[0].DisplayOrder)
	}

	// Re-ranking with only the second item must unrank the first.
	if err := models.CatalogItems.SetDisplayOrder([]string{second.UUID}); err != nil {
		t.Fatalf("Failed to replace display order: %v", err)
	}

	ranked, err = models.CatalogItems.GetRanked()
	if err != nil {
		t.Fatalf("Failed to get ranked after replace: %v", err)
	}
	if len(ranked) != 1 {
		t.Fatalf("Expected 1 ranked item after replace, got %d", len(ranked))
	}
	if ranked[0].Name != "Second" {
		t.Errorf("Expected Second to be the only ranked item, got %s", ranked[0].Name)
	}
}

func TestCatalogItem_Update_PreservesDisplayOrder(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	priceGroup := createTestPriceGroup(t, models)

	item := newTestCatalogItem("KEEP-1", "Keeper", "A-ANIMALS", priceGroup.ID)
	if err := models.CatalogItems.Insert(item); err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	if err := models.CatalogItems.SetDisplayOrder([]string{item.UUID}); err != nil {
		t.Fatalf("Failed to set display order: %v", err)
	}

	// Simulate an admin editing the item: re-read, change a field, write back.
	fetched, found, err := models.CatalogItems.GetByID(item.ID)
	if err != nil || !found {
		t.Fatalf("Failed to re-read item: %v (found=%v)", err, found)
	}
	fetched.Name = "Renamed"
	if err := models.CatalogItems.Update(fetched); err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}

	after, found, err := models.CatalogItems.GetByID(item.ID)
	if err != nil || !found {
		t.Fatalf("Failed to read item after update: %v (found=%v)", err, found)
	}
	if after.DisplayOrder == nil {
		t.Fatal("Editing a catalog item wiped its best-seller rank")
	}
	if *after.DisplayOrder != 1 {
		t.Errorf("Expected rank 1 preserved, got %d", *after.DisplayOrder)
	}
}
