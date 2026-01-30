package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestCatalogInlay(t *testing.T, models Models, projectID int, catalogItemID int) *Inlay {
	t.Helper()

	inlay := &Inlay{
		ProjectID:  projectID,
		Name:       "Test Catalog Inlay",
		PreviewURL: "https://example.com/preview.png",
		PriceGroup: 1,
		Type:       InlayTypes.Catalog,
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID: catalogItemID,
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to create test catalog inlay: %v", err)
	}

	return inlay
}

func createTestCustomInlay(t *testing.T, models Models, projectID int) *Inlay {
	t.Helper()

	inlay := &Inlay{
		ProjectID:  projectID,
		Name:       "Test Custom Inlay",
		PreviewURL: "https://example.com/custom-preview.png",
		PriceGroup: 2,
		Type:       InlayTypes.Custom,
		CustomInfo: &InlayCustomInfo{
			Description: "A custom glass inlay",
			Width:       24.5,
			Height:      36.75,
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to create test custom inlay: %v", err)
	}

	return inlay
}

func TestInlayModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)

	t.Run("successful insert catalog inlay", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Catalog Inlay",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Catalog,
			CatalogInfo: &InlayCatalogInfo{
				CatalogItemID: catalogItemID,
			},
		}

		err := models.Inlays.Insert(inlay)
		require.NoError(t, err)

		assert.NotZero(t, inlay.ID)
		assert.NotEmpty(t, inlay.UUID)
		assert.NotZero(t, inlay.CreatedAt)
		assert.NotZero(t, inlay.UpdatedAt)
		assert.Equal(t, 1, inlay.Version)

		assert.NotNil(t, inlay.CatalogInfo)
		assert.NotZero(t, inlay.CatalogInfo.ID)
		assert.NotEmpty(t, inlay.CatalogInfo.UUID)
		assert.Equal(t, inlay.ID, inlay.CatalogInfo.InlayID)
		assert.Equal(t, catalogItemID, inlay.CatalogInfo.CatalogItemID)

		_, err = uuid.Parse(inlay.UUID)
		assert.NoError(t, err)
		_, err = uuid.Parse(inlay.CatalogInfo.UUID)
		assert.NoError(t, err)
	})

	t.Run("successful insert custom inlay", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Custom Inlay",
			PreviewURL: "https://example.com/custom.png",
			PriceGroup: 3,
			Type:       InlayTypes.Custom,
			CustomInfo: &InlayCustomInfo{
				Description: "A beautiful custom design",
				Width:       18.5,
				Height:      24.0,
			},
		}

		err := models.Inlays.Insert(inlay)
		require.NoError(t, err)

		assert.NotZero(t, inlay.ID)
		assert.NotEmpty(t, inlay.UUID)
		assert.Equal(t, InlayTypes.Custom, inlay.Type)

		assert.NotNil(t, inlay.CustomInfo)
		assert.NotZero(t, inlay.CustomInfo.ID)
		assert.NotEmpty(t, inlay.CustomInfo.UUID)
		assert.Equal(t, inlay.ID, inlay.CustomInfo.InlayID)
		assert.Equal(t, "A beautiful custom design", inlay.CustomInfo.Description)
		assert.Equal(t, 18.5, inlay.CustomInfo.Width)
		assert.Equal(t, 24.0, inlay.CustomInfo.Height)
	})

	t.Run("insert catalog inlay without CatalogInfo fails", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:   project.ID,
			Name:        "Missing CatalogInfo",
			PreviewURL:  "https://example.com/preview.png",
			PriceGroup:  1,
			Type:        InlayTypes.Catalog,
			CatalogInfo: nil,
		}

		err := models.Inlays.Insert(inlay)
		assert.Error(t, err)
	})

	t.Run("insert custom inlay without CustomInfo fails", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Missing CustomInfo",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Custom,
			CustomInfo: nil,
		}

		err := models.Inlays.Insert(inlay)
		assert.Error(t, err)
	})

	t.Run("insert with invalid project fails", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:  99999,
			Name:       "Invalid Project Inlay",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Catalog,
			CatalogInfo: &InlayCatalogInfo{
				CatalogItemID: catalogItemID,
			},
		}

		err := models.Inlays.Insert(inlay)
		assert.Error(t, err)
	})

	t.Run("insert with invalid type fails", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Invalid Type Inlay",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayType("invalid-type"),
		}

		err := models.Inlays.Insert(inlay)
		assert.Error(t, err)
	})

	t.Run("insert catalog inlay with invalid catalog item fails", func(t *testing.T) {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Invalid Catalog Item",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Catalog,
			CatalogInfo: &InlayCatalogInfo{
				CatalogItemID: 99999,
			},
		}

		err := models.Inlays.Insert(inlay)
		assert.Error(t, err)
	})

	t.Run("insert multiple inlays for same project", func(t *testing.T) {
		inlays := make([]*Inlay, 5)
		for i := 0; i < 5; i++ {
			inlay := &Inlay{
				ProjectID:  project.ID,
				Name:       fmt.Sprintf("Inlay %d", i),
				PreviewURL: fmt.Sprintf("https://example.com/inlay%d.png", i),
				PriceGroup: i + 1,
				Type:       InlayTypes.Catalog,
				CatalogInfo: &InlayCatalogInfo{
					CatalogItemID: catalogItemID,
				},
			}
			err := models.Inlays.Insert(inlay)
			require.NoError(t, err)
			inlays[i] = inlay
		}

		ids := make(map[int]bool)
		for _, inlay := range inlays {
			assert.False(t, ids[inlay.ID], "duplicate ID found")
			ids[inlay.ID] = true
		}
	})

	t.Run("insert custom inlay with edge case dimensions", func(t *testing.T) {
		testCases := []struct {
			name   string
			width  float64
			height float64
		}{
			{"zero dimensions", 0, 0},
			{"very small dimensions", 0.001, 0.001},
			{"very large dimensions", 10000.5, 20000.75},
			{"negative dimensions", -10.5, -20.5},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				inlay := &Inlay{
					ProjectID:  project.ID,
					Name:       tc.name,
					PreviewURL: "https://example.com/preview.png",
					PriceGroup: 1,
					Type:       InlayTypes.Custom,
					CustomInfo: &InlayCustomInfo{
						Description: tc.name,
						Width:       tc.width,
						Height:      tc.height,
					},
				}

				err := models.Inlays.Insert(inlay)
				require.NoError(t, err)
				assert.Equal(t, tc.width, inlay.CustomInfo.Width)
				assert.Equal(t, tc.height, inlay.CustomInfo.Height)
			})
		}
	})

	t.Run("insert with all price groups", func(t *testing.T) {
		for pg := 0; pg <= 10; pg++ {
			inlay := &Inlay{
				ProjectID:  project.ID,
				Name:       fmt.Sprintf("Price Group %d", pg),
				PreviewURL: "https://example.com/preview.png",
				PriceGroup: pg,
				Type:       InlayTypes.Catalog,
				CatalogInfo: &InlayCatalogInfo{
					CatalogItemID: catalogItemID,
				},
			}

			err := models.Inlays.Insert(inlay)
			require.NoError(t, err)
			assert.Equal(t, pg, inlay.PriceGroup)
		}
	})
}

