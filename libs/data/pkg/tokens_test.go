package data

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenModel_New(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("successful token creation", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		assert.NotEmpty(t, token.Plaintext)
		assert.NotEmpty(t, token.Hash)
		assert.Equal(t, user.ID, token.UserID)
		assert.Equal(t, ScopeAccess, token.Scope)
		assert.True(t, token.Expiry.After(time.Now()))
	})

	t.Run("hash matches plaintext", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		expectedHash := sha256.Sum256([]byte(token.Plaintext))
		assert.Equal(t, expectedHash[:], token.Hash)
	})

	t.Run("creates unique tokens", func(t *testing.T) {
		token1, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		token2, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		assert.NotEqual(t, token1.Plaintext, token2.Plaintext)
		assert.NotEqual(t, token1.Hash, token2.Hash)
	})

	t.Run("different scopes", func(t *testing.T) {
		scopes := []string{ScopeLogin, ScopeAccess, ScopeRefresh}

		for _, scope := range scopes {
			token, err := models.Tokens.New(user.ID, time.Hour, scope)
			require.NoError(t, err)
			assert.Equal(t, scope, token.Scope)
		}
	})

	t.Run("different TTLs", func(t *testing.T) {
		ttls := []time.Duration{
			time.Minute,
			time.Hour,
			24 * time.Hour,
			7 * 24 * time.Hour,
		}

		for _, ttl := range ttls {
			token, err := models.Tokens.New(user.ID, ttl, ScopeAccess)
			require.NoError(t, err)

			expectedExpiry := time.Now().Add(ttl)
			assert.WithinDuration(t, expectedExpiry, token.Expiry, time.Second)
		}
	})

	t.Run("negative TTL creates expired token", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, -time.Hour, ScopeAccess)
		require.NoError(t, err)

		assert.True(t, token.Expiry.Before(time.Now()))
	})

	t.Run("zero TTL creates immediately expired token", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, 0, ScopeAccess)
		require.NoError(t, err)

		assert.WithinDuration(t, time.Now(), token.Expiry, time.Second)
	})

	t.Run("invalid user ID fails", func(t *testing.T) {
		_, err := models.Tokens.New(99999, time.Hour, ScopeAccess)
		assert.Error(t, err)
	})

	t.Run("multiple tokens for same user and scope", func(t *testing.T) {
		tokens := make([]*Token, 5)
		for i := 0; i < 5; i++ {
			token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
			require.NoError(t, err)
			tokens[i] = token
		}

		plaintexts := make(map[string]bool)
		for _, token := range tokens {
			assert.False(t, plaintexts[token.Plaintext], "duplicate plaintext found")
			plaintexts[token.Plaintext] = true
		}
	})
}

func TestTokenModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("successful insert", func(t *testing.T) {
		token := generateToken(user.ID, time.Hour, ScopeAccess)

		err := models.Tokens.Insert(token)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, user.ID, retrieved.ID)
	})

	t.Run("insert same hash fails", func(t *testing.T) {
		token := generateToken(user.ID, time.Hour, ScopeAccess)

		err := models.Tokens.Insert(token)
		require.NoError(t, err)

		err = models.Tokens.Insert(token)
		assert.Error(t, err, "should fail due to primary key constraint")
	})

	t.Run("insert for invalid user fails", func(t *testing.T) {
		token := generateToken(99999, time.Hour, ScopeAccess)

		err := models.Tokens.Insert(token)
		assert.Error(t, err)
	})

	t.Run("insert with empty scope succeeds", func(t *testing.T) {
		token := generateToken(user.ID, time.Hour, "")

		err := models.Tokens.Insert(token)
		require.NoError(t, err)
	})

	t.Run("insert with custom scope succeeds", func(t *testing.T) {
		token := generateToken(user.ID, time.Hour, "custom-scope")

		err := models.Tokens.Insert(token)
		require.NoError(t, err)
	})
}

