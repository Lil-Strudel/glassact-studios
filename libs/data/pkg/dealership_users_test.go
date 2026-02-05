package data

import (
	"testing"
)

func TestDealershipUser_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "John Dealer",
		Email:        "john@dealership.com",
		Avatar:       "https://example.com/john.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	if user.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", user.ID)
	}
	if user.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if user.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestDealershipUser_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	original := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Jane Dealer",
		Email:        "jane@dealership.com",
		Avatar:       "https://example.com/jane.png",
		Role:         DealershipUserRoles.Submitter,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.DealershipUsers.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("User not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.Email != original.Email {
		t.Errorf("Expected email %s, got %s", original.Email, retrieved.Email)
	}
	if retrieved.Role != original.Role {
		t.Errorf("Expected role %s, got %s", original.Role, retrieved.Role)
	}
}

func TestDealershipUser_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	original := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Bob Dealer",
		Email:        "bob@dealership.com",
		Avatar:       "https://example.com/bob.png",
		Role:         DealershipUserRoles.Approver,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.DealershipUsers.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("User not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestDealershipUser_GetByEmail(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Alice Dealer",
		Email:        "alice@dealership.com",
		Avatar:       "https://example.com/alice.png",
		Role:         DealershipUserRoles.Viewer,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.DealershipUsers.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("Failed to get by email: %v", err)
	}
	if !found {
		t.Errorf("User not found by email")
	}
	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}
}

func TestDealershipUser_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	original := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Original Name",
		Email:        "original@dealership.com",
		Avatar:       "https://example.com/original.png",
		Role:         DealershipUserRoles.Viewer,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.Name = "Updated Name"
	original.Role = DealershipUserRoles.Admin
	original.IsActive = false

	err = models.DealershipUsers.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.DealershipUsers.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("User not found after update")
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Role != DealershipUserRoles.Admin {
		t.Errorf("Expected role Admin, got %s", retrieved.Role)
	}
	if retrieved.IsActive != false {
		t.Errorf("Expected IsActive to be false")
	}
}

func TestDealershipUser_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Delete Me",
		Email:        "delete@dealership.com",
		Avatar:       "https://example.com/delete.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.DealershipUsers.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.DealershipUsers.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected user to be deleted")
	}
}

func TestDealershipUser_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user1 := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "User 1",
		Email:        "user1@dealership.com",
		Avatar:       "https://example.com/user1.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	user2 := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "User 2",
		Email:        "user2@dealership.com",
		Avatar:       "https://example.com/user2.png",
		Role:         DealershipUserRoles.Submitter,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user1)
	if err != nil {
		t.Fatalf("Failed to insert user1: %v", err)
	}
	err = models.DealershipUsers.Insert(user2)
	if err != nil {
		t.Fatalf("Failed to insert user2: %v", err)
	}

	users, err := models.DealershipUsers.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}
}
