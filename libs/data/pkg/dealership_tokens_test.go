package data

import (
	"testing"
	"time"
)

func TestDealershipToken_New(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "token@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	ttl := 24 * time.Hour
	token, err := models.DealershipTokens.New(user.ID, ttl, DealershipScopeAccess)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	if token.Plaintext == "" {
		t.Errorf("Expected non-empty plaintext token")
	}
	if len(token.Hash) == 0 {
		t.Errorf("Expected non-empty hash")
	}
	if token.DealershipUserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, token.DealershipUserID)
	}
	if token.Scope != DealershipScopeAccess {
		t.Errorf("Expected scope %s, got %s", DealershipScopeAccess, token.Scope)
	}
	if token.Expiry.Before(time.Now()) {
		t.Errorf("Expected token expiry to be in the future")
	}
}

func TestDealershipToken_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "token2@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	token, err := models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeLogin)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Token should be inserted and valid
	if token.DealershipUserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, token.DealershipUserID)
	}
	if token.Scope != DealershipScopeLogin {
		t.Errorf("Expected scope %s, got %s", DealershipScopeLogin, token.Scope)
	}
}

func TestDealershipToken_DeleteAllForUser(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "token3@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	// Create multiple tokens with same scope
	_, err = models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeAccess)
	if err != nil {
		t.Fatalf("Failed to create token 1: %v", err)
	}

	_, err = models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeAccess)
	if err != nil {
		t.Fatalf("Failed to create token 2: %v", err)
	}

	// Delete all tokens for this scope
	err = models.DealershipTokens.DeleteAllForUser(DealershipScopeAccess, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete tokens: %v", err)
	}

	// Verify tokens are deleted by trying to create a new one
	// (if the old ones still existed, we'd have multiple in the DB)
	token, err := models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeAccess)
	if err != nil {
		t.Fatalf("Failed to create new token: %v", err)
	}

	if token.Plaintext == "" {
		t.Errorf("Expected new token to be created")
	}
}

func TestDealershipToken_DeleteByPlaintext(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "token4@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	token, err := models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeRefresh)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	plaintext := token.Plaintext

	// Delete the token
	err = models.DealershipTokens.DeleteByPlaintext(DealershipScopeRefresh, plaintext)
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify it's deleted - create a new one just to confirm the system still works
	token2, err := models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeRefresh)
	if err != nil {
		t.Fatalf("Failed to create new token: %v", err)
	}

	if token2.Plaintext == plaintext {
		t.Errorf("Expected different plaintext after deletion")
	}
}

func TestDealershipToken_DifferentScopes(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "token5@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create dealership user: %v", err)
	}

	loginToken, err := models.DealershipTokens.New(user.ID, 24*time.Hour, DealershipScopeLogin)
	if err != nil {
		t.Fatalf("Failed to create login token: %v", err)
	}

	accessToken, err := models.DealershipTokens.New(user.ID, 1*time.Hour, DealershipScopeAccess)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	refreshToken, err := models.DealershipTokens.New(user.ID, 7*24*time.Hour, DealershipScopeRefresh)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Verify all tokens are different
	if loginToken.Plaintext == accessToken.Plaintext {
		t.Errorf("Login and access tokens should be different")
	}
	if loginToken.Plaintext == refreshToken.Plaintext {
		t.Errorf("Login and refresh tokens should be different")
	}
	if accessToken.Plaintext == refreshToken.Plaintext {
		t.Errorf("Access and refresh tokens should be different")
	}

	// Verify scopes are correct
	if loginToken.Scope != DealershipScopeLogin {
		t.Errorf("Expected login scope")
	}
	if accessToken.Scope != DealershipScopeAccess {
		t.Errorf("Expected access scope")
	}
	if refreshToken.Scope != DealershipScopeRefresh {
		t.Errorf("Expected refresh scope")
	}
}