func TestInlayModel_TxInsert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)

	t.Run("successful transaction insert", func(t *testing.T) {
		tx, err := testDB.STDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Transaction Inlay",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Catalog,
			CatalogInfo: &InlayCatalogInfo{
				CatalogItemID: catalogItemID,
			},
		}

		err = models.Inlays.TxInsert(tx, inlay)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Transaction Inlay", retrieved.Name)
		assert.NotNil(t, retrieved.CatalogInfo)
	})

	t.Run("rollback reverts insert", func(t *testing.T) {
		tx, err := testDB.STDB.Begin()
		require.NoError(t, err)

		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       "Rollback Inlay",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Custom,
			CustomInfo: &InlayCustomInfo{
				Description: "Will be rolled back",
				Width:       10.0,
				Height:      20.0,
			},
		}

		err = models.Inlays.TxInsert(tx, inlay)
		require.NoError(t, err)

		inlayID := inlay.ID

		err = tx.Rollback()
		require.NoError(t, err)

		_, found, err := models.Inlays.GetByID(inlayID)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestInlayModel_GetByID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)

	t.Run("existing catalog inlay", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, inlay.ID, retrieved.ID)
		assert.Equal(t, inlay.UUID, retrieved.UUID)
		assert.Equal(t, inlay.Name, retrieved.Name)
		assert.Equal(t, inlay.Type, retrieved.Type)
		assert.Equal(t, inlay.PreviewURL, retrieved.PreviewURL)
		assert.Equal(t, inlay.PriceGroup, retrieved.PriceGroup)
		assert.Equal(t, inlay.ProjectID, retrieved.ProjectID)

		assert.NotNil(t, retrieved.CatalogInfo)
		assert.Equal(t, catalogItemID, retrieved.CatalogInfo.CatalogItemID)
		assert.Nil(t, retrieved.CustomInfo)
	})

	t.Run("existing custom inlay", func(t *testing.T) {
		inlay := createTestCustomInlay(t, models, project.ID)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, inlay.ID, retrieved.ID)
		assert.Equal(t, InlayTypes.Custom, retrieved.Type)

		assert.NotNil(t, retrieved.CustomInfo)
		assert.Equal(t, inlay.CustomInfo.Description, retrieved.CustomInfo.Description)
		assert.Equal(t, inlay.CustomInfo.Width, retrieved.CustomInfo.Width)
		assert.Equal(t, inlay.CustomInfo.Height, retrieved.CustomInfo.Height)
		assert.Nil(t, retrieved.CatalogInfo)
	})

	t.Run("non-existing inlay", func(t *testing.T) {
		retrieved, found, err := models.Inlays.GetByID(99999)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("negative ID", func(t *testing.T) {
		retrieved, found, err := models.Inlays.GetByID(-1)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("zero ID", func(t *testing.T) {
		retrieved, found, err := models.Inlays.GetByID(0)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestInlayModel_GetByUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)

	t.Run("existing catalog inlay", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

		retrieved, found, err := models.Inlays.GetByUUID(inlay.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, inlay.ID, retrieved.ID)
		assert.Equal(t, inlay.UUID, retrieved.UUID)
		assert.NotNil(t, retrieved.CatalogInfo)
	})

	t.Run("existing custom inlay", func(t *testing.T) {
		inlay := createTestCustomInlay(t, models, project.ID)

		retrieved, found, err := models.Inlays.GetByUUID(inlay.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, inlay.ID, retrieved.ID)
		assert.NotNil(t, retrieved.CustomInfo)
	})

	t.Run("non-existing UUID", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		retrieved, found, err := models.Inlays.GetByUUID(nonExistentUUID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		_, _, err := models.Inlays.GetByUUID("not-a-valid-uuid")
		assert.Error(t, err)
	})

	t.Run("empty UUID", func(t *testing.T) {
		_, _, err := models.Inlays.GetByUUID("")
		assert.Error(t, err)
	})
}

func TestInlayModel_GetAll(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("empty table", func(t *testing.T) {
		inlays, err := models.Inlays.GetAll()
		require.NoError(t, err)
		assert.Empty(t, inlays)
	})

	t.Run("multiple inlays mixed types", func(t *testing.T) {
		dealership := createTestDealership(t, models)
		project := createTestProject(t, models, dealership.ID)
		catalogItemID := createTestCatalogItem(t, models)

		for i := 0; i < 3; i++ {
			createTestCatalogInlay(t, models, project.ID, catalogItemID)
		}

		for i := 0; i < 2; i++ {
			createTestCustomInlay(t, models, project.ID)
		}

		inlays, err := models.Inlays.GetAll()
		require.NoError(t, err)
		assert.Len(t, inlays, 5)

		catalogCount := 0
		customCount := 0
		for _, inlay := range inlays {
			if inlay.Type == InlayTypes.Catalog {
				catalogCount++
				assert.NotNil(t, inlay.CatalogInfo)
				assert.Nil(t, inlay.CustomInfo)
			} else if inlay.Type == InlayTypes.Custom {
				customCount++
				assert.NotNil(t, inlay.CustomInfo)
				assert.Nil(t, inlay.CatalogInfo)
			}
		}
		assert.Equal(t, 3, catalogCount)
		assert.Equal(t, 2, customCount)
	})

	t.Run("inlays from multiple projects", func(t *testing.T) {
		cleanupTables(t)

		dealership := createTestDealership(t, models)
		project1 := createTestProject(t, models, dealership.ID)
		project2 := &Project{
			Name:         "Second Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}
		err := models.Projects.Insert(project2)
		require.NoError(t, err)

		catalogItemID := createTestCatalogItem(t, models)

		createTestCatalogInlay(t, models, project1.ID, catalogItemID)
		createTestCatalogInlay(t, models, project1.ID, catalogItemID)
		createTestCustomInlay(t, models, project2.ID)

		inlays, err := models.Inlays.GetAll()
		require.NoError(t, err)
		assert.Len(t, inlays, 3)

		project1Count := 0
		project2Count := 0
		for _, inlay := range inlays {
			if inlay.ProjectID == project1.ID {
				project1Count++
			} else if inlay.ProjectID == project2.ID {
				project2Count++
			}
		}
		assert.Equal(t, 2, project1Count)
		assert.Equal(t, 1, project2Count)
	})
}

func TestInlayModel_Update(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)

	t.Run("successful update catalog inlay", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)
		originalVersion := inlay.Version
		originalUpdatedAt := inlay.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		inlay.Name = "Updated Catalog Inlay"
		inlay.PreviewURL = "https://example.com/new-preview.png"
		inlay.PriceGroup = 5

		err := models.Inlays.Update(inlay)
		require.NoError(t, err)

		assert.Equal(t, originalVersion+1, inlay.Version)
		assert.True(t, inlay.UpdatedAt.After(originalUpdatedAt) || inlay.UpdatedAt.Equal(originalUpdatedAt))

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Updated Catalog Inlay", retrieved.Name)
		assert.Equal(t, "https://example.com/new-preview.png", retrieved.PreviewURL)
		assert.Equal(t, 5, retrieved.PriceGroup)
	})

	t.Run("successful update custom inlay", func(t *testing.T) {
		inlay := createTestCustomInlay(t, models, project.ID)

		inlay.Name = "Updated Custom Inlay"
		inlay.CustomInfo.Description = "Updated description"
		inlay.CustomInfo.Width = 100.0
		inlay.CustomInfo.Height = 200.0

		err := models.Inlays.Update(inlay)
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Updated Custom Inlay", retrieved.Name)
		assert.Equal(t, "Updated description", retrieved.CustomInfo.Description)
		assert.Equal(t, 100.0, retrieved.CustomInfo.Width)
		assert.Equal(t, 200.0, retrieved.CustomInfo.Height)
	})

	t.Run("update with wrong version fails", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

		inlay.Version = 999

		inlay.Name = "Should Fail"
		err := models.Inlays.Update(inlay)
		assert.Error(t, err)
	})

	t.Run("concurrent update detection", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

		inlay1, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)

		inlay2, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)

		inlay1.Name = "First Update"
		err = models.Inlays.Update(inlay1)
		require.NoError(t, err)

		inlay2.Name = "Second Update"
		err = models.Inlays.Update(inlay2)
		assert.Error(t, err, "should fail due to optimistic locking")
	})

	t.Run("update preserves ID and UUID", func(t *testing.T) {
		inlay := createTestCustomInlay(t, models, project.ID)
		originalID := inlay.ID
		originalUUID := inlay.UUID

		inlay.Name = "New Name"
		err := models.Inlays.Update(inlay)
		require.NoError(t, err)

		assert.Equal(t, originalID, inlay.ID)
		assert.Equal(t, originalUUID, inlay.UUID)
	})

	t.Run("update catalog info catalog item ID", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

		newCatalogItemID := createTestCatalogItem(t, models)

		inlay.CatalogInfo.CatalogItemID = newCatalogItemID
		err := models.Inlays.Update(inlay)
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, newCatalogItemID, retrieved.CatalogInfo.CatalogItemID)
	})
}

