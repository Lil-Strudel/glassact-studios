package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findStatusCount(counts []StatusCount, status string) int64 {
	for _, c := range counts {
		if c.Status == status {
			return c.Count
		}
	}
	return 0
}

func findStepCount(counts []ManufacturingStepCount, step string) int64 {
	for _, c := range counts {
		if c.Step == step {
			return c.Count
		}
	}
	return 0
}

func insertProjectWithStatus(t *testing.T, models Models, dealershipID int, status ProjectStatus) *Project {
	t.Helper()
	project := &Project{
		DealershipID: dealershipID,
		Name:         "Dashboard Test Project",
		Status:       status,
	}
	err := models.Projects.Insert(project)
	require.NoError(t, err)
	return project
}

func TestGetDealershipDashboard_WithProjects_ReturnsCorrectStatusCounts(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealershipA := createTestDealership(t, models)
	dealershipB := createTestDealership(t, models)

	insertProjectWithStatus(t, models, dealershipA.ID, ProjectStatuses.Draft)
	insertProjectWithStatus(t, models, dealershipA.ID, ProjectStatuses.Draft)
	insertProjectWithStatus(t, models, dealershipA.ID, ProjectStatuses.Ordered)
	insertProjectWithStatus(t, models, dealershipB.ID, ProjectStatuses.Draft)

	dashboard, err := models.Dashboard.GetDealershipDashboard(dealershipA.ID)
	require.NoError(t, err)

	assert.Equal(t, int64(2), findStatusCount(dashboard.ProjectStatusCounts, string(ProjectStatuses.Draft)))
	assert.Equal(t, int64(1), findStatusCount(dashboard.ProjectStatusCounts, string(ProjectStatuses.Ordered)))
	assert.Len(t, dashboard.ProjectStatusCounts, 2)
	assert.Len(t, dashboard.RecentProjects, 3)
}

func TestGetDealershipDashboard_WithPendingProof_ReturnsPendingApprovalCount(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	inlay := createTestInlay(t, models, project.ID)
	createTestInlayProof(t, models, inlay.ID, priceGroup.ID)

	dashboard, err := models.Dashboard.GetDealershipDashboard(dealership.ID)
	require.NoError(t, err)

	assert.Equal(t, int64(1), dashboard.PendingApprovalCount)
}

func TestGetDealershipDashboard_WithSentInvoice_ReturnsOutstandingAmount(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	inlay := createTestInlay(t, models, project.ID)
	proof := createTestInlayProof(t, models, inlay.ID, priceGroup.ID)

	snapshot := &OrderSnapshot{
		ProjectID:    project.ID,
		InlayID:      inlay.ID,
		ProofID:      &proof.ID,
		PriceGroupID: priceGroup.ID,
		PriceCents:   75000,
		Width:        100.0,
		Height:       150.0,
	}
	require.NoError(t, models.OrderSnapshots.Insert(snapshot))

	invoice := &Invoice{
		ProjectID: project.ID,
		Status:    InvoiceStatuses.Sent,
	}
	require.NoError(t, models.Invoices.Insert(invoice))

	dashboard, err := models.Dashboard.GetDealershipDashboard(dealership.ID)
	require.NoError(t, err)

	assert.Equal(t, int64(1), dashboard.OutstandingInvoiceCount)
	assert.Equal(t, int64(75000), dashboard.OutstandingInvoiceAmountCents)
}

func TestGetDealershipDashboard_EmptyDealership_ReturnsZeros(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)

	dashboard, err := models.Dashboard.GetDealershipDashboard(dealership.ID)
	require.NoError(t, err)

	assert.Empty(t, dashboard.ProjectStatusCounts)
	assert.Equal(t, int64(0), dashboard.PendingApprovalCount)
	assert.Equal(t, int64(0), dashboard.OutstandingInvoiceCount)
	assert.Equal(t, int64(0), dashboard.OutstandingInvoiceAmountCents)
	assert.Empty(t, dashboard.RecentProjects)
}

