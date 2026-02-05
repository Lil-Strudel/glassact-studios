package data

import (
	"testing"
)

func TestDealershipAccount_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "test@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	account := &DealershipAccount{
		DealershipUserID:  user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-123",
	}

	err = models.DealershipAccounts.Insert(account)
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

func TestDealershipAccount_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "test@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	original := &DealershipAccount{
		DealershipUserID:  user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-456",
	}

	err = models.DealershipAccounts.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.DealershipAccounts.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Account not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.Provider != original.Provider {
		t.Errorf("Expected provider %s, got %s", original.Provider, retrieved.Provider)
	}
}

func TestDealershipAccount_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "test2@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	original := &DealershipAccount{
		DealershipUserID:  user.ID,
		Type:              "oauth",
		Provider:          "github",
		ProviderAccountID: "github-789",
	}

	err = models.DealershipAccounts.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.DealershipAccounts.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Account not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestDealershipAccount_GetByProvider(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "test3@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	account := &DealershipAccount{
		DealershipUserID:  user.ID,
		Type:              "oauth",
		Provider:          "azure",
		ProviderAccountID: "azure-999",
	}

	err = models.DealershipAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.DealershipAccounts.GetByProvider("azure", "azure-999")
	if err != nil {
		t.Fatalf("Failed to get by provider: %v", err)
	}
	if !found {
		t.Errorf("Account not found by provider")
	}
	if retrieved.Provider != "azure" {
		t.Errorf("Expected provider azure, got %s", retrieved.Provider)
	}
	if retrieved.ProviderAccountID != "azure-999" {
		t.Errorf("Expected provider account ID azure-999, got %s", retrieved.ProviderAccountID)
	}
}

func TestDealershipAccount_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "test4@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	original := &DealershipAccount{
		DealershipUserID:  user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-old",
	}

	err = models.DealershipAccounts.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.ProviderAccountID = "google-new"
	err = models.DealershipAccounts.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.DealershipAccounts.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Account not found after update")
	}
	if retrieved.ProviderAccountID != "google-new" {
		t.Errorf("Expected provider account ID google-new, got %s", retrieved.ProviderAccountID)
	}
}

func TestDealershipAccount_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "test5@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	account := &DealershipAccount{
		DealershipUserID:  user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-delete",
	}

	err = models.DealershipAccounts.Insert(account)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.DealershipAccounts.Delete(account.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.DealershipAccounts.GetByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected account to be deleted")
	}
}