func TestInlayModel_Delete(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)

	t.Run("successful delete catalog inlay", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

		err := models.Inlays.Delete(inlay.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("successful delete custom inlay", func(t *testing.T) {
		inlay := createTestCustomInlay(t, models, project.ID)

		err := models.Inlays.Delete(inlay.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("delete cascades to catalog info", func(t *testing.T) {
		inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)
		catalogInfoID := inlay.CatalogInfo.ID

		err := models.Inlays.Delete(inlay.ID)
		require.NoError(t, err)

		var count int
		err = testDB.STDB.QueryRow("SELECT COUNT(*) FROM inlay_catalog_infos WHERE id = $1", catalogInfoID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete cascades to custom info", func(t *testing.T) {
		inlay := createTestCustomInlay(t, models, project.ID)
		customInfoID := inlay.CustomInfo.ID

		err := models.Inlays.Delete(inlay.ID)
		require.NoError(t, err)

		var count int
		err = testDB.STDB.QueryRow("SELECT COUNT(*) FROM inlay_custom_infos WHERE id = $1", customInfoID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete non-existent inlay", func(t *testing.T) {
		err := models.Inlays.Delete(99999)
		require.NoError(t, err)
	})

	t.Run("delete one inlay leaves others", func(t *testing.T) {
		inlay1 := createTestCatalogInlay(t, models, project.ID, catalogItemID)
		inlay2 := createTestCustomInlay(t, models, project.ID)

		err := models.Inlays.Delete(inlay1.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay2.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrieved)
	})
}

func TestInlayType_Constants(t *testing.T) {
	t.Run("type values", func(t *testing.T) {
		assert.Equal(t, InlayType("catalog"), InlayTypes.Catalog)
		assert.Equal(t, InlayType("custom"), InlayTypes.Custom)
	})

	t.Run("type string conversion", func(t *testing.T) {
		assert.Equal(t, "catalog", string(InlayTypes.Catalog))
		assert.Equal(t, "custom", string(InlayTypes.Custom))
	})

	t.Run("types are distinct", func(t *testing.T) {
		assert.NotEqual(t, InlayTypes.Catalog, InlayTypes.Custom)
	})
}

func TestInlay_StandardTable(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)

	t.Run("standard fields populated", func(t *testing.T) {
		assert.NotZero(t, inlay.StandardTable.ID)
		assert.NotEmpty(t, inlay.StandardTable.UUID)
		assert.NotZero(t, inlay.StandardTable.CreatedAt)
		assert.NotZero(t, inlay.StandardTable.UpdatedAt)
		assert.Equal(t, 1, inlay.StandardTable.Version)
	})

	t.Run("created_at immutable", func(t *testing.T) {
		originalCreatedAt := inlay.CreatedAt

		inlay.Name = "New Name"
		err := models.Inlays.Update(inlay)
		require.NoError(t, err)

		retrieved, found, err := models.Inlays.GetByID(inlay.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, originalCreatedAt.Unix(), retrieved.CreatedAt.Unix())
	})
}

func TestInlayCatalogInfo_Struct(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	catalogItemID := createTestCatalogItem(t, models)
	inlay := createTestCatalogInlay(t, models, project.ID, catalogItemID)

	t.Run("all fields populated", func(t *testing.T) {
		assert.NotZero(t, inlay.CatalogInfo.ID)
		assert.NotEmpty(t, inlay.CatalogInfo.UUID)
		assert.NotZero(t, inlay.CatalogInfo.CreatedAt)
		assert.NotZero(t, inlay.CatalogInfo.UpdatedAt)
		assert.Equal(t, 1, inlay.CatalogInfo.Version)
		assert.Equal(t, inlay.ID, inlay.CatalogInfo.InlayID)
		assert.Equal(t, catalogItemID, inlay.CatalogInfo.CatalogItemID)
	})
}

func TestInlayCustomInfo_Struct(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)

	t.Run("all fields populated", func(t *testing.T) {
		assert.NotZero(t, inlay.CustomInfo.ID)
		assert.NotEmpty(t, inlay.CustomInfo.UUID)
		assert.NotZero(t, inlay.CustomInfo.CreatedAt)
		assert.NotZero(t, inlay.CustomInfo.UpdatedAt)
		assert.Equal(t, 1, inlay.CustomInfo.Version)
		assert.Equal(t, inlay.ID, inlay.CustomInfo.InlayID)
		assert.NotEmpty(t, inlay.CustomInfo.Description)
		assert.NotZero(t, inlay.CustomInfo.Width)
		assert.NotZero(t, inlay.CustomInfo.Height)
	})
}

func BenchmarkInlayModel_Insert_Catalog(b *testing.B) {
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

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	var catalogItemID int
	_ = testDB.STDB.QueryRow("INSERT INTO catalog_items DEFAULT VALUES RETURNING id").Scan(&catalogItemID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       fmt.Sprintf("Benchmark Inlay %d", i),
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Catalog,
			CatalogInfo: &InlayCatalogInfo{
				CatalogItemID: catalogItemID,
			},
		}
		_ = models.Inlays.Insert(inlay)
	}
}

func BenchmarkInlayModel_Insert_Custom(b *testing.B) {
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

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       fmt.Sprintf("Benchmark Inlay %d", i),
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 2,
			Type:       InlayTypes.Custom,
			CustomInfo: &InlayCustomInfo{
				Description: "Benchmark custom inlay",
				Width:       24.5,
				Height:      36.75,
			},
		}
		_ = models.Inlays.Insert(inlay)
	}
}

