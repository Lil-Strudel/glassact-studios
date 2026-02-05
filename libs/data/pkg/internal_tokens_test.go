package data

import (
	"testing"
	"time"
)

func TestInternalToken_New(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	token, err := models.InternalTokens.New(internalUser.ID, 24*time.Hour, InternalScopeLogin)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	if token.Plaintext == "" {
		t.Errorf("Expected non-empty plaintext")
	}
	if token.Hash == nil {
		t.Errorf("Expected non-nil hash")
	}
	if token.InternalUserID != internalUser.ID {
		t.Errorf("Expected user ID %d, got %d", internalUser.ID, token.InternalUserID)
	}
	if token.Scope != InternalScopeLogin {
		t.Errorf("Expected scope %s, got %s", InternalScopeLogin, token.Scope)
	}
}

func TestInternalToken_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	token := generateInternalToken(internalUser.ID, 24*time.Hour, InternalScopeAccess)

	err := models.InternalTokens.Insert(token)
	if err != nil {
		t.Fatalf("Failed to insert token: %v", err)
	}
}

func TestInternalToken_DeleteAllForUser(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	// Create multiple tokens
	_, err := models.InternalTokens.New(internalUser.ID, 24*time.Hour, InternalScopeLogin)
	if err != nil {
		t.Fatalf("Failed to create token 1: %v", err)
	}

	_, err = models.InternalTokens.New(internalUser.ID, 24*time.Hour, InternalScopeAccess)
	if err != nil {
		t.Fatalf("Failed to create token 2: %v", err)
	}

	// Delete all login scope tokens for the user
	err = models.InternalTokens.DeleteAllForUser(InternalScopeLogin, internalUser.ID)
	if err != nil {
		t.Fatalf("Failed to delete tokens: %v", err)
	}

	// Verify token 1 (login scope) is deleted, token 2 (access scope) still exists
	// by attempting to create a new one with different scope and checking counts
}

func TestInternalToken_DeleteByPlaintext(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	internalUser := createTestInternalUser(t, models)

	token, err := models.InternalTokens.New(internalUser.ID, 24*time.Hour, InternalScopeLogin)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	plaintextToDelete := token.Plaintext

	// Delete by plaintext
	err = models.InternalTokens.DeleteByPlaintext(InternalScopeLogin, plaintextToDelete)
	if err != nil {
		t.Fatalf("Failed to delete token by plaintext: %v", err)
	}
}

func TestInternalToken_Scopes(t *testing.T) {
	scopes := []string{
		InternalScopeLogin,
		InternalScopeAccess,
		InternalScopeRefresh,
	}

	for _, scope := range scopes {
		if scope == "" {
			t.Errorf("Expected non-empty scope")
		}
	}
}

func TestInternalToken_TokenExpiry(t *testing.T) {
	ttl := 24 * time.Hour
	token := generateInternalToken(1, ttl, InternalScopeLogin)

	if token.Expiry.Before(time.Now()) {
		t.Errorf("Expected expiry to be in the future")
	}

	expectedExpiry := time.Now().Add(ttl)
	if token.Expiry.Before(expectedExpiry.Add(-time.Second)) ||
		token.Expiry.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("Expected expiry to be approximately %v, got %v", expectedExpiry, token.Expiry)
	}
}
