package data

import (
	"testing"
)

func TestCatalogItem_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	catalogItem := &CatalogItem{
		CatalogCode:         "CAT-001",
		Name:                "Test Item",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	// Insert test item
	original := &CatalogItem{
		CatalogCode:         "CAT-002",
		Name:                "Test Item 2",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	original := &CatalogItem{
		CatalogCode:         "CAT-003",
		Name:                "Test Item 3",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	original := &CatalogItem{
		CatalogCode:         "UNIQUE-CODE",
		Name:                "Test Item",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	original := &CatalogItem{
		CatalogCode:         "CAT-004",
		Name:                "Original Name",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	item := &CatalogItem{
		CatalogCode:         "CAT-005",
		Name:                "Item to Delete",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	item := &CatalogItem{
		CatalogCode:         "CAT-006",
		Name:                "Item with Tags",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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

	item1 := &CatalogItem{
		CatalogCode:         "CAT-007",
		Name:                "Tagged Item 1",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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
		DefaultPriceGroupID: 1,
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

	item1 := &CatalogItem{
		CatalogCode:         "CAT-009",
		Name:                "Glass Item",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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
		DefaultPriceGroupID: 1,
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

	item1 := &CatalogItem{
		CatalogCode:         "CAT-011",
		Name:                "All Test 1",
		Category:            "Glass",
		DefaultWidth:        100.0,
		DefaultHeight:       200.0,
		MinWidth:            50.0,
		MinHeight:           100.0,
		DefaultPriceGroupID: 1,
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
		DefaultPriceGroupID: 1,
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
