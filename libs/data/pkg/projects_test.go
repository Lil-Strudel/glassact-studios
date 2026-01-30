package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful insert", func(t *testing.T) {
		project := &Project{
			Name:         "Test Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err := models.Projects.Insert(project)
		require.NoError(t, err)

		assert.NotZero(t, project.ID)
		assert.NotEmpty(t, project.UUID)
		assert.NotZero(t, project.CreatedAt)
		assert.NotZero(t, project.UpdatedAt)
		assert.Equal(t, 1, project.Version)

		_, err = uuid.Parse(project.UUID)
		assert.NoError(t, err)
	})

	t.Run("insert with all status values", func(t *testing.T) {
		statuses := []ProjectStatus{
			ProjectStatusi.AwaitingProof,
			ProjectStatusi.ProofInRevision,
			ProjectStatusi.AllProofsAccepted,
			ProjectStatusi.Cancelled,
			ProjectStatusi.Ordered,
			ProjectStatusi.InProduction,
			ProjectStatusi.AwaitingInvoice,
			ProjectStatusi.AwaitingPayment,
			ProjectStatusi.Completed,
		}

		for _, status := range statuses {
			project := &Project{
				Name:         fmt.Sprintf("Project - %s", status),
				Status:       status,
				Approved:     false,
				DealershipID: dealership.ID,
			}

			err := models.Projects.Insert(project)
			require.NoError(t, err, "failed to insert project with status %s", status)
			assert.Equal(t, status, project.Status)
		}
	})

	t.Run("insert with approved true", func(t *testing.T) {
		project := &Project{
			Name:         "Approved Project",
			Status:       ProjectStatusi.AllProofsAccepted,
			Approved:     true,
			DealershipID: dealership.ID,
		}

		err := models.Projects.Insert(project)
		require.NoError(t, err)

		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.True(t, retrieved.Approved)
	})

	t.Run("insert with invalid dealership fails", func(t *testing.T) {
		project := &Project{
			Name:         "Invalid Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: 99999,
		}

		err := models.Projects.Insert(project)
		assert.Error(t, err)
	})

	t.Run("insert with invalid status fails", func(t *testing.T) {
		project := &Project{
			Name:         "Invalid Status Project",
			Status:       ProjectStatus("invalid-status"),
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err := models.Projects.Insert(project)
		assert.Error(t, err)
	})

	t.Run("insert multiple projects for same dealership", func(t *testing.T) {
		projects := make([]*Project, 5)
		for i := 0; i < 5; i++ {
			project := &Project{
				Name:         fmt.Sprintf("Project %d", i),
				Status:       ProjectStatusi.AwaitingProof,
				Approved:     false,
				DealershipID: dealership.ID,
			}
			err := models.Projects.Insert(project)
			require.NoError(t, err)
			projects[i] = project
		}

		ids := make(map[int]bool)
		for _, p := range projects {
			assert.False(t, ids[p.ID], "duplicate ID found")
			ids[p.ID] = true
		}
	})

	t.Run("insert with empty name succeeds", func(t *testing.T) {
		project := &Project{
			Name:         "",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err := models.Projects.Insert(project)
		require.NoError(t, err)
	})

	t.Run("insert with long name succeeds", func(t *testing.T) {
		longName := ""
		for i := 0; i < 1000; i++ {
			longName += "a"
		}

		project := &Project{
			Name:         longName,
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err := models.Projects.Insert(project)
		require.NoError(t, err)
		assert.Equal(t, longName, project.Name)
	})
}

func TestProjectModel_TxInsert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful transaction insert", func(t *testing.T) {
		tx, err := testDB.STDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		project := &Project{
			Name:         "Transaction Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err = models.Projects.TxInsert(tx, project)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Transaction Project", retrieved.Name)
	})

	t.Run("rollback reverts insert", func(t *testing.T) {
		tx, err := testDB.STDB.Begin()
		require.NoError(t, err)

		project := &Project{
			Name:         "Rollback Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err = models.Projects.TxInsert(tx, project)
		require.NoError(t, err)

		projectID := project.ID

		err = tx.Rollback()
		require.NoError(t, err)

		_, found, err := models.Projects.GetByID(projectID)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestProjectModel_GetByID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	t.Run("existing project", func(t *testing.T) {
		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, project.ID, retrieved.ID)
		assert.Equal(t, project.UUID, retrieved.UUID)
		assert.Equal(t, project.Name, retrieved.Name)
		assert.Equal(t, project.Status, retrieved.Status)
		assert.Equal(t, project.Approved, retrieved.Approved)
		assert.Equal(t, project.DealershipID, retrieved.DealershipID)
	})

	t.Run("non-existing project", func(t *testing.T) {
		retrieved, found, err := models.Projects.GetByID(99999)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("negative ID", func(t *testing.T) {
		retrieved, found, err := models.Projects.GetByID(-1)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("zero ID", func(t *testing.T) {
		retrieved, found, err := models.Projects.GetByID(0)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestProjectModel_GetByUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	t.Run("existing project", func(t *testing.T) {
		retrieved, found, err := models.Projects.GetByUUID(project.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, project.ID, retrieved.ID)
		assert.Equal(t, project.UUID, retrieved.UUID)
		assert.Equal(t, project.Name, retrieved.Name)
	})

	t.Run("non-existing UUID", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		retrieved, found, err := models.Projects.GetByUUID(nonExistentUUID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		_, _, err := models.Projects.GetByUUID("not-a-valid-uuid")
		assert.Error(t, err)
	})

	t.Run("empty UUID", func(t *testing.T) {
		_, _, err := models.Projects.GetByUUID("")
		assert.Error(t, err)
	})
}

func TestProjectModel_GetAll(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("empty table", func(t *testing.T) {
		projects, err := models.Projects.GetAll()
		require.NoError(t, err)
		assert.Empty(t, projects)
	})

	t.Run("multiple projects", func(t *testing.T) {
		dealership := createTestDealership(t, models)

		for i := 0; i < 5; i++ {
			project := &Project{
				Name:         fmt.Sprintf("Project %d", i),
				Status:       ProjectStatusi.AwaitingProof,
				Approved:     false,
				DealershipID: dealership.ID,
			}
			err := models.Projects.Insert(project)
			require.NoError(t, err)
		}

		projects, err := models.Projects.GetAll()
		require.NoError(t, err)
		assert.Len(t, projects, 5)

		for _, p := range projects {
			assert.NotZero(t, p.ID)
			assert.NotEmpty(t, p.UUID)
			assert.NotEmpty(t, p.Name)
		}
	})

	t.Run("mixed statuses", func(t *testing.T) {
		cleanupTables(t)
		dealership := createTestDealership(t, models)

		statuses := []ProjectStatus{
			ProjectStatusi.AwaitingProof,
			ProjectStatusi.InProduction,
			ProjectStatusi.Completed,
		}

		for i, status := range statuses {
			project := &Project{
				Name:         fmt.Sprintf("Project %d", i),
				Status:       status,
				Approved:     false,
				DealershipID: dealership.ID,
			}
			err := models.Projects.Insert(project)
			require.NoError(t, err)
		}

		projects, err := models.Projects.GetAll()
		require.NoError(t, err)
		assert.Len(t, projects, 3)

		statusCount := map[ProjectStatus]int{}
		for _, p := range projects {
			statusCount[p.Status]++
		}
		for _, status := range statuses {
			assert.Equal(t, 1, statusCount[status])
		}
	})

	t.Run("projects from multiple dealerships", func(t *testing.T) {
		cleanupTables(t)

		dealership1 := createTestDealership(t, models)
		dealership2 := &Dealership{
			Name: "Second Dealership",
			Address: Address{
				Street:     "456 Other St",
				StreetExt:  "",
				City:       "Other City",
				State:      "OS",
				PostalCode: "67890",
				Country:    "USA",
				Latitude:   34.0522,
				Longitude:  -118.2437,
			},
		}
		err := models.Dealerships.Insert(dealership2)
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			p := &Project{
				Name:         fmt.Sprintf("Dealership1 Project %d", i),
				Status:       ProjectStatusi.AwaitingProof,
				Approved:     false,
				DealershipID: dealership1.ID,
			}
			err := models.Projects.Insert(p)
			require.NoError(t, err)
		}

		for i := 0; i < 2; i++ {
			p := &Project{
				Name:         fmt.Sprintf("Dealership2 Project %d", i),
				Status:       ProjectStatusi.AwaitingProof,
				Approved:     false,
				DealershipID: dealership2.ID,
			}
			err := models.Projects.Insert(p)
			require.NoError(t, err)
		}

		projects, err := models.Projects.GetAll()
		require.NoError(t, err)
		assert.Len(t, projects, 5)
	})
}

func TestProjectModel_Update(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful update", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)
		originalVersion := project.Version
		originalUpdatedAt := project.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		project.Name = "Updated Project Name"
		project.Status = ProjectStatusi.InProduction
		project.Approved = true

		err := models.Projects.Update(project)
		require.NoError(t, err)

		assert.Equal(t, originalVersion+1, project.Version)
		assert.True(t, project.UpdatedAt.After(originalUpdatedAt) || project.UpdatedAt.Equal(originalUpdatedAt))

		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Updated Project Name", retrieved.Name)
		assert.Equal(t, ProjectStatusi.InProduction, retrieved.Status)
		assert.True(t, retrieved.Approved)
	})

	t.Run("update with wrong version fails", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)

		project.Version = 999

		project.Name = "Should Fail"
		err := models.Projects.Update(project)
		assert.Error(t, err)
	})

	t.Run("update preserves ID and UUID", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)
		originalID := project.ID
		originalUUID := project.UUID

		project.Name = "New Name"
		err := models.Projects.Update(project)
		require.NoError(t, err)

		assert.Equal(t, originalID, project.ID)
		assert.Equal(t, originalUUID, project.UUID)
	})

	t.Run("concurrent update detection", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)

		proj1, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)

		proj2, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)

		proj1.Name = "First Update"
		err = models.Projects.Update(proj1)
		require.NoError(t, err)

		proj2.Name = "Second Update"
		err = models.Projects.Update(proj2)
		assert.Error(t, err, "should fail due to optimistic locking")
	})

	t.Run("update status transitions", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)

		transitions := []ProjectStatus{
			ProjectStatusi.ProofInRevision,
			ProjectStatusi.AllProofsAccepted,
			ProjectStatusi.Ordered,
			ProjectStatusi.InProduction,
			ProjectStatusi.AwaitingInvoice,
			ProjectStatusi.AwaitingPayment,
			ProjectStatusi.Completed,
		}

		for _, status := range transitions {
			project, found, err := models.Projects.GetByID(project.ID)
			require.NoError(t, err)
			require.True(t, found)

			project.Status = status
			err = models.Projects.Update(project)
			require.NoError(t, err)

			retrieved, found, err := models.Projects.GetByID(project.ID)
			require.NoError(t, err)
			require.True(t, found)
			assert.Equal(t, status, retrieved.Status)
		}
	})

	t.Run("update to invalid status fails", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)

		project.Status = ProjectStatus("invalid-status")
		err := models.Projects.Update(project)
		assert.Error(t, err)
	})

	t.Run("update dealership ID", func(t *testing.T) {
		dealership2 := &Dealership{
			Name: "Second Dealership",
			Address: Address{
				Street:     "456 Other St",
				StreetExt:  "",
				City:       "Other City",
				State:      "OS",
				PostalCode: "67890",
				Country:    "USA",
				Latitude:   34.0522,
				Longitude:  -118.2437,
			},
		}
		err := models.Dealerships.Insert(dealership2)
		require.NoError(t, err)

		project := createTestProject(t, models, dealership.ID)
		project.DealershipID = dealership2.ID
		err = models.Projects.Update(project)
		require.NoError(t, err)

		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, dealership2.ID, retrieved.DealershipID)
	})
}

