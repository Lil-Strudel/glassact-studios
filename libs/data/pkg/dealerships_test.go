package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDealershipModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("successful insert", func(t *testing.T) {
		dealership := &Dealership{
			Name: "Test Dealership",
			Address: Address{
				Street:     "123 Main St",
				StreetExt:  "Suite 100",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "USA",
				Latitude:   40.7128,
				Longitude:  -74.0060,
			},
		}

		err := models.Dealerships.Insert(dealership)
		require.NoError(t, err)

		assert.NotZero(t, dealership.ID)
		assert.NotEmpty(t, dealership.UUID)
		assert.NotZero(t, dealership.CreatedAt)
		assert.NotZero(t, dealership.UpdatedAt)
		assert.Equal(t, 1, dealership.Version)

		_, err = uuid.Parse(dealership.UUID)
		assert.NoError(t, err)
	})

	t.Run("insert with empty street_ext", func(t *testing.T) {
		dealership := &Dealership{
			Name: "Simple Dealership",
			Address: Address{
				Street:     "456 Oak Ave",
				StreetExt:  "",
				City:       "Los Angeles",
				State:      "CA",
				PostalCode: "90001",
				Country:    "USA",
				Latitude:   34.0522,
				Longitude:  -118.2437,
			},
		}

		err := models.Dealerships.Insert(dealership)
		require.NoError(t, err)
		assert.NotZero(t, dealership.ID)
	})

	t.Run("insert with extreme coordinates", func(t *testing.T) {
		dealership := &Dealership{
			Name: "Arctic Dealership",
			Address: Address{
				Street:     "1 Ice Road",
				StreetExt:  "",
				City:       "Arctic",
				State:      "AK",
				PostalCode: "99999",
				Country:    "USA",
				Latitude:   89.9,
				Longitude:  0,
			},
		}

		err := models.Dealerships.Insert(dealership)
		require.NoError(t, err)
	})

	t.Run("insert with negative coordinates", func(t *testing.T) {
		dealership := &Dealership{
			Name: "Sydney Dealership",
			Address: Address{
				Street:     "1 Opera House",
				StreetExt:  "",
				City:       "Sydney",
				State:      "NSW",
				PostalCode: "2000",
				Country:    "Australia",
				Latitude:   -33.8688,
				Longitude:  151.2093,
			},
		}

		err := models.Dealerships.Insert(dealership)
		require.NoError(t, err)
	})

	t.Run("insert with zero coordinates", func(t *testing.T) {
		dealership := &Dealership{
			Name: "Null Island Dealership",
			Address: Address{
				Street:     "1 Null St",
				StreetExt:  "",
				City:       "Null Island",
				State:      "NI",
				PostalCode: "00000",
				Country:    "Ocean",
				Latitude:   0,
				Longitude:  0,
			},
		}

		err := models.Dealerships.Insert(dealership)
		require.NoError(t, err)
	})
}

