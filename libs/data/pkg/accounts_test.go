package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("successful insert", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-123456789",
		}

		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		assert.NotZero(t, account.ID)
		assert.NotEmpty(t, account.UUID)
		assert.NotZero(t, account.CreatedAt)
		assert.NotZero(t, account.UpdatedAt)
		assert.Equal(t, 1, account.Version)

		_, err = uuid.Parse(account.UUID)
		assert.NoError(t, err)
	})

	t.Run("insert multiple accounts for same user", func(t *testing.T) {
		account1 := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "github",
			ProviderAccountID: "github-123",
		}
		err := models.Accounts.Insert(account1)
		require.NoError(t, err)

		account2 := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "twitter",
			ProviderAccountID: "twitter-123",
		}
		err = models.Accounts.Insert(account2)
		require.NoError(t, err)

		assert.NotEqual(t, account1.ID, account2.ID)
	})

	t.Run("insert with different types", func(t *testing.T) {
		testCases := []struct {
			accountType string
			provider    string
		}{
			{"oauth", "google"},
			{"email", "local"},
			{"api_key", "internal"},
		}

		for _, tc := range testCases {
			account := &Account{
				UserID:            user.ID,
				Type:              tc.accountType,
				Provider:          tc.provider,
				ProviderAccountID: fmt.Sprintf("%s-%d", tc.provider, time.Now().UnixNano()),
			}
			err := models.Accounts.Insert(account)
			require.NoError(t, err)
			assert.Equal(t, tc.accountType, account.Type)
			assert.Equal(t, tc.provider, account.Provider)
		}
	})

	t.Run("insert with invalid user fails", func(t *testing.T) {
		account := &Account{
			UserID:            99999,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-invalid",
		}

		err := models.Accounts.Insert(account)
		assert.Error(t, err)
	})
}

func TestAccountModel_GetByID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	account := &Account{
		UserID:            user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-getbyid-test",
	}
	err := models.Accounts.Insert(account)
	require.NoError(t, err)

	t.Run("existing account", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByID(account.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, account.ID, retrieved.ID)
		assert.Equal(t, account.UUID, retrieved.UUID)
		assert.Equal(t, account.UserID, retrieved.UserID)
		assert.Equal(t, account.Type, retrieved.Type)
		assert.Equal(t, account.Provider, retrieved.Provider)
		assert.Equal(t, account.ProviderAccountID, retrieved.ProviderAccountID)
	})

	t.Run("non-existing account", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByID(99999)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("negative ID", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByID(-1)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("zero ID", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByID(0)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestAccountModel_GetByUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	account := &Account{
		UserID:            user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-getbyuuid-test",
	}
	err := models.Accounts.Insert(account)
	require.NoError(t, err)

	t.Run("existing account", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByUUID(account.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, account.ID, retrieved.ID)
		assert.Equal(t, account.UUID, retrieved.UUID)
	})

	t.Run("non-existing UUID", func(t *testing.T) {
		nonExistingUUID := uuid.New().String()
		retrieved, found, err := models.Accounts.GetByUUID(nonExistingUUID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByUUID("invalid-uuid-format")
		assert.Error(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("empty UUID", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByUUID("")
		assert.Error(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestAccountModel_GetByProvider(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	account := &Account{
		UserID:            user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: "google-unique-id-123",
	}
	err := models.Accounts.Insert(account)
	require.NoError(t, err)

	t.Run("existing provider account", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByProvider("google", "google-unique-id-123")
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, account.ID, retrieved.ID)
		assert.Equal(t, account.Provider, retrieved.Provider)
		assert.Equal(t, account.ProviderAccountID, retrieved.ProviderAccountID)
	})

	t.Run("non-existing provider", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByProvider("facebook", "facebook-123")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("existing provider but wrong account ID", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByProvider("google", "wrong-account-id")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("empty provider", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByProvider("", "google-unique-id-123")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("empty provider account ID", func(t *testing.T) {
		retrieved, found, err := models.Accounts.GetByProvider("google", "")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("multiple accounts same provider different users", func(t *testing.T) {
		user2 := &User{
			Name:         "Second User",
			Email:        fmt.Sprintf("user2-%d@example.com", time.Now().UnixNano()),
			Avatar:       "https://example.com/avatar2.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err := models.Users.Insert(user2)
		require.NoError(t, err)

		account2 := &Account{
			UserID:            user2.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-different-id-456",
		}
		err = models.Accounts.Insert(account2)
		require.NoError(t, err)

		retrieved1, found, err := models.Accounts.GetByProvider("google", "google-unique-id-123")
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, account.ID, retrieved1.ID)

		retrieved2, found, err := models.Accounts.GetByProvider("google", "google-different-id-456")
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, account2.ID, retrieved2.ID)
	})
}

func TestAccountModel_Update(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("successful update", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-update-test",
		}
		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		originalVersion := account.Version
		originalUpdatedAt := account.UpdatedAt

		time.Sleep(100 * time.Millisecond)

		account.Type = "email"
		account.Provider = "updated"
		account.ProviderAccountID = "updated-123"

		err = models.Accounts.Update(account)
		require.NoError(t, err)

		assert.Equal(t, originalVersion+1, account.Version)
		assert.True(t, account.UpdatedAt.After(originalUpdatedAt))

		retrieved, found, err := models.Accounts.GetByID(account.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "email", retrieved.Type)
		assert.Equal(t, "updated", retrieved.Provider)
	})

	t.Run("update with wrong version fails", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "github",
			ProviderAccountID: "github-version-test",
		}
		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		account.Version = 999

		err = models.Accounts.Update(account)
		assert.Error(t, err)
	})

	t.Run("update preserves ID and UUID", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "twitter",
			ProviderAccountID: "twitter-preserve-test",
		}
		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		originalID := account.ID
		originalUUID := account.UUID

		account.Type = "updated"
		err = models.Accounts.Update(account)
		require.NoError(t, err)

		assert.Equal(t, originalID, account.ID)
		assert.Equal(t, originalUUID, account.UUID)
	})

	t.Run("concurrent update detection", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "discord",
			ProviderAccountID: "discord-concurrent-test",
		}
		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		account2, found, err := models.Accounts.GetByID(account.ID)
		require.NoError(t, err)
		require.True(t, found)

		account.Type = "update1"
		err = models.Accounts.Update(account)
		require.NoError(t, err)

		account2.Type = "update2"
		err = models.Accounts.Update(account2)
		assert.Error(t, err)
	})
}

func TestAccountModel_Delete(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("successful delete", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-delete-test",
		}
		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		accountID := account.ID

		err = models.Accounts.Delete(accountID)
		require.NoError(t, err)

		_, found, err := models.Accounts.GetByID(accountID)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("delete non-existent account", func(t *testing.T) {
		err := models.Accounts.Delete(99999)
		assert.NoError(t, err)
	})

	t.Run("delete one account leaves others", func(t *testing.T) {
		account1 := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-multi-delete-1",
		}
		err := models.Accounts.Insert(account1)
		require.NoError(t, err)

		account2 := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "github",
			ProviderAccountID: "github-multi-delete-1",
		}
		err = models.Accounts.Insert(account2)
		require.NoError(t, err)

		err = models.Accounts.Delete(account1.ID)
		require.NoError(t, err)

		_, found, err := models.Accounts.GetByID(account1.ID)
		require.NoError(t, err)
		assert.False(t, found)

		_, found, err = models.Accounts.GetByID(account2.ID)
		require.NoError(t, err)
		assert.True(t, found)
	})
}