func BenchmarkInlayModel_GetByID(b *testing.B) {
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

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Benchmark Inlay",
		PreviewURL: "https://example.com/preview.png",
		PriceGroup: 2,
		Type:       InlayTypes.Custom,
		CustomInfo: &InlayCustomInfo{
			Description: "Benchmark custom inlay",
			Width:       24.5,
			Height:      36.75,
		},
	}
	_ = models.Inlays.Insert(inlay)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = models.Inlays.GetByID(inlay.ID)
	}
}

func BenchmarkInlayModel_GetAll(b *testing.B) {
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

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	for i := 0; i < 50; i++ {
		inlay := &Inlay{
			ProjectID:  project.ID,
			Name:       fmt.Sprintf("Benchmark Inlay %d", i),
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 2,
			Type:       InlayTypes.Custom,
			CustomInfo: &InlayCustomInfo{
				Description: "Benchmark custom inlay",
				Width:       24.5,
				Height:      36.75,
			},
		}
		_ = models.Inlays.Insert(inlay)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = models.Inlays.GetAll()
	}
}

func TestInlayToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in inlayToGen", func(t *testing.T) {
		inlay := &Inlay{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			ProjectID:  1,
			Name:       "Test Inlay",
			PreviewURL: "https://example.com/preview.png",
			PriceGroup: 1,
			Type:       InlayTypes.Custom,
			CustomInfo: &InlayCustomInfo{
				Description: "Test",
				Width:       24.5,
				Height:      36.75,
			},
		}

		_, err := inlayToGen(inlay)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestCatalogInfoToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in catalogInfoToGen", func(t *testing.T) {
		catalogInfo := &InlayCatalogInfo{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			CatalogItemID: 1,
		}

		_, err := catalogInfoToGen(catalogInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestCustomInfoToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in customInfoToGen", func(t *testing.T) {
		customInfo := &InlayCustomInfo{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			Description: "Test custom info",
			Width:       24.5,
			Height:      36.75,
		}

		_, err := customInfoToGen(customInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}
