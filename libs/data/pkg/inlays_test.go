package data

import (
	"slices"
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

func TestInlay_GetDeleteBlockers_WithNoDependents_ReportsDeletable(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	blockers, err := models.Inlays.GetDeleteBlockers([]int{inlay.ID})
	if err != nil {
		t.Fatalf("Failed to get delete blockers: %v", err)
	}

	found, ok := blockers[inlay.ID]
	if !ok {
		t.Fatalf("Expected inlay %d to be present in the blocker map", inlay.ID)
	}
	if len(found) != 0 {
		t.Errorf("Expected no blockers for a bare inlay, got %v", found)
	}
}

func TestInlay_GetDeleteBlockers_WithProof_ReportsProofBlocker(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	inlay := createTestInlay(t, models, project.ID)
	createTestInlayProof(t, models, inlay.ID, priceGroup.ID)

	blockers, err := models.Inlays.GetDeleteBlockers([]int{inlay.ID})
	if err != nil {
		t.Fatalf("Failed to get delete blockers: %v", err)
	}

	found := blockers[inlay.ID]
	if !slices.Contains(found, InlayDeleteBlockers.Proof) {
		t.Errorf("Expected a proof blocker, got %v", found)
	}

	// A chat message accompanies the proof but cascades, so it must not appear.
	if slices.Contains(found, InlayDeleteBlockers.Milestone) ||
		slices.Contains(found, InlayDeleteBlockers.Order) {
		t.Errorf("Expected only a proof blocker, got %v", found)
	}
}

// The proof is exactly what makes a customized inlay undeletable, so confirm
// the RESTRICT actually fires and the pre-flight check agrees with it.
func TestInlay_Delete_WithExistingProof_IsRejectedByDatabase(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	inlay := createTestInlay(t, models, project.ID)
	createTestInlayProof(t, models, inlay.ID, priceGroup.ID)

	blockers, err := models.Inlays.GetDeleteBlockers([]int{inlay.ID})
	if err != nil {
		t.Fatalf("Failed to get delete blockers: %v", err)
	}
	if len(blockers[inlay.ID]) == 0 {
		t.Fatal("Expected the proof to be reported as a delete blocker")
	}

	if err := models.Inlays.Delete(inlay.ID); err == nil {
		t.Error("Expected the database to reject deleting an inlay that has a proof")
	}
}

func TestInlay_GetDeleteBlockers_BatchesMultipleInlays(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)

	blocked := createTestInlay(t, models, project.ID)
	createTestInlayProof(t, models, blocked.ID, priceGroup.ID)
	free := createTestInlay(t, models, project.ID)

	blockers, err := models.Inlays.GetDeleteBlockers([]int{blocked.ID, free.ID})
	if err != nil {
		t.Fatalf("Failed to get delete blockers: %v", err)
	}

	if len(blockers) != 2 {
		t.Fatalf("Expected results for 2 inlays, got %d", len(blockers))
	}
	if len(blockers[blocked.ID]) == 0 {
		t.Errorf("Expected blockers for the inlay with a proof")
	}
	if len(blockers[free.ID]) != 0 {
		t.Errorf("Expected no blockers for the bare inlay, got %v", blockers[free.ID])
	}
}

func TestInlay_GetDeleteBlockers_WithEmptyInput_ReturnsEmptyMap(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	blockers, err := models.Inlays.GetDeleteBlockers(nil)
	if err != nil {
		t.Fatalf("Expected no error for an empty request, got %v", err)
	}
	if len(blockers) != 0 {
		t.Errorf("Expected an empty map, got %v", blockers)
	}
}
