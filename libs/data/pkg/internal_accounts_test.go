package data

import (
	"testing"
)

func createTestInternalUser(t *testing.T, models Models) *InternalUser {
	t.Helper()

	user := &InternalUser{
		Name:     "Test Internal User",
		Email:    "internal@example.com",
		Avatar:   "https://example.com/avatar.png",
		Role:     InternalUserRoles.Designer,
		IsActive: true,
	}

	err := models.InternalUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create test internal user: %v", err)
	}

	return user
}

func TestInternalAccount_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	account := &InternalAccount{
		InternalUserID:    internalUser.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-123",
	}

	err := models.InternalAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	if account.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", account.ID)
	}
	if account.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInternalAccount_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	account := &InternalAccount{
		InternalUserID:    internalUser.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-123",
	}
	err := models.InternalAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	retrieved, found, err := models.InternalAccounts.GetByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if !found {
		t.Errorf("Expected account to be found")
	}

	if retrieved.ID != account.ID {
		t.Errorf("Expected ID %d, got %d", account.ID, retrieved.ID)
	}
	if retrieved.Provider != "google" {
		t.Errorf("Expected provider 'google', got '%s'", retrieved.Provider)
	}
}

func TestInternalAccount_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	account := &InternalAccount{
		InternalUserID:    internalUser.ID,
		Type:              "oauth",
		Provider:          "github",
		ProviderAccountID: "github-456",
	}
	err := models.InternalAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	retrieved, found, err := models.InternalAccounts.GetByUUID(account.UUID)
	if err != nil {
		t.Fatalf("Failed to get account by UUID: %v", err)
	}

	if !found {
		t.Errorf("Expected account to be found")
	}

	if retrieved.UUID != account.UUID {
		t.Errorf("Expected UUID %s, got %s", account.UUID, retrieved.UUID)
	}
}

func TestInternalAccount_GetByProvider(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	account := &InternalAccount{
		InternalUserID:    internalUser.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-789",
	}
	err := models.InternalAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	retrieved, found, err := models.InternalAccounts.GetByProvider("google", "google-789")
	if err != nil {
		t.Fatalf("Failed to get account by provider: %v", err)
	}

	if !found {
		t.Errorf("Expected account to be found")
	}

	if retrieved.ProviderAccountID != "google-789" {
		t.Errorf("Expected provider account ID 'google-789', got '%s'", retrieved.ProviderAccountID)
	}
}

func TestInternalAccount_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	account := &InternalAccount{
		InternalUserID:    internalUser.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-123",
	}
	err := models.InternalAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	account.ProviderAccountID = "google-999"
	err = models.InternalAccounts.Update(account)
	if err != nil {
		t.Fatalf("Failed to update account: %v", err)
	}

	retrieved, _, err := models.InternalAccounts.GetByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if retrieved.ProviderAccountID != "google-999" {
		t.Errorf("Expected provider account ID to be updated, got %s", retrieved.ProviderAccountID)
	}
}

func TestInternalAccount_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	account := &InternalAccount{
		InternalUserID:    internalUser.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-123",
	}
	err := models.InternalAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	err = models.InternalAccounts.Delete(account.ID)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	_, found, err := models.InternalAccounts.GetByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if found {
		t.Errorf("Expected account to be deleted")
	}
}
