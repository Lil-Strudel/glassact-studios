package data

import (
	"testing"
	"time"
)

func TestInlayBlocker_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	blocker := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Hard,
		Reason:      "Material unavailable",
		StepBlocked: "materials-prep",
	}

	err := models.InlayBlockers.Insert(blocker)
	if err != nil {
		t.Fatalf("Failed to insert blocker: %v", err)
	}

	if blocker.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", blocker.ID)
	}
	if blocker.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInlayBlocker_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	original := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "Awaiting approval",
		StepBlocked: "cutting",
	}

	err := models.InlayBlockers.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.InlayBlockers.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Blocker not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.BlockerType != original.BlockerType {
		t.Errorf("Expected blocker type %s, got %s", original.BlockerType, retrieved.BlockerType)
	}
	if retrieved.Reason != original.Reason {
		t.Errorf("Expected reason %s, got %s", original.Reason, retrieved.Reason)
	}
}

func TestInlayBlocker_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	original := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Hard,
		Reason:      "Equipment failure",
		StepBlocked: "fire-polish",
	}

	err := models.InlayBlockers.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.InlayBlockers.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Blocker not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestInlayBlocker_GetByInlayID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	inlayID := inlay.ID
	blocker1 := &InlayBlocker{
		InlayID:     inlayID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "Reason 1",
		StepBlocked: "ordered",
	}

	blocker2 := &InlayBlocker{
		InlayID:     inlayID,
		BlockerType: BlockerTypes.Hard,
		Reason:      "Reason 2",
		StepBlocked: "cutting",
	}

	err := models.InlayBlockers.Insert(blocker1)
	if err != nil {
		t.Fatalf("Failed to insert blocker1: %v", err)
	}
	err = models.InlayBlockers.Insert(blocker2)
	if err != nil {
		t.Fatalf("Failed to insert blocker2: %v", err)
	}

	blockers, err := models.InlayBlockers.GetByInlayID(inlayID)
	if err != nil {
		t.Fatalf("Failed to get by inlay ID: %v", err)
	}
	if len(blockers) != 2 {
		t.Errorf("Expected 2 blockers, got %d", len(blockers))
	}
}

func TestInlayBlocker_GetUnresolved(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	inlayID := inlay.ID
	blocker1 := &InlayBlocker{
		InlayID:     inlayID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "Unresolved 1",
		StepBlocked: "ordered",
	}

	blocker2 := &InlayBlocker{
		InlayID:     inlayID,
		BlockerType: BlockerTypes.Hard,
		Reason:      "Resolved",
		StepBlocked: "cutting",
		ResolvedAt:  &time.Time{},
	}
	*blocker2.ResolvedAt = time.Now()

	err := models.InlayBlockers.Insert(blocker1)
	if err != nil {
		t.Fatalf("Failed to insert blocker1: %v", err)
	}
	err = models.InlayBlockers.Insert(blocker2)
	if err != nil {
		t.Fatalf("Failed to insert blocker2: %v", err)
	}

	// Update to mark as resolved
	err = models.InlayBlockers.Update(blocker2)
	if err != nil {
		t.Fatalf("Failed to update blocker2: %v", err)
	}

	unresolvedBlockers, err := models.InlayBlockers.GetUnresolved(inlayID)
	if err != nil {
		t.Fatalf("Failed to get unresolved: %v", err)
	}
	if len(unresolvedBlockers) != 1 {
		t.Errorf("Expected 1 unresolved blocker, got %d", len(unresolvedBlockers))
	}
}

func TestInlayBlocker_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	original := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "Original reason",
		StepBlocked: "packaging",
	}

	err := models.InlayBlockers.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.BlockerType = BlockerTypes.Hard
	original.Reason = "Updated reason"
	original.ResolvedAt = &time.Time{}
	*original.ResolvedAt = time.Now()
	original.ResolutionNotes = stringPtr("Issue fixed")

	err = models.InlayBlockers.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.InlayBlockers.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Blocker not found after update")
	}
	if retrieved.BlockerType != BlockerTypes.Hard {
		t.Errorf("Expected hard blocker type")
	}
	if retrieved.Reason != "Updated reason" {
		t.Errorf("Expected updated reason")
	}
	if retrieved.ResolvedAt == nil {
		t.Errorf("Expected ResolvedAt to be set")
	}
}

func TestInlayBlocker_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	blocker := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "To be deleted",
		StepBlocked: "delivered",
	}

	err := models.InlayBlockers.Insert(blocker)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.InlayBlockers.Delete(blocker.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.InlayBlockers.GetByID(blocker.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected blocker to be deleted")
	}
}

func TestInlayBlocker_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay1 := createTestInlay(t, models, project.ID)
	inlay2 := createTestInlay(t, models, project.ID)

	blocker1 := &InlayBlocker{
		InlayID:     inlay1.ID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "All test 1",
		StepBlocked: "ordered",
	}

	blocker2 := &InlayBlocker{
		InlayID:     inlay2.ID,
		BlockerType: BlockerTypes.Hard,
		Reason:      "All test 2",
		StepBlocked: "cutting",
	}

	err := models.InlayBlockers.Insert(blocker1)
	if err != nil {
		t.Fatalf("Failed to insert blocker1: %v", err)
	}
	err = models.InlayBlockers.Insert(blocker2)
	if err != nil {
		t.Fatalf("Failed to insert blocker2: %v", err)
	}

	blockers, err := models.InlayBlockers.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(blockers) < 2 {
		t.Errorf("Expected at least 2 blockers, got %d", len(blockers))
	}
}