func TestTokenModel_DeleteAllForUser(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("delete all tokens for user and scope", func(t *testing.T) {
		tokens := make([]*Token, 3)
		for i := 0; i < 3; i++ {
			token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
			require.NoError(t, err)
			tokens[i] = token
		}

		err := models.Tokens.DeleteAllForUser(ScopeAccess, user.ID)
		require.NoError(t, err)

		for _, token := range tokens {
			retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
			require.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, retrieved)
		}
	})

	t.Run("only deletes tokens with matching scope", func(t *testing.T) {
		accessToken, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		refreshToken, err := models.Tokens.New(user.ID, time.Hour, ScopeRefresh)
		require.NoError(t, err)

		err = models.Tokens.DeleteAllForUser(ScopeAccess, user.ID)
		require.NoError(t, err)

		_, found, err := models.Users.GetForToken(ScopeAccess, accessToken.Plaintext)
		require.NoError(t, err)
		assert.False(t, found)

		retrieved, found, err := models.Users.GetForToken(ScopeRefresh, refreshToken.Plaintext)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, user.ID, retrieved.ID)
	})

	t.Run("only deletes tokens for specified user", func(t *testing.T) {
		user2 := &User{
			Name:         "Second User",
			Email:        fmt.Sprintf("user2-%d@example.com", time.Now().UnixNano()),
			Avatar:       "https://example.com/avatar2.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err := models.Users.Insert(user2)
		require.NoError(t, err)

		token1, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		token2, err := models.Tokens.New(user2.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		err = models.Tokens.DeleteAllForUser(ScopeAccess, user.ID)
		require.NoError(t, err)

		_, found, err := models.Users.GetForToken(ScopeAccess, token1.Plaintext)
		require.NoError(t, err)
		assert.False(t, found)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token2.Plaintext)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, user2.ID, retrieved.ID)
	})

	t.Run("delete for non-existent user succeeds", func(t *testing.T) {
		err := models.Tokens.DeleteAllForUser(ScopeAccess, 99999)
		require.NoError(t, err)
	})

	t.Run("delete for non-existent scope succeeds", func(t *testing.T) {
		err := models.Tokens.DeleteAllForUser("non-existent-scope", user.ID)
		require.NoError(t, err)
	})

	t.Run("delete with empty scope", func(t *testing.T) {
		token := generateToken(user.ID, time.Hour, "")
		err := models.Tokens.Insert(token)
		require.NoError(t, err)

		err = models.Tokens.DeleteAllForUser("", user.ID)
		require.NoError(t, err)
	})
}

func TestTokenModel_DeleteByPlaintext(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("delete specific token", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		err = models.Tokens.DeleteByPlaintext(ScopeAccess, token.Plaintext)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("only deletes token with matching scope", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		err = models.Tokens.DeleteByPlaintext(ScopeRefresh, token.Plaintext)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, user.ID, retrieved.ID)
	})

	t.Run("delete non-existent token succeeds", func(t *testing.T) {
		err := models.Tokens.DeleteByPlaintext(ScopeAccess, "non-existent-plaintext")
		require.NoError(t, err)
	})

	t.Run("delete with empty plaintext succeeds", func(t *testing.T) {
		err := models.Tokens.DeleteByPlaintext(ScopeAccess, "")
		require.NoError(t, err)
	})

	t.Run("leaves other tokens intact", func(t *testing.T) {
		tokens := make([]*Token, 3)
		for i := 0; i < 3; i++ {
			token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
			require.NoError(t, err)
			tokens[i] = token
		}

		err := models.Tokens.DeleteByPlaintext(ScopeAccess, tokens[0].Plaintext)
		require.NoError(t, err)

		_, found, err := models.Users.GetForToken(ScopeAccess, tokens[0].Plaintext)
		require.NoError(t, err)
		assert.False(t, found)

		for i := 1; i < 3; i++ {
			retrieved, found, err := models.Users.GetForToken(ScopeAccess, tokens[i].Plaintext)
			require.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, user.ID, retrieved.ID)
		}
	})
}