func TestGetInternalDashboard_WithProjects_ReturnsAllStatusCounts(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealershipA := createTestDealership(t, models)
	dealershipB := createTestDealership(t, models)

	insertProjectWithStatus(t, models, dealershipA.ID, ProjectStatuses.Draft)
	insertProjectWithStatus(t, models, dealershipA.ID, ProjectStatuses.Ordered)
	insertProjectWithStatus(t, models, dealershipB.ID, ProjectStatuses.Draft)
	insertProjectWithStatus(t, models, dealershipB.ID, ProjectStatuses.Ordered)

	dashboard, err := models.Dashboard.GetInternalDashboard()
	require.NoError(t, err)

	assert.Equal(t, int64(2), findStatusCount(dashboard.ProjectStatusCounts, string(ProjectStatuses.Draft)))
	assert.Equal(t, int64(2), findStatusCount(dashboard.ProjectStatusCounts, string(ProjectStatuses.Ordered)))
}

func TestGetInternalDashboard_WithActiveBlockers_ReturnsBlockerCounts(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	hardBlocker := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Hard,
		Reason:      "Material unavailable",
		StepBlocked: "materials-prep",
	}
	require.NoError(t, models.InlayBlockers.Insert(hardBlocker))

	softBlocker := &InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: BlockerTypes.Soft,
		Reason:      "Awaiting clarification",
		StepBlocked: "cutting",
	}
	require.NoError(t, models.InlayBlockers.Insert(softBlocker))

	dashboard, err := models.Dashboard.GetInternalDashboard()
	require.NoError(t, err)

	assert.Equal(t, int64(2), dashboard.ActiveBlockerCount)
	assert.Equal(t, int64(1), dashboard.HardBlockerCount)
}

func TestGetInternalDashboard_WithManufacturingInlays_ReturnsStepCounts(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	inlay1 := createTestInlay(t, models, project.ID)
	inlay2 := createTestInlay(t, models, project.ID)
	inlay3 := createTestInlay(t, models, project.ID)
	createTestInlay(t, models, project.ID)

	_, err := testDB.STDB.Exec(`UPDATE inlays SET manufacturing_step = $1 WHERE id = $2`, "materials-prep", inlay1.ID)
	require.NoError(t, err)
	_, err = testDB.STDB.Exec(`UPDATE inlays SET manufacturing_step = $1 WHERE id = $2`, "materials-prep", inlay2.ID)
	require.NoError(t, err)
	_, err = testDB.STDB.Exec(`UPDATE inlays SET manufacturing_step = $1 WHERE id = $2`, "cutting", inlay3.ID)
	require.NoError(t, err)

	dashboard, err := models.Dashboard.GetInternalDashboard()
	require.NoError(t, err)

	assert.Equal(t, int64(2), findStepCount(dashboard.ManufacturingStepCounts, "materials-prep"))
	assert.Equal(t, int64(1), findStepCount(dashboard.ManufacturingStepCounts, "cutting"))
	assert.Len(t, dashboard.ManufacturingStepCounts, 2)
}

func TestGetInternalDashboard_EmptyDatabase_ReturnsZeros(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	dashboard, err := models.Dashboard.GetInternalDashboard()
	require.NoError(t, err)

	assert.Empty(t, dashboard.ProjectStatusCounts)
	assert.Empty(t, dashboard.ManufacturingStepCounts)
	assert.Equal(t, int64(0), dashboard.ActiveBlockerCount)
	assert.Equal(t, int64(0), dashboard.HardBlockerCount)
	assert.Equal(t, int64(0), dashboard.PendingProofCount)
	assert.Equal(t, int64(0), dashboard.OutstandingInvoiceCount)
	assert.Equal(t, int64(0), dashboard.OutstandingInvoiceAmountCents)
	assert.Empty(t, dashboard.RecentProjects)
}
