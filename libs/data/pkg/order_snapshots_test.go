package data

import (
	"testing"
)

func TestOrderSnapshot_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	snapshot := &OrderSnapshot{
		ProjectID:    1,
		InlayID:      1,
		ProofID:      1,
		PriceGroupID: 1,
		PriceCents:   50000,
		Width:        100.5,
		Height:       200.5,
	}

	err := models.OrderSnapshots.Insert(snapshot)
	if err != nil {
		t.Fatalf("Failed to insert snapshot: %v", err)
	}

	if snapshot.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", snapshot.ID)
	}
	if snapshot.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if snapshot.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestOrderSnapshot_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &OrderSnapshot{
		ProjectID:    2,
		InlayID:      2,
		ProofID:      2,
		PriceGroupID: 2,
		PriceCents:   75000,
		Width:        150.0,
		Height:       250.0,
	}

	err := models.OrderSnapshots.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.OrderSnapshots.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Snapshot not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.ProjectID != original.ProjectID {
		t.Errorf("Expected project ID %d, got %d", original.ProjectID, retrieved.ProjectID)
	}
	if retrieved.PriceCents != original.PriceCents {
		t.Errorf("Expected price %d, got %d", original.PriceCents, retrieved.PriceCents)
	}
	if retrieved.Width != original.Width {
		t.Errorf("Expected width %f, got %f", original.Width, retrieved.Width)
	}
	if retrieved.Height != original.Height {
		t.Errorf("Expected height %f, got %f", original.Height, retrieved.Height)
	}
}

func TestOrderSnapshot_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := &OrderSnapshot{
		ProjectID:    3,
		InlayID:      3,
		ProofID:      3,
		PriceGroupID: 3,
		PriceCents:   100000,
		Width:        200.0,
		Height:       300.0,
	}

	err := models.OrderSnapshots.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.OrderSnapshots.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Snapshot not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestOrderSnapshot_GetByProjectID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	projectID := 4
	snapshot1 := &OrderSnapshot{
		ProjectID:    projectID,
		InlayID:      4,
		ProofID:      4,
		PriceGroupID: 4,
		PriceCents:   50000,
		Width:        100.0,
		Height:       200.0,
	}

	snapshot2 := &OrderSnapshot{
		ProjectID:    projectID,
		InlayID:      5,
		ProofID:      5,
		PriceGroupID: 5,
		PriceCents:   60000,
		Width:        120.0,
		Height:       220.0,
	}

	err := models.OrderSnapshots.Insert(snapshot1)
	if err != nil {
		t.Fatalf("Failed to insert snapshot1: %v", err)
	}
	err = models.OrderSnapshots.Insert(snapshot2)
	if err != nil {
		t.Fatalf("Failed to insert snapshot2: %v", err)
	}

	snapshots, err := models.OrderSnapshots.GetByProjectID(projectID)
	if err != nil {
		t.Fatalf("Failed to get by project ID: %v", err)
	}
	if len(snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(snapshots))
	}
}

func TestOrderSnapshot_GetByInlayID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	inlayID := 6
	snapshot := &OrderSnapshot{
		ProjectID:    5,
		InlayID:      inlayID,
		ProofID:      6,
		PriceGroupID: 6,
		PriceCents:   55000,
		Width:        110.0,
		Height:       210.0,
	}

	err := models.OrderSnapshots.Insert(snapshot)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.OrderSnapshots.GetByInlayID(inlayID)
	if err != nil {
		t.Fatalf("Failed to get by inlay ID: %v", err)
	}
	if !found {
		t.Errorf("Snapshot not found by inlay ID")
	}
	if retrieved.InlayID != inlayID {
		t.Errorf("Expected inlay ID %d, got %d", inlayID, retrieved.InlayID)
	}
}

func TestOrderSnapshot_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	snapshot := &OrderSnapshot{
		ProjectID:    6,
		InlayID:      7,
		ProofID:      7,
		PriceGroupID: 7,
		PriceCents:   45000,
		Width:        90.0,
		Height:       190.0,
	}

	err := models.OrderSnapshots.Insert(snapshot)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.OrderSnapshots.Delete(snapshot.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.OrderSnapshots.GetByID(snapshot.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected snapshot to be deleted")
	}
}

func TestOrderSnapshot_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	snapshot1 := &OrderSnapshot{
		ProjectID:    7,
		InlayID:      8,
		ProofID:      8,
		PriceGroupID: 8,
		PriceCents:   40000,
		Width:        80.0,
		Height:       180.0,
	}

	snapshot2 := &OrderSnapshot{
		ProjectID:    8,
		InlayID:      9,
		ProofID:      9,
		PriceGroupID: 9,
		PriceCents:   35000,
		Width:        70.0,
		Height:       170.0,
	}

	err := models.OrderSnapshots.Insert(snapshot1)
	if err != nil {
		t.Fatalf("Failed to insert snapshot1: %v", err)
	}
	err = models.OrderSnapshots.Insert(snapshot2)
	if err != nil {
		t.Fatalf("Failed to insert snapshot2: %v", err)
	}

	snapshots, err := models.OrderSnapshots.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(snapshots) < 2 {
		t.Errorf("Expected at least 2 snapshots, got %d", len(snapshots))
	}
}
