package data

import (
	"testing"
)

func TestInlayUpdate_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	update := &InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: InlayUpdateTypes.Issue,
		Message:    "Dropped it during fire-polish, restarting from materials-prep",
		Step:       stringPtr("fire-polish"),
	}

	err := models.InlayUpdates.Insert(update)
	if err != nil {
		t.Fatalf("Failed to insert update: %v", err)
	}

	if update.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", update.ID)
	}
	if update.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInlayUpdate_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	original := &InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: InlayUpdateTypes.Info,
		Message:    "Ahead of schedule",
		Step:       stringPtr("cutting"),
	}

	err := models.InlayUpdates.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.InlayUpdates.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Update not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.UpdateType != original.UpdateType {
		t.Errorf("Expected update type %s, got %s", original.UpdateType, retrieved.UpdateType)
	}
	if retrieved.Message != original.Message {
		t.Errorf("Expected message %s, got %s", original.Message, retrieved.Message)
	}
}

func TestInlayUpdate_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	original := &InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: InlayUpdateTypes.Issue,
		Message:    "Equipment failure",
		Step:       stringPtr("fire-polish"),
	}

	err := models.InlayUpdates.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.InlayUpdates.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Update not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestInlayUpdate_GetByInlayID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	inlayID := inlay.ID
	update1 := &InlayUpdate{
		InlayID:    inlayID,
		UpdateType: InlayUpdateTypes.Info,
		Message:    "Message 1",
		Step:       stringPtr("ordered"),
	}

	update2 := &InlayUpdate{
		InlayID:    inlayID,
		UpdateType: InlayUpdateTypes.Issue,
		Message:    "Message 2",
		Step:       stringPtr("cutting"),
	}

	err := models.InlayUpdates.Insert(update1)
	if err != nil {
		t.Fatalf("Failed to insert update1: %v", err)
	}
	err = models.InlayUpdates.Insert(update2)
	if err != nil {
		t.Fatalf("Failed to insert update2: %v", err)
	}

	updates, err := models.InlayUpdates.GetByInlayID(inlayID)
	if err != nil {
		t.Fatalf("Failed to get by inlay ID: %v", err)
	}
	if len(updates) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}
}

func TestInlayUpdate_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	original := &InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: InlayUpdateTypes.Info,
		Message:    "Original message",
		Step:       stringPtr("packaging"),
	}

	err := models.InlayUpdates.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.UpdateType = InlayUpdateTypes.Issue
	original.Message = "Updated message"

	err = models.InlayUpdates.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.InlayUpdates.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Update not found after update")
	}
	if retrieved.UpdateType != InlayUpdateTypes.Issue {
		t.Errorf("Expected issue update type")
	}
	if retrieved.Message != "Updated message" {
		t.Errorf("Expected updated message")
	}
}

func TestInlayUpdate_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	update := &InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: InlayUpdateTypes.Info,
		Message:    "To be deleted",
		Step:       stringPtr("delivered"),
	}

	err := models.InlayUpdates.Insert(update)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.InlayUpdates.Delete(update.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.InlayUpdates.GetByID(update.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected update to be deleted")
	}
}

func TestInlayUpdate_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay1 := createTestInlay(t, models, project.ID)
	inlay2 := createTestInlay(t, models, project.ID)

	update1 := &InlayUpdate{
		InlayID:    inlay1.ID,
		UpdateType: InlayUpdateTypes.Info,
		Message:    "All test 1",
		Step:       stringPtr("ordered"),
	}

	update2 := &InlayUpdate{
		InlayID:    inlay2.ID,
		UpdateType: InlayUpdateTypes.Issue,
		Message:    "All test 2",
		Step:       stringPtr("cutting"),
	}

	err := models.InlayUpdates.Insert(update1)
	if err != nil {
		t.Fatalf("Failed to insert update1: %v", err)
	}
	err = models.InlayUpdates.Insert(update2)
	if err != nil {
		t.Fatalf("Failed to insert update2: %v", err)
	}

	updates, err := models.InlayUpdates.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(updates) < 2 {
		t.Errorf("Expected at least 2 updates, got %d", len(updates))
	}
}
