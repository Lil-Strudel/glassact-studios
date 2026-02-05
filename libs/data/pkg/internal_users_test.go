package data

import (
	"testing"
)

func TestInternalUser_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	user := &InternalUser{
		Name:     "Test User",
		Email:    "test@example.com",
		Avatar:   "https://example.com/avatar.png",
		Role:     InternalUserRoles.Designer,
		IsActive: true,
	}

	err := models.InternalUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	if user.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", user.ID)
	}
	if user.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInternalUser_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	created := createTestInternalUser(t, models)

	retrieved, found, err := models.InternalUsers.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if !found {
		t.Errorf("Expected user to be found")
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
	if retrieved.Email != created.Email {
		t.Errorf("Expected email %s, got %s", created.Email, retrieved.Email)
	}
}

func TestInternalUser_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	created := createTestInternalUser(t, models)

	retrieved, found, err := models.InternalUsers.GetByUUID(created.UUID)
	if err != nil {
		t.Fatalf("Failed to get user by UUID: %v", err)
	}

	if !found {
		t.Errorf("Expected user to be found")
	}

	if retrieved.UUID != created.UUID {
		t.Errorf("Expected UUID %s, got %s", created.UUID, retrieved.UUID)
	}
}

func TestInternalUser_GetByEmail(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	created := createTestInternalUser(t, models)

	retrieved, found, err := models.InternalUsers.GetByEmail(created.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if !found {
		t.Errorf("Expected user to be found")
	}

	if retrieved.Email != created.Email {
		t.Errorf("Expected email %s, got %s", created.Email, retrieved.Email)
	}
}

func TestInternalUser_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	createTestInternalUser(t, models)

	user2 := &InternalUser{
		Name:     "Another User",
		Email:    "another@example.com",
		Avatar:   "https://example.com/avatar2.png",
		Role:     InternalUserRoles.Admin,
		IsActive: true,
	}
	err := models.InternalUsers.Insert(user2)
	if err != nil {
		t.Fatalf("Failed to insert user 2: %v", err)
	}

	users, err := models.InternalUsers.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestInternalUser_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	user := createTestInternalUser(t, models)

	user.Name = "Updated Name"
	user.Role = InternalUserRoles.Admin

	err := models.InternalUsers.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, _, err := models.InternalUsers.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name to be updated, got %s", retrieved.Name)
	}
	if retrieved.Role != InternalUserRoles.Admin {
		t.Errorf("Expected role to be updated, got %s", retrieved.Role)
	}
}

func TestInternalUser_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	user := createTestInternalUser(t, models)

	err := models.InternalUsers.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, found, err := models.InternalUsers.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if found {
		t.Errorf("Expected user to be deleted")
	}
}

func TestInternalUserRoles(t *testing.T) {
	roles := []InternalUserRole{
		InternalUserRoles.Designer,
		InternalUserRoles.Production,
		InternalUserRoles.Billing,
		InternalUserRoles.Admin,
	}

	for _, role := range roles {
		if role == "" {
			t.Errorf("Expected non-empty role")
		}
	}
}