func TestProjectModel_Delete(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("successful delete", func(t *testing.T) {
		project := createTestProject(t, models, dealership.ID)

		err := models.Projects.Delete(project.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent project", func(t *testing.T) {
		err := models.Projects.Delete(99999)
		require.NoError(t, err)
	})

	t.Run("delete one project leaves others", func(t *testing.T) {
		project1 := &Project{
			Name:         "Project 1",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}
		err := models.Projects.Insert(project1)
		require.NoError(t, err)

		project2 := &Project{
			Name:         "Project 2",
			Status:       ProjectStatusi.InProduction,
			Approved:     false,
			DealershipID: dealership.ID,
		}
		err = models.Projects.Insert(project2)
		require.NoError(t, err)

		err = models.Projects.Delete(project1.ID)
		require.NoError(t, err)

		retrieved, found, err := models.Projects.GetByID(project2.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrieved)
	})

	t.Run("delete with negative ID", func(t *testing.T) {
		err := models.Projects.Delete(-1)
		require.NoError(t, err)
	})

	t.Run("delete with zero ID", func(t *testing.T) {
		err := models.Projects.Delete(0)
		require.NoError(t, err)
	})
}

func TestProjectStatus_Constants(t *testing.T) {
	t.Run("status values", func(t *testing.T) {
		assert.Equal(t, ProjectStatus("awaiting-proof"), ProjectStatusi.AwaitingProof)
		assert.Equal(t, ProjectStatus("proof-in-revision"), ProjectStatusi.ProofInRevision)
		assert.Equal(t, ProjectStatus("all-proofs-accepted"), ProjectStatusi.AllProofsAccepted)
		assert.Equal(t, ProjectStatus("cancelled"), ProjectStatusi.Cancelled)
		assert.Equal(t, ProjectStatus("ordered"), ProjectStatusi.Ordered)
		assert.Equal(t, ProjectStatus("in-production"), ProjectStatusi.InProduction)
		assert.Equal(t, ProjectStatus("awaiting-invoice"), ProjectStatusi.AwaitingInvoice)
		assert.Equal(t, ProjectStatus("awaiting-payment"), ProjectStatusi.AwaitingPayment)
		assert.Equal(t, ProjectStatus("completed"), ProjectStatusi.Completed)
	})

	t.Run("status string conversion", func(t *testing.T) {
		assert.Equal(t, "awaiting-proof", string(ProjectStatusi.AwaitingProof))
		assert.Equal(t, "in-production", string(ProjectStatusi.InProduction))
		assert.Equal(t, "completed", string(ProjectStatusi.Completed))
	})

	t.Run("all statuses are distinct", func(t *testing.T) {
		statuses := []ProjectStatus{
			ProjectStatusi.AwaitingProof,
			ProjectStatusi.ProofInRevision,
			ProjectStatusi.AllProofsAccepted,
			ProjectStatusi.Cancelled,
			ProjectStatusi.Ordered,
			ProjectStatusi.InProduction,
			ProjectStatusi.AwaitingInvoice,
			ProjectStatusi.AwaitingPayment,
			ProjectStatusi.Completed,
		}

		seen := make(map[ProjectStatus]bool)
		for _, s := range statuses {
			assert.False(t, seen[s], "duplicate status found: %s", s)
			seen[s] = true
		}
	})
}

func TestProject_StandardTable(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	t.Run("standard fields populated", func(t *testing.T) {
		assert.NotZero(t, project.StandardTable.ID)
		assert.NotEmpty(t, project.StandardTable.UUID)
		assert.NotZero(t, project.StandardTable.CreatedAt)
		assert.NotZero(t, project.StandardTable.UpdatedAt)
		assert.Equal(t, 1, project.StandardTable.Version)
	})

	t.Run("created_at immutable", func(t *testing.T) {
		originalCreatedAt := project.CreatedAt

		project.Name = "New Name"
		err := models.Projects.Update(project)
		require.NoError(t, err)

		retrieved, found, err := models.Projects.GetByID(project.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, originalCreatedAt.Unix(), retrieved.CreatedAt.Unix())
	})
}

func BenchmarkProjectModel_Insert(b *testing.B) {
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
		project := &Project{
			Name:         fmt.Sprintf("Benchmark Project %d", i),
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}
		_ = models.Projects.Insert(project)
	}
}

