package data

import (
	"testing"
	"time"
)

func TestInlayMilestone_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser := createTestInternalUser(t, models)

	milestone := &InlayMilestone{
		InlayID:     inlay.ID,
		Step:        ManufacturingSteps.Ordered,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser.ID,
		EventTime:   time.Now(),
	}

	err := models.InlayMilestones.Insert(milestone)
	if err != nil {
		t.Fatalf("Failed to insert milestone: %v", err)
	}

	if milestone.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", milestone.ID)
	}
	if milestone.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInlayMilestone_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser := createTestInternalUser(t, models)

	now := time.Now()
	original := &InlayMilestone{
		InlayID:     inlay.ID,
		Step:        ManufacturingSteps.MaterialsPrep,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser.ID,
		EventTime:   now,
	}

	err := models.InlayMilestones.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.InlayMilestones.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Milestone not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.Step != original.Step {
		t.Errorf("Expected step %s, got %s", original.Step, retrieved.Step)
	}
	if retrieved.EventType != original.EventType {
		t.Errorf("Expected event type %s, got %s", original.EventType, retrieved.EventType)
	}
}

func TestInlayMilestone_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser := createTestInternalUser(t, models)

	original := &InlayMilestone{
		InlayID:     inlay.ID,
		Step:        ManufacturingSteps.Cutting,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser.ID,
		EventTime:   time.Now(),
	}

	err := models.InlayMilestones.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.InlayMilestones.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Milestone not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestInlayMilestone_GetByInlayID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser := createTestInternalUser(t, models)

	inlayID := inlay.ID
	milestone1 := &InlayMilestone{
		InlayID:     inlayID,
		Step:        ManufacturingSteps.FirePolish,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser.ID,
		EventTime:   time.Now(),
	}

	milestone2 := &InlayMilestone{
		InlayID:     inlayID,
		Step:        ManufacturingSteps.FirePolish,
		EventType:   MilestoneEventTypes.Exited,
		PerformedBy: internalUser.ID,
		EventTime:   time.Now().Add(2 * time.Hour),
	}

	err := models.InlayMilestones.Insert(milestone1)
	if err != nil {
		t.Fatalf("Failed to insert milestone1: %v", err)
	}
	err = models.InlayMilestones.Insert(milestone2)
	if err != nil {
		t.Fatalf("Failed to insert milestone2: %v", err)
	}

	milestones, err := models.InlayMilestones.GetByInlayID(inlayID)
	if err != nil {
		t.Fatalf("Failed to get by inlay ID: %v", err)
	}
	if len(milestones) != 2 {
		t.Errorf("Expected 2 milestones, got %d", len(milestones))
	}
}

func TestInlayMilestone_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser1 := createTestInternalUser(t, models)
	internalUser2 := createTestInternalUser(t, models)

	original := &InlayMilestone{
		InlayID:     inlay.ID,
		Step:        ManufacturingSteps.Packaging,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser1.ID,
		EventTime:   time.Now(),
	}

	err := models.InlayMilestones.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.EventType = MilestoneEventTypes.Exited
	original.PerformedBy = internalUser2.ID

	err = models.InlayMilestones.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.InlayMilestones.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Milestone not found after update")
	}
	if retrieved.EventType != MilestoneEventTypes.Exited {
		t.Errorf("Expected event type Exited, got %s", retrieved.EventType)
	}
	if retrieved.PerformedBy != internalUser2.ID {
		t.Errorf("Expected PerformedBy %d, got %d", internalUser2.ID, retrieved.PerformedBy)
	}
}

func TestInlayMilestone_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser := createTestInternalUser(t, models)

	milestone := &InlayMilestone{
		InlayID:     inlay.ID,
		Step:        ManufacturingSteps.Shipped,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser.ID,
		EventTime:   time.Now(),
	}

	err := models.InlayMilestones.Insert(milestone)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.InlayMilestones.Delete(milestone.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.InlayMilestones.GetByID(milestone.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected milestone to be deleted")
	}
}

func TestInlayMilestone_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay1 := createTestInlay(t, models, project.ID)
	inlay2 := createTestInlay(t, models, project.ID)
	internalUser1 := createTestInternalUser(t, models)
	internalUser2 := createTestInternalUser(t, models)

	milestone1 := &InlayMilestone{
		InlayID:     inlay1.ID,
		Step:        ManufacturingSteps.Delivered,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser1.ID,
		EventTime:   time.Now(),
	}

	milestone2 := &InlayMilestone{
		InlayID:     inlay2.ID,
		Step:        ManufacturingSteps.Ordered,
		EventType:   MilestoneEventTypes.Entered,
		PerformedBy: internalUser2.ID,
		EventTime:   time.Now(),
	}

	err := models.InlayMilestones.Insert(milestone1)
	if err != nil {
		t.Fatalf("Failed to insert milestone1: %v", err)
	}
	err = models.InlayMilestones.Insert(milestone2)
	if err != nil {
		t.Fatalf("Failed to insert milestone2: %v", err)
	}

	milestones, err := models.InlayMilestones.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(milestones) < 2 {
		t.Errorf("Expected at least 2 milestones, got %d", len(milestones))
	}
}

func TestInlayMilestone_ManufacturingSteps(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)
	internalUser := createTestInternalUser(t, models)

	steps := []ManufacturingStep{
		ManufacturingSteps.Ordered,
		ManufacturingSteps.MaterialsPrep,
		ManufacturingSteps.Cutting,
		ManufacturingSteps.FirePolish,
		ManufacturingSteps.Packaging,
		ManufacturingSteps.Shipped,
		ManufacturingSteps.Delivered,
	}

	for i, step := range steps {
		milestone := &InlayMilestone{
			InlayID:     inlay.ID,
			Step:        step,
			EventType:   MilestoneEventTypes.Entered,
			PerformedBy: internalUser.ID,
			EventTime:   time.Now().Add(time.Duration(i) * time.Hour),
		}

		err := models.InlayMilestones.Insert(milestone)
		if err != nil {
			t.Fatalf("Failed to insert milestone for step %s: %v", step, err)
		}

		if milestone.Step != step {
			t.Errorf("Expected step %s, got %s", step, milestone.Step)
		}
	}
}
