package data

import (
	"testing"
	"time"
)

func TestProject_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	project := &Project{
		DealershipID: dealership.ID,
		Name:         "Test Project",
		Status:       ProjectStatuses.Draft,
	}

	err := models.Projects.Insert(project)
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	if project.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", project.ID)
	}
	if project.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestProject_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	created := createTestProject(t, models, dealership.ID)

	retrieved, found, err := models.Projects.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if !found {
		t.Errorf("Expected project to be found")
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
	if retrieved.Name != created.Name {
		t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
	}
}

func TestProject_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	created := createTestProject(t, models, dealership.ID)

	retrieved, found, err := models.Projects.GetByUUID(created.UUID)
	if err != nil {
		t.Fatalf("Failed to get project by UUID: %v", err)
	}

	if !found {
		t.Errorf("Expected project to be found")
	}

	if retrieved.UUID != created.UUID {
		t.Errorf("Expected UUID %s, got %s", created.UUID, retrieved.UUID)
	}
}

func TestProject_GetByDealershipID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	createTestProject(t, models, dealership.ID)

	project2 := &Project{
		DealershipID: dealership.ID,
		Name:         "Another Project",
		Status:       ProjectStatuses.Draft,
	}
	err := models.Projects.Insert(project2)
	if err != nil {
		t.Fatalf("Failed to insert project 2: %v", err)
	}

	projects, err := models.Projects.GetByDealershipID(dealership.ID)
	if err != nil {
		t.Fatalf("Failed to get projects by dealership ID: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}
}

func TestProject_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	createTestProject(t, models, dealership.ID)

	project2 := &Project{
		DealershipID: dealership.ID,
		Name:         "Another Project",
		Status:       ProjectStatuses.Designing,
	}
	err := models.Projects.Insert(project2)
	if err != nil {
		t.Fatalf("Failed to insert project 2: %v", err)
	}

	projects, err := models.Projects.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all projects: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}
}

func TestProject_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	project.Name = "Updated Project"
	project.Status = ProjectStatuses.Approved

	err := models.Projects.Update(project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	retrieved, _, err := models.Projects.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if retrieved.Name != "Updated Project" {
		t.Errorf("Expected name to be updated, got %s", retrieved.Name)
	}
	if retrieved.Status != ProjectStatuses.Approved {
		t.Errorf("Expected status to be updated, got %s", retrieved.Status)
	}
}

func TestProject_UpdateWithOrderedAt(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	now := time.Now()
	project.OrderedAt = &now
	project.Status = ProjectStatuses.Ordered

	err := models.Projects.Update(project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	retrieved, _, err := models.Projects.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if retrieved.Status != ProjectStatuses.Ordered {
		t.Errorf("Expected status to be Ordered, got %s", retrieved.Status)
	}
	if retrieved.OrderedAt == nil {
		t.Errorf("Expected OrderedAt to be set")
	}
}

func TestProject_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	err := models.Projects.Delete(project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	_, found, err := models.Projects.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if found {
		t.Errorf("Expected project to be deleted")
	}
}

func TestProjectStatuses(t *testing.T) {
	statuses := []ProjectStatus{
		ProjectStatuses.Draft,
		ProjectStatuses.Designing,
		ProjectStatuses.PendingApproval,
		ProjectStatuses.Approved,
		ProjectStatuses.Ordered,
		ProjectStatuses.InProduction,
		ProjectStatuses.Shipped,
		ProjectStatuses.Delivered,
		ProjectStatuses.Invoiced,
		ProjectStatuses.Completed,
		ProjectStatuses.Cancelled,
	}

	for _, status := range statuses {
		if status == "" {
			t.Errorf("Expected non-empty status")
		}
	}
}