func BenchmarkProjectModel_GetByID(b *testing.B) {
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
		_, _, _ = models.Projects.GetByID(project.ID)
	}
}

func BenchmarkProjectModel_GetAll(b *testing.B) {
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

	for i := 0; i < 100; i++ {
		project := &Project{
			Name:         fmt.Sprintf("Benchmark Project %d", i),
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}
		_ = models.Projects.Insert(project)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = models.Projects.GetAll()
	}
}

func BenchmarkProjectModel_Update(b *testing.B) {
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
		project, _, _ = models.Projects.GetByID(project.ID)
		project.Name = fmt.Sprintf("Updated %d", i)
		_ = models.Projects.Update(project)
	}
}

func TestProjectToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in projectToGen", func(t *testing.T) {
		project := &Project{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			Name:         "Test Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: 1,
		}

		_, err := projectToGen(project)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestProjectModel_Update_InvalidUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)

	t.Run("update with invalid UUID fails", func(t *testing.T) {
		project := &Project{
			StandardTable: StandardTable{
				ID:      1,
				UUID:    "invalid-uuid-format",
				Version: 1,
			},
			Name:         "Test Project",
			Status:       ProjectStatusi.AwaitingProof,
			Approved:     false,
			DealershipID: dealership.ID,
		}

		err := models.Projects.Update(project)
		assert.Error(t, err)
	})
}