func TestToken_ScopeConstants(t *testing.T) {
	t.Run("scope values", func(t *testing.T) {
		assert.Equal(t, "login", ScopeLogin)
		assert.Equal(t, "access", ScopeAccess)
		assert.Equal(t, "refresh", ScopeRefresh)
	})

	t.Run("scopes are distinct", func(t *testing.T) {
		assert.NotEqual(t, ScopeLogin, ScopeAccess)
		assert.NotEqual(t, ScopeLogin, ScopeRefresh)
		assert.NotEqual(t, ScopeAccess, ScopeRefresh)
	})
}

func TestGenerateToken(t *testing.T) {
	t.Run("generates valid token", func(t *testing.T) {
		token := generateToken(1, time.Hour, ScopeAccess)

		assert.NotEmpty(t, token.Plaintext)
		assert.NotEmpty(t, token.Hash)
		assert.Equal(t, 1, token.UserID)
		assert.Equal(t, ScopeAccess, token.Scope)
	})

	t.Run("hash is 32 bytes (SHA256)", func(t *testing.T) {
		token := generateToken(1, time.Hour, ScopeAccess)
		assert.Len(t, token.Hash, 32)
	})

	t.Run("hash is deterministic for plaintext", func(t *testing.T) {
		token := generateToken(1, time.Hour, ScopeAccess)

		expectedHash := sha256.Sum256([]byte(token.Plaintext))
		assert.Equal(t, expectedHash[:], token.Hash)
	})

	t.Run("generates unique plaintexts", func(t *testing.T) {
		plaintexts := make(map[string]bool)

		for i := 0; i < 100; i++ {
			token := generateToken(1, time.Hour, ScopeAccess)
			assert.False(t, plaintexts[token.Plaintext], "duplicate plaintext generated")
			plaintexts[token.Plaintext] = true
		}
	})

	t.Run("expiry is set correctly", func(t *testing.T) {
		before := time.Now()
		token := generateToken(1, time.Hour, ScopeAccess)
		after := time.Now()

		expectedExpiryMin := before.Add(time.Hour)
		expectedExpiryMax := after.Add(time.Hour)

		assert.True(t, !token.Expiry.Before(expectedExpiryMin) && !token.Expiry.After(expectedExpiryMax))
	})
}

func TestToken_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		token := &Token{
			Plaintext: "test-plaintext",
			Hash:      []byte("test-hash"),
			UserID:    123,
			Expiry:    time.Now().Add(time.Hour),
			Scope:     ScopeAccess,
		}

		assert.Equal(t, "test-plaintext", token.Plaintext)
		assert.Equal(t, []byte("test-hash"), token.Hash)
		assert.Equal(t, 123, token.UserID)
		assert.Equal(t, ScopeAccess, token.Scope)
	})
}

func BenchmarkTokenModel_New(b *testing.B) {
	models := NewModels(testDB.Pool, testDB.STDB)

	dealership := &Dealership{
		Name: "Benchmark Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	_ = models.Dealerships.Insert(dealership)

	user := &User{
		Name:         "Benchmark User",
		Email:        fmt.Sprintf("benchmarktoken%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = models.Tokens.New(user.ID, time.Hour, ScopeAccess)
	}
}

func BenchmarkTokenModel_DeleteAllForUser(b *testing.B) {
	models := NewModels(testDB.Pool, testDB.STDB)

	dealership := &Dealership{
		Name: "Benchmark Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	_ = models.Dealerships.Insert(dealership)

	user := &User{
		Name:         "Benchmark User",
		Email:        fmt.Sprintf("benchmarktokendel%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 5; j++ {
			_, _ = models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		}
		_ = models.Tokens.DeleteAllForUser(ScopeAccess, user.ID)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateToken(1, time.Hour, ScopeAccess)
	}
}
