package data

import (
	"testing"
)

func TestPriceGroup_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	priceGroup := &PriceGroup{
		Name:           "Standard Glass",
		BasePriceCents: 50000,
		Description:    stringPtr("Standard glass pricing"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(priceGroup)
	if err != nil {
		t.Fatalf("Failed to insert price group: %v", err)
	}

	if priceGroup.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", priceGroup.ID)
	}
	if priceGroup.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if priceGroup.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestPriceGroup_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &PriceGroup{
		Name:           "Premium Glass",
		BasePriceCents: 75000,
		Description:    stringPtr("Premium glass pricing"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.PriceGroups.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Price group not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, retrieved.Name)
	}
	if retrieved.BasePriceCents != original.BasePriceCents {
		t.Errorf("Expected price %d, got %d", original.BasePriceCents, retrieved.BasePriceCents)
	}
}

func TestPriceGroup_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &PriceGroup{
		Name:           "Deluxe Glass",
		BasePriceCents: 100000,
		Description:    stringPtr("Deluxe glass pricing"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.PriceGroups.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Price group not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestPriceGroup_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &PriceGroup{
		Name:           "Budget Glass",
		BasePriceCents: 25000,
		Description:    stringPtr("Budget glass pricing"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.Name = "Updated Budget Glass"
	original.BasePriceCents = 30000
	original.IsActive = false

	err = models.PriceGroups.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.PriceGroups.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Price group not found after update")
	}
	if retrieved.Name != "Updated Budget Glass" {
		t.Errorf("Expected updated name 'Updated Budget Glass', got %s", retrieved.Name)
	}
	if retrieved.BasePriceCents != 30000 {
		t.Errorf("Expected updated price 30000, got %d", retrieved.BasePriceCents)
	}
	if retrieved.IsActive != false {
		t.Errorf("Expected IsActive to be false")
	}
}

func TestPriceGroup_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	priceGroup := &PriceGroup{
		Name:           "Temporary Price Group",
		BasePriceCents: 10000,
		Description:    stringPtr("Temporary pricing"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(priceGroup)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.PriceGroups.Delete(priceGroup.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.PriceGroups.GetByID(priceGroup.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected price group to be deleted")
	}
}

func TestPriceGroup_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	priceGroup1 := &PriceGroup{
		Name:           "All Test 1",
		BasePriceCents: 40000,
		Description:    stringPtr("Test pricing 1"),
		IsActive:       true,
	}

	priceGroup2 := &PriceGroup{
		Name:           "All Test 2",
		BasePriceCents: 60000,
		Description:    stringPtr("Test pricing 2"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(priceGroup1)
	if err != nil {
		t.Fatalf("Failed to insert priceGroup1: %v", err)
	}
	err = models.PriceGroups.Insert(priceGroup2)
	if err != nil {
		t.Fatalf("Failed to insert priceGroup2: %v", err)
	}

	priceGroups, err := models.PriceGroups.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(priceGroups) < 2 {
		t.Errorf("Expected at least 2 price groups, got %d", len(priceGroups))
	}
}

func TestPriceGroup_VersionIncrement(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	priceGroup := &PriceGroup{
		Name:           "Version Test",
		BasePriceCents: 50000,
		Description:    stringPtr("Test version tracking"),
		IsActive:       true,
	}

	err := models.PriceGroups.Insert(priceGroup)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	initialVersion := priceGroup.Version
	if initialVersion == 0 {
		t.Errorf("Expected non-zero initial version, got 0")
	}

	priceGroup.Name = "Updated Name"
	err = models.PriceGroups.Update(priceGroup)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	if priceGroup.Version <= initialVersion {
		t.Errorf("Expected version to increment, was %d, now %d", initialVersion, priceGroup.Version)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
