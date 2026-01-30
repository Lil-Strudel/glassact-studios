package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful insert", func(t *testing.T) {
		user := &User{
			Name:         "John Doe",
			Email:        "john@example.com",
			Avatar:       "https://example.com/avatar.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}

		err := models.Users.Insert(user)
		require.NoError(t, err)

		assert.NotZero(t, user.ID)
		assert.NotEmpty(t, user.UUID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
		assert.Equal(t, 1, user.Version)

		_, err = uuid.Parse(user.UUID)
		assert.NoError(t, err)
	})

	t.Run("insert with admin role", func(t *testing.T) {
		user := &User{
			Name:         "Admin User",
			Email:        "admin@example.com",
			Avatar:       "https://example.com/admin.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.Admin,
		}

		err := models.Users.Insert(user)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, UserRoles.Admin, retrieved.Role)
	})

	t.Run("insert duplicate email fails", func(t *testing.T) {
		user1 := &User{
			Name:         "User One",
			Email:        "duplicate@example.com",
			Avatar:       "https://example.com/avatar1.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err := models.Users.Insert(user1)
		require.NoError(t, err)

		user2 := &User{
			Name:         "User Two",
			Email:        "duplicate@example.com",
			Avatar:       "https://example.com/avatar2.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err = models.Users.Insert(user2)
		assert.Error(t, err)
	})

	t.Run("insert with invalid dealership fails", func(t *testing.T) {
		user := &User{
			Name:         "Invalid User",
			Email:        "invalid@example.com",
			Avatar:       "https://example.com/avatar.png",
			DealershipID: 99999,
			Role:         UserRoles.User,
		}

		err := models.Users.Insert(user)
		assert.Error(t, err)
	})

	t.Run("email case insensitivity", func(t *testing.T) {
		user1 := &User{
			Name:         "Case Test",
			Email:        "CaseTest@Example.COM",
			Avatar:       "https://example.com/avatar.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err := models.Users.Insert(user1)
		require.NoError(t, err)

		user2 := &User{
			Name:         "Case Test 2",
			Email:        "casetest@example.com",
			Avatar:       "https://example.com/avatar2.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err = models.Users.Insert(user2)
		assert.Error(t, err, "citext should make email comparison case-insensitive")
	})
}

func TestUserModel_GetByID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("existing user", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.UUID, retrieved.UUID)
		assert.Equal(t, user.Name, retrieved.Name)
		assert.Equal(t, user.Email, retrieved.Email)
		assert.Equal(t, user.Avatar, retrieved.Avatar)
		assert.Equal(t, user.DealershipID, retrieved.DealershipID)
		assert.Equal(t, user.Role, retrieved.Role)
	})

	t.Run("non-existing user", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByID(99999)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("negative ID", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByID(-1)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("zero ID", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByID(0)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestUserModel_GetByUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("existing user", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByUUID(user.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.UUID, retrieved.UUID)
		assert.Equal(t, user.Name, retrieved.Name)
	})

	t.Run("non-existing UUID", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		retrieved, found, err := models.Users.GetByUUID(nonExistentUUID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		_, _, err := models.Users.GetByUUID("not-a-valid-uuid")
		assert.Error(t, err)
	})

	t.Run("empty UUID", func(t *testing.T) {
		_, _, err := models.Users.GetByUUID("")
		assert.Error(t, err)
	})
}

func TestUserModel_GetByEmail(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	user := &User{
		Name:         "Email Test User",
		Email:        "emailtest@example.com",
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	err := models.Users.Insert(user)
	require.NoError(t, err)

	t.Run("existing email", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByEmail("emailtest@example.com")
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("email case insensitive lookup", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByEmail("EMAILTEST@EXAMPLE.COM")
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, user.ID, retrieved.ID)
	})

	t.Run("non-existing email", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByEmail("nonexistent@example.com")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("empty email", func(t *testing.T) {
		retrieved, found, err := models.Users.GetByEmail("")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestUserModel_GetForToken(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("valid token", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, -time.Hour, ScopeAccess)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("wrong scope", func(t *testing.T) {
		token, err := models.Tokens.New(user.ID, time.Hour, ScopeRefresh)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetForToken(ScopeAccess, token.Plaintext)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid token plaintext", func(t *testing.T) {
		retrieved, found, err := models.Users.GetForToken(ScopeAccess, "invalid-token-plaintext")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("empty token", func(t *testing.T) {
		retrieved, found, err := models.Users.GetForToken(ScopeAccess, "")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestUserModel_GetAll(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("empty table", func(t *testing.T) {
		users, err := models.Users.GetAll()
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("multiple users", func(t *testing.T) {
		dealership := createTestDealership(t, models)

		for i := 0; i < 5; i++ {
			user := &User{
				Name:         fmt.Sprintf("User %d", i),
				Email:        fmt.Sprintf("user%d@example.com", i),
				Avatar:       fmt.Sprintf("https://example.com/avatar%d.png", i),
				DealershipID: dealership.ID,
				Role:         UserRoles.User,
			}
			err := models.Users.Insert(user)
			require.NoError(t, err)
		}

		users, err := models.Users.GetAll()
		require.NoError(t, err)
		assert.Len(t, users, 5)

		for _, u := range users {
			assert.NotZero(t, u.ID)
			assert.NotEmpty(t, u.UUID)
			assert.NotEmpty(t, u.Name)
			assert.NotEmpty(t, u.Email)
		}
	})

	t.Run("mixed roles", func(t *testing.T) {
		cleanupTables(t)
		dealership := createTestDealership(t, models)

		admin := &User{
			Name:         "Admin",
			Email:        "admin@example.com",
			Avatar:       "https://example.com/admin.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.Admin,
		}
		err := models.Users.Insert(admin)
		require.NoError(t, err)

		regularUser := &User{
			Name:         "User",
			Email:        "user@example.com",
			Avatar:       "https://example.com/user.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		err = models.Users.Insert(regularUser)
		require.NoError(t, err)

		users, err := models.Users.GetAll()
		require.NoError(t, err)
		assert.Len(t, users, 2)

		roleCount := map[UserRole]int{}
		for _, u := range users {
			roleCount[u.Role]++
		}
		assert.Equal(t, 1, roleCount[UserRoles.Admin])
		assert.Equal(t, 1, roleCount[UserRoles.User])
	})
}

func TestUserModel_Update(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful update", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)
		originalVersion := user.Version
		originalUpdatedAt := user.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		user.Name = "Updated Name"
		user.Email = "updated@example.com"
		user.Avatar = "https://example.com/new-avatar.png"
		user.Role = UserRoles.Admin

		err := models.Users.Update(user)
		require.NoError(t, err)

		assert.Equal(t, originalVersion+1, user.Version)
		assert.True(t, user.UpdatedAt.After(originalUpdatedAt) || user.UpdatedAt.Equal(originalUpdatedAt))

		retrieved, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, "updated@example.com", retrieved.Email)
		assert.Equal(t, UserRoles.Admin, retrieved.Role)
	})

	t.Run("update with wrong version fails", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)

		user.Version = 999

		user.Name = "Should Fail"
		err := models.Users.Update(user)
		assert.Error(t, err)
	})

	t.Run("update preserves ID and UUID", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)
		originalID := user.ID
		originalUUID := user.UUID

		user.Name = "New Name"
		err := models.Users.Update(user)
		require.NoError(t, err)

		assert.Equal(t, originalID, user.ID)
		assert.Equal(t, originalUUID, user.UUID)
	})

	t.Run("concurrent update detection", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)

		user1, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		require.True(t, found)

		user2, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		require.True(t, found)

		user1.Name = "First Update"
		err = models.Users.Update(user1)
		require.NoError(t, err)

		user2.Name = "Second Update"
		err = models.Users.Update(user2)
		assert.Error(t, err, "should fail due to optimistic locking")
	})
}

func TestUserModel_Delete(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful delete", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)

		err := models.Users.Delete(user.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		err := models.Users.Delete(99999)
		require.NoError(t, err)
	})

	t.Run("delete cascades to tokens", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)

		_, err := models.Tokens.New(user.ID, time.Hour, ScopeAccess)
		require.NoError(t, err)
		_, err = models.Tokens.New(user.ID, time.Hour, ScopeRefresh)
		require.NoError(t, err)

		err = models.Users.Delete(user.ID)
		require.NoError(t, err)

		_, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("delete cascades to accounts", func(t *testing.T) {
		user := createTestUser(t, models, dealership.ID)

		account := &Account{
			UserID:            user.ID,
			Type:              "oauth",
			Provider:          "google",
			ProviderAccountID: "google-123",
		}
		err := models.Accounts.Insert(account)
		require.NoError(t, err)

		err = models.Users.Delete(user.ID)
		require.NoError(t, err)

		_, found, err := models.Accounts.GetByID(account.ID)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestUserRole_Constants(t *testing.T) {
	t.Run("role values", func(t *testing.T) {
		assert.Equal(t, UserRole("admin"), UserRoles.Admin)
		assert.Equal(t, UserRole("user"), UserRoles.User)
	})

	t.Run("role string conversion", func(t *testing.T) {
		assert.Equal(t, "admin", string(UserRoles.Admin))
		assert.Equal(t, "user", string(UserRoles.User))
	})
}

func TestUser_StandardTable(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)

	t.Run("standard fields populated", func(t *testing.T) {
		assert.NotZero(t, user.StandardTable.ID)
		assert.NotEmpty(t, user.StandardTable.UUID)
		assert.NotZero(t, user.StandardTable.CreatedAt)
		assert.NotZero(t, user.StandardTable.UpdatedAt)
		assert.Equal(t, 1, user.StandardTable.Version)
	})

	t.Run("created_at immutable", func(t *testing.T) {
		originalCreatedAt := user.CreatedAt

		user.Name = "New Name"
		err := models.Users.Update(user)
		require.NoError(t, err)

		retrieved, found, err := models.Users.GetByID(user.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, originalCreatedAt.Unix(), retrieved.CreatedAt.Unix())
	})
}

func BenchmarkUserModel_Insert(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &User{
			Name:         fmt.Sprintf("Benchmark User %d", i),
			Email:        fmt.Sprintf("benchmark%d%d@example.com", i, time.Now().UnixNano()),
			Avatar:       "https://example.com/avatar.png",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}
		_ = models.Users.Insert(user)
	}
}

func BenchmarkUserModel_GetByID(b *testing.B) {
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
		Email:        fmt.Sprintf("benchmarkget%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = models.Users.GetByID(user.ID)
	}
}

func TestUserToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in userToGen", func(t *testing.T) {
		user := &User{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			Name:         "Test User",
			Email:        "test@example.com",
			DealershipID: 1,
			Role:         "user",
		}

		_, err := userToGen(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestUserModel_Update_InvalidUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("update with invalid UUID fails", func(t *testing.T) {
		user := &User{
			StandardTable: StandardTable{
				ID:      1,
				UUID:    "invalid-uuid-format",
				Version: 1,
			},
			Name:         "Test User",
			Email:        "invalid-uuid-test@example.com",
			DealershipID: dealership.ID,
			Role:         UserRoles.User,
		}

		err := models.Users.Update(user)
		assert.Error(t, err)
	})
}