func TestDealershipModel_GetByID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("existing dealership", func(t *testing.T) {
		retrieved, found, err := models.Dealerships.GetByID(dealership.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, dealership.ID, retrieved.ID)
		assert.Equal(t, dealership.UUID, retrieved.UUID)
		assert.Equal(t, dealership.Name, retrieved.Name)
		assert.Equal(t, dealership.Address.Street, retrieved.Address.Street)
		assert.Equal(t, dealership.Address.StreetExt, retrieved.Address.StreetExt)
		assert.Equal(t, dealership.Address.City, retrieved.Address.City)
		assert.Equal(t, dealership.Address.State, retrieved.Address.State)
		assert.Equal(t, dealership.Address.PostalCode, retrieved.Address.PostalCode)
		assert.Equal(t, dealership.Address.Country, retrieved.Address.Country)

		assert.InDelta(t, dealership.Address.Latitude, retrieved.Address.Latitude, 0.0001)
		assert.InDelta(t, dealership.Address.Longitude, retrieved.Address.Longitude, 0.0001)
	})

	t.Run("non-existing dealership", func(t *testing.T) {
		retrieved, found, err := models.Dealerships.GetByID(99999)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("negative ID", func(t *testing.T) {
		retrieved, found, err := models.Dealerships.GetByID(-1)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("zero ID", func(t *testing.T) {
		retrieved, found, err := models.Dealerships.GetByID(0)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestDealershipModel_GetByUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("existing dealership", func(t *testing.T) {
		retrieved, found, err := models.Dealerships.GetByUUID(dealership.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, dealership.ID, retrieved.ID)
		assert.Equal(t, dealership.UUID, retrieved.UUID)
		assert.Equal(t, dealership.Name, retrieved.Name)
	})

	t.Run("non-existing UUID", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		retrieved, found, err := models.Dealerships.GetByUUID(nonExistentUUID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		_, _, err := models.Dealerships.GetByUUID("not-a-valid-uuid")
		assert.Error(t, err)
	})

	t.Run("empty UUID", func(t *testing.T) {
		_, _, err := models.Dealerships.GetByUUID("")
		assert.Error(t, err)
	})
}

func TestDealershipModel_GetAll(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("empty table", func(t *testing.T) {
		dealerships, err := models.Dealerships.GetAll()
		require.NoError(t, err)
		assert.Empty(t, dealerships)
	})

	t.Run("multiple dealerships", func(t *testing.T) {
		locations := []struct {
			city string
			lat  float64
			lng  float64
		}{
			{"New York", 40.7128, -74.0060},
			{"Los Angeles", 34.0522, -118.2437},
			{"Chicago", 41.8781, -87.6298},
			{"Houston", 29.7604, -95.3698},
			{"Phoenix", 33.4484, -112.0740},
		}

		for i, loc := range locations {
			dealership := &Dealership{
				Name: fmt.Sprintf("%s Dealership", loc.city),
				Address: Address{
					Street:     fmt.Sprintf("%d Main St", i+1),
					StreetExt:  "",
					City:       loc.city,
					State:      "XX",
					PostalCode: fmt.Sprintf("%05d", 10000+i),
					Country:    "USA",
					Latitude:   loc.lat,
					Longitude:  loc.lng,
				},
			}
			err := models.Dealerships.Insert(dealership)
			require.NoError(t, err)
		}

		dealerships, err := models.Dealerships.GetAll()
		require.NoError(t, err)
		assert.Len(t, dealerships, 5)

		for _, d := range dealerships {
			assert.NotZero(t, d.ID)
			assert.NotEmpty(t, d.UUID)
			assert.NotEmpty(t, d.Name)
			assert.NotEmpty(t, d.Address.City)
		}
	})
}

func TestDealershipModel_Update(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("successful update", func(t *testing.T) {
		dealership := createTestDealership(t, models)
		originalVersion := dealership.Version
		originalUpdatedAt := dealership.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		dealership.Name = "Updated Dealership"
		dealership.Address.Street = "999 New St"
		dealership.Address.City = "Updated City"
		dealership.Address.Latitude = 51.5074
		dealership.Address.Longitude = -0.1278

		err := models.Dealerships.Update(dealership)
		require.NoError(t, err)

		assert.Equal(t, originalVersion+1, dealership.Version)
		assert.True(t, dealership.UpdatedAt.After(originalUpdatedAt) || dealership.UpdatedAt.Equal(originalUpdatedAt))

		retrieved, found, err := models.Dealerships.GetByID(dealership.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Updated Dealership", retrieved.Name)
		assert.Equal(t, "999 New St", retrieved.Address.Street)
		assert.Equal(t, "Updated City", retrieved.Address.City)
		assert.InDelta(t, 51.5074, retrieved.Address.Latitude, 0.0001)
		assert.InDelta(t, -0.1278, retrieved.Address.Longitude, 0.0001)
	})

	t.Run("update with wrong version fails", func(t *testing.T) {
		dealership := createTestDealership(t, models)

		dealership.Version = 999

		dealership.Name = "Should Fail"
		err := models.Dealerships.Update(dealership)
		assert.Error(t, err)
	})

	t.Run("update preserves ID and UUID", func(t *testing.T) {
		dealership := createTestDealership(t, models)
		originalID := dealership.ID
		originalUUID := dealership.UUID

		dealership.Name = "New Name"
		err := models.Dealerships.Update(dealership)
		require.NoError(t, err)

		assert.Equal(t, originalID, dealership.ID)
		assert.Equal(t, originalUUID, dealership.UUID)
	})

	t.Run("concurrent update detection", func(t *testing.T) {
		dealership := createTestDealership(t, models)

		deal1, found, err := models.Dealerships.GetByID(dealership.ID)
		require.NoError(t, err)
		require.True(t, found)

		deal2, found, err := models.Dealerships.GetByID(dealership.ID)
		require.NoError(t, err)
		require.True(t, found)

		deal1.Name = "First Update"
		err = models.Dealerships.Update(deal1)
		require.NoError(t, err)

		deal2.Name = "Second Update"
		err = models.Dealerships.Update(deal2)
		assert.Error(t, err, "should fail due to optimistic locking")
	})

	t.Run("update coordinates", func(t *testing.T) {
		dealership := createTestDealership(t, models)

		dealership.Address.Latitude = -33.8688
		dealership.Address.Longitude = 151.2093

		err := models.Dealerships.Update(dealership)
		require.NoError(t, err)

		retrieved, found, err := models.Dealerships.GetByID(dealership.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.InDelta(t, -33.8688, retrieved.Address.Latitude, 0.0001)
		assert.InDelta(t, 151.2093, retrieved.Address.Longitude, 0.0001)
	})
}

func TestDealershipModel_Delete(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("successful delete", func(t *testing.T) {
		dealership := createTestDealership(t, models)

		err := models.Dealerships.Delete(dealership.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Dealerships.GetByID(dealership.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent dealership", func(t *testing.T) {
		err := models.Dealerships.Delete(99999)
		require.NoError(t, err)
	})

	t.Run("delete restricted by user reference", func(t *testing.T) {
		dealership := createTestDealership(t, models)
		_ = createTestUser(t, models, dealership.ID)

		err := models.Dealerships.Delete(dealership.ID)
		assert.Error(t, err, "should fail due to foreign key constraint")
	})

	t.Run("delete restricted by project reference", func(t *testing.T) {
		dealership := createTestDealership(t, models)
		_ = createTestProject(t, models, dealership.ID)

		err := models.Dealerships.Delete(dealership.ID)
		assert.Error(t, err, "should fail due to foreign key constraint")
	})
}

func TestAddress_Struct(t *testing.T) {
	t.Run("address fields", func(t *testing.T) {
		addr := Address{
			Street:     "123 Main St",
			StreetExt:  "Suite 100",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		}

		assert.Equal(t, "123 Main St", addr.Street)
		assert.Equal(t, "Suite 100", addr.StreetExt)
		assert.Equal(t, "Test City", addr.City)
		assert.Equal(t, "TS", addr.State)
		assert.Equal(t, "12345", addr.PostalCode)
		assert.Equal(t, "USA", addr.Country)
		assert.Equal(t, 40.7128, addr.Latitude)
		assert.Equal(t, -74.0060, addr.Longitude)
	})
}

func BenchmarkDealershipModel_Insert(b *testing.B) {
	models := NewModels(testDB.Pool, testDB.STDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dealership := &Dealership{
			Name: fmt.Sprintf("Benchmark Dealership %d", i),
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
	}
}

func BenchmarkDealershipModel_GetByID(b *testing.B) {
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
		_, _, _ = models.Dealerships.GetByID(dealership.ID)
	}
}

func TestDealershipToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in dealershipToGen", func(t *testing.T) {
		dealership := &Dealership{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			Name: "Test Dealership",
			Address: Address{
				Street:     "123 Main St",
				City:       "Test City",
				State:      "TS",
				PostalCode: "12345",
				Country:    "USA",
				Latitude:   40.7128,
				Longitude:  -74.0060,
			},
		}

		_, err := dealershipToGen(dealership)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestDealershipModel_Update_InvalidUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("update with invalid UUID fails", func(t *testing.T) {
		dealership := &Dealership{
			StandardTable: StandardTable{
				ID:      1,
				UUID:    "invalid-uuid-format",
				Version: 1,
			},
			Name: "Test Dealership",
			Address: Address{
				Street:     "123 Main St",
				City:       "Test City",
				State:      "TS",
				PostalCode: "12345",
				Country:    "USA",
				Latitude:   40.7128,
				Longitude:  -74.0060,
			},
		}

		err := models.Dealerships.Update(dealership)
		assert.Error(t, err)
	})
}
