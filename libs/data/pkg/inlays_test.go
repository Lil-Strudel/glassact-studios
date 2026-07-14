package data

import (
	"testing"
)

func TestInlay_InsertCatalog(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Catalog Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test customization notes",
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	if inlay.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", inlay.ID)
	}
	if inlay.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if inlay.CatalogInfo.ID == 0 {
		t.Errorf("Expected non-zero CatalogInfo ID, got %d", inlay.CatalogInfo.ID)
	}
}

func TestInlay_InsertCustom(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Custom Inlay",
		Type:       InlayTypes.Custom,
		PreviewURL: "https://example.com/preview.png",
		CustomInfo: &InlayCustomInfo{
			Description:     "Custom inlay description",
			RequestedWidth:  100.0,
			RequestedHeight: 150.0,
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	if inlay.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", inlay.ID)
	}
	if inlay.CustomInfo.ID == 0 {
		t.Errorf("Expected non-zero CustomInfo ID, got %d", inlay.CustomInfo.ID)
	}
}

func TestInlay_InsertCustom_WithReferenceImages_ReturnsThemOrdered(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Custom Inlay",
		Type:       InlayTypes.Custom,
		PreviewURL: "",
		CustomInfo: &InlayCustomInfo{
			Description:     "Custom inlay description",
			RequestedWidth:  100.0,
			RequestedHeight: 150.0,
			ReferenceImages: []InlayCustomReferenceImage{
				{ImageURL: "/file/inlay-references/first.png"},
				{ImageURL: "/file/inlay-references/second.png"},
				{ImageURL: "/file/inlay-references/third.png"},
			},
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	retrieved, found, err := models.Inlays.GetByUUID(inlay.UUID)
	if err != nil {
		t.Fatalf("Failed to get inlay by UUID: %v", err)
	}
	if !found {
		t.Fatalf("Expected inlay to be found")
	}
	if retrieved.CustomInfo == nil {
		t.Fatalf("Expected custom info to be present")
	}

	images := retrieved.CustomInfo.ReferenceImages
	if len(images) != 3 {
		t.Fatalf("Expected 3 reference images, got %d", len(images))
	}

	expected := []string{
		"/file/inlay-references/first.png",
		"/file/inlay-references/second.png",
		"/file/inlay-references/third.png",
	}
	for i, want := range expected {
		if images[i].ImageURL != want {
			t.Errorf("Expected image %d URL %q, got %q", i, want, images[i].ImageURL)
		}
		if images[i].SortOrder != i {
			t.Errorf("Expected image %d SortOrder %d, got %d", i, i, images[i].SortOrder)
		}
		if images[i].ID == 0 {
			t.Errorf("Expected image %d to have a non-zero ID", i)
		}
	}
}

func TestInlay_ReplaceReferenceImages_RewritesFullSet(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Custom Inlay",
		Type:       InlayTypes.Custom,
		PreviewURL: "",
		CustomInfo: &InlayCustomInfo{
			Description:     "Custom inlay description",
			RequestedWidth:  100.0,
			RequestedHeight: 150.0,
			ReferenceImages: []InlayCustomReferenceImage{
				{ImageURL: "/file/inlay-references/old-1.png"},
				{ImageURL: "/file/inlay-references/old-2.png"},
			},
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	err = models.Inlays.ReplaceReferenceImages(inlay.CustomInfo.ID, []string{
		"/file/inlay-references/new-1.png",
	})
	if err != nil {
		t.Fatalf("Failed to replace reference images: %v", err)
	}

	retrieved, _, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to get inlay: %v", err)
	}

	images := retrieved.CustomInfo.ReferenceImages
	if len(images) != 1 {
		t.Fatalf("Expected 1 reference image after replace, got %d", len(images))
	}
	if images[0].ImageURL != "/file/inlay-references/new-1.png" {
		t.Errorf("Expected replaced image URL, got %q", images[0].ImageURL)
	}

	// Replacing with an empty set clears all images.
	err = models.Inlays.ReplaceReferenceImages(inlay.CustomInfo.ID, []string{})
	if err != nil {
		t.Fatalf("Failed to clear reference images: %v", err)
	}

	cleared, _, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to get inlay: %v", err)
	}
	if len(cleared.CustomInfo.ReferenceImages) != 0 {
		t.Errorf("Expected 0 reference images after clearing, got %d", len(cleared.CustomInfo.ReferenceImages))
	}
}

func TestInlay_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	retrieved, found, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to get inlay: %v", err)
	}

	if !found {
		t.Errorf("Expected inlay to be found")
	}

	if retrieved.ID != inlay.ID {
		t.Errorf("Expected ID %d, got %d", inlay.ID, retrieved.ID)
	}
	if retrieved.Name != "Test Inlay" {
		t.Errorf("Expected name 'Test Inlay', got '%s'", retrieved.Name)
	}
	if retrieved.Type != InlayTypes.Catalog {
		t.Errorf("Expected type Catalog, got %s", retrieved.Type)
	}
}

func TestInlay_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	retrieved, found, err := models.Inlays.GetByUUID(inlay.UUID)
	if err != nil {
		t.Fatalf("Failed to get inlay by UUID: %v", err)
	}

	if !found {
		t.Errorf("Expected inlay to be found")
	}

	if retrieved.UUID != inlay.UUID {
		t.Errorf("Expected UUID %s, got %s", inlay.UUID, retrieved.UUID)
	}
}

func TestInlay_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay1 := &Inlay{
		ProjectID:  project.ID,
		Name:       "Inlay 1",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview1.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Notes 1",
		},
	}
	err := models.Inlays.Insert(inlay1)
	if err != nil {
		t.Fatalf("Failed to insert inlay 1: %v", err)
	}

	inlay2 := &Inlay{
		ProjectID:  project.ID,
		Name:       "Inlay 2",
		Type:       InlayTypes.Custom,
		PreviewURL: "https://example.com/preview2.png",
		CustomInfo: &InlayCustomInfo{
			Description:     "Custom description",
			RequestedWidth:  100.0,
			RequestedHeight: 150.0,
		},
	}
	err = models.Inlays.Insert(inlay2)
	if err != nil {
		t.Fatalf("Failed to insert inlay 2: %v", err)
	}

	inlays, err := models.Inlays.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all inlays: %v", err)
	}

	if len(inlays) != 2 {
		t.Errorf("Expected 2 inlays, got %d", len(inlays))
	}
}

func TestInlay_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Original notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	inlay.Name = "Updated Inlay"
	inlay.CatalogInfo.CustomizationNotes = "Updated notes"

	err = models.Inlays.Update(inlay)
	if err != nil {
		t.Fatalf("Failed to update inlay: %v", err)
	}

	retrieved, _, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to get inlay: %v", err)
	}

	if retrieved.Name != "Updated Inlay" {
		t.Errorf("Expected name to be updated, got %s", retrieved.Name)
	}
	if retrieved.CatalogInfo.CustomizationNotes != "Updated notes" {
		t.Errorf("Expected notes to be updated, got %s", retrieved.CatalogInfo.CustomizationNotes)
	}
}

func TestInlay_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	err = models.Inlays.Delete(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to delete inlay: %v", err)
	}

	_, found, err := models.Inlays.GetByID(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to get inlay: %v", err)
	}

	if found {
		t.Errorf("Expected inlay to be deleted")
	}
}