func TestAccount_StandardTable(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("standard fields populated", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-standard-test",
		}

		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		assert.NotZero(t, account.ID)
		assert.NotEmpty(t, account.UUID)
		assert.NotZero(t, account.CreatedAt)
		assert.NotZero(t, account.UpdatedAt)
		assert.Equal(t, 1, account.Version)
	})

	t.Run("created at immutable", func(t *testing.T) {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-immutable-test",
		}

		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		originalCreatedAt := account.CreatedAt

		time.Sleep(100 * time.Millisecond)

		account.Type = "email"
		err = models.Accounts.Update(account)
		require.NoError(t, err)

		assert.Equal(t, originalCreatedAt, account.CreatedAt)
	})
}

func BenchmarkAccountModel_Insert(b *testing.B) {
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
		Email:        fmt.Sprintf("benchmark%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: fmt.Sprintf("google-%d-%d", i, time.Now().UnixNano()),
		}
		_ = models.Accounts.Insert(account)
	}
}

func BenchmarkAccountModel_GetByProvider(b *testing.B) {
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
		Email:        fmt.Sprintf("benchmarkprovider%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	account := &Account{
		UserID:            user.ID,
		Type:              "oauth",
		Provider:          "google",
		ProviderAccountID: fmt.Sprintf("google-bench-%d", time.Now().UnixNano()),
	}
	_ = models.Accounts.Insert(account)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = models.Accounts.GetByProvider(account.Provider, account.ProviderAccountID)
	}
}

func TestAccountToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in accountToGen", func(t *testing.T) {
		account := &Account{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			UserID:            1,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-123",
		}

		_, err := accountToGen(account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestAccountModel_Update_InvalidUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("update with invalid UUID fails", func(t *testing.T) {
		account := &Account{
			StandardTable: StandardTable{
				ID:      1,
				UUID:    "invalid-uuid-format",
				Version: 1,
			},
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-invalid-uuid-test",
		}

		err := models.Accounts.Update(account)
		assert.Error(t, err)
	})
}
