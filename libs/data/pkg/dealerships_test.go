package data

import (
	"testing"
)

func TestDealership_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	dealership := &Dealership{
		Name: "Test Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "Suite 100",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}

	err := models.Dealerships.Insert(dealership)
	if err != nil {
		t.Fatalf("Failed to insert dealership: %v", err)
	}

	if dealership.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", dealership.ID)
	}
	if dealership.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestDealership_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	created := createTestDealership(t, models)

	retrieved, found, err := models.Dealerships.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get dealership: %v", err)
	}

	if !found {
		t.Errorf("Expected dealership to be found")
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
	if retrieved.Name != created.Name {
		t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
	}
	if retrieved.Address.City != created.Address.City {
		t.Errorf("Expected city %s, got %s", created.Address.City, retrieved.Address.City)
	}
}

func TestDealership_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	created := createTestDealership(t, models)

	retrieved, found, err := models.Dealerships.GetByUUID(created.UUID)
	if err != nil {
		t.Fatalf("Failed to get dealership by UUID: %v", err)
	}

	if !found {
		t.Errorf("Expected dealership to be found")
	}

	if retrieved.UUID != created.UUID {
		t.Errorf("Expected UUID %s, got %s", created.UUID, retrieved.UUID)
	}
}

func TestDealership_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	// Create multiple dealerships
	createTestDealership(t, models)
	dealership2 := &Dealership{
		Name: "Another Dealership",
		Address: Address{
			Street:     "456 Oak Ave",
			City:       "Another City",
			State:      "AC",
			PostalCode: "54321",
			Country:    "USA",
			Latitude:   34.0522,
			Longitude:  -118.2437,
		},
	}
	err := models.Dealerships.Insert(dealership2)
	if err != nil {
		t.Fatalf("Failed to insert second dealership: %v", err)
	}

	dealerships, err := models.Dealerships.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all dealerships: %v", err)
	}

	if len(dealerships) != 2 {
		t.Errorf("Expected 2 dealerships, got %d", len(dealerships))
	}
}

func TestDealership_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	dealership.Name = "Updated Dealership"
	dealership.Address.City = "Updated City"

	err := models.Dealerships.Update(dealership)
	if err != nil {
		t.Fatalf("Failed to update dealership: %v", err)
	}

	retrieved, _, err := models.Dealerships.GetByID(dealership.ID)
	if err != nil {
		t.Fatalf("Failed to get dealership: %v", err)
	}

	if retrieved.Name != "Updated Dealership" {
		t.Errorf("Expected name to be updated, got %s", retrieved.Name)
	}
	if retrieved.Address.City != "Updated City" {
		t.Errorf("Expected city to be updated, got %s", retrieved.Address.City)
	}
}

func TestDealership_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	err := models.Dealerships.Delete(dealership.ID)
	if err != nil {
		t.Fatalf("Failed to delete dealership: %v", err)
	}

	_, found, err := models.Dealerships.GetByID(dealership.ID)
	if err != nil {
		t.Fatalf("Failed to get dealership: %v", err)
	}

	if found {
		t.Errorf("Expected dealership to be deleted")
	}
}
