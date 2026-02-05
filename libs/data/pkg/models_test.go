package data

import (
	"testing"
)

func TestNewModels(t *testing.T) {
	models := getTestModels(t)

	// Verify all model fields are initialized
	if models.CatalogItems.DB == nil {
		t.Errorf("Expected CatalogItems.DB to be initialized")
	}
	if models.DealershipAccounts.DB == nil {
		t.Errorf("Expected DealershipAccounts.DB to be initialized")
	}
	if models.DealershipTokens.DB == nil {
		t.Errorf("Expected DealershipTokens.DB to be initialized")
	}
	if models.DealershipUsers.DB == nil {
		t.Errorf("Expected DealershipUsers.DB to be initialized")
	}
	if models.Dealerships.DB == nil {
		t.Errorf("Expected Dealerships.DB to be initialized")
	}
	if models.InlayBlockers.DB == nil {
		t.Errorf("Expected InlayBlockers.DB to be initialized")
	}
	if models.InlayChats.DB == nil {
		t.Errorf("Expected InlayChats.DB to be initialized")
	}
	if models.InlayMilestones.DB == nil {
		t.Errorf("Expected InlayMilestones.DB to be initialized")
	}
	if models.InlayProofs.DB == nil {
		t.Errorf("Expected InlayProofs.DB to be initialized")
	}
	if models.Inlays.DB == nil {
		t.Errorf("Expected Inlays.DB to be initialized")
	}
	if models.InternalAccounts.DB == nil {
		t.Errorf("Expected InternalAccounts.DB to be initialized")
	}
	if models.InternalTokens.DB == nil {
		t.Errorf("Expected InternalTokens.DB to be initialized")
	}
	if models.InternalUsers.DB == nil {
		t.Errorf("Expected InternalUsers.DB to be initialized")
	}
	if models.Invoices.DB == nil {
		t.Errorf("Expected Invoices.DB to be initialized")
	}
	if models.Notifications.DB == nil {
		t.Errorf("Expected Notifications.DB to be initialized")
	}
	if models.OrderSnapshots.DB == nil {
		t.Errorf("Expected OrderSnapshots.DB to be initialized")
	}
	if models.PriceGroups.DB == nil {
		t.Errorf("Expected PriceGroups.DB to be initialized")
	}
	if models.ProjectChats.DB == nil {
		t.Errorf("Expected ProjectChats.DB to be initialized")
	}
	if models.Projects.DB == nil {
		t.Errorf("Expected Projects.DB to be initialized")
	}
	if models.Pool == nil {
		t.Errorf("Expected Pool to be initialized")
	}
	if models.STDB == nil {
		t.Errorf("Expected STDB to be initialized")
	}
}

func TestModelsCatalogItems(t *testing.T) {
	models := getTestModels(t)

	if models.CatalogItems.DB != models.Pool {
		t.Errorf("Expected CatalogItems.DB to be same as Pool")
	}
	if models.CatalogItems.STDB != models.STDB {
		t.Errorf("Expected CatalogItems.STDB to be same as STDB")
	}
}

func TestModelsDealerships(t *testing.T) {
	models := getTestModels(t)

	if models.Dealerships.DB != models.Pool {
		t.Errorf("Expected Dealerships.DB to be same as Pool")
	}
	if models.Dealerships.STDB != models.STDB {
		t.Errorf("Expected Dealerships.STDB to be same as STDB")
	}
}

func TestModelsInlays(t *testing.T) {
	models := getTestModels(t)

	if models.Inlays.DB != models.Pool {
		t.Errorf("Expected Inlays.DB to be same as Pool")
	}
	if models.Inlays.STDB != models.STDB {
		t.Errorf("Expected Inlays.STDB to be same as STDB")
	}
}

func TestModelsProjects(t *testing.T) {
	models := getTestModels(t)

	if models.Projects.DB != models.Pool {
		t.Errorf("Expected Projects.DB to be same as Pool")
	}
	if models.Projects.STDB != models.STDB {
		t.Errorf("Expected Projects.STDB to be same as STDB")
	}
}

func TestModelsInternalUsers(t *testing.T) {
	models := getTestModels(t)

	if models.InternalUsers.DB != models.Pool {
		t.Errorf("Expected InternalUsers.DB to be same as Pool")
	}
	if models.InternalUsers.STDB != models.STDB {
		t.Errorf("Expected InternalUsers.STDB to be same as STDB")
	}
}

func TestModelsInternalAccounts(t *testing.T) {
	models := getTestModels(t)

	if models.InternalAccounts.DB != models.Pool {
		t.Errorf("Expected InternalAccounts.DB to be same as Pool")
	}
	if models.InternalAccounts.STDB != models.STDB {
		t.Errorf("Expected InternalAccounts.STDB to be same as STDB")
	}
}

func TestModelsInternalTokens(t *testing.T) {
	models := getTestModels(t)

	if models.InternalTokens.DB != models.Pool {
		t.Errorf("Expected InternalTokens.DB to be same as Pool")
	}
	if models.InternalTokens.STDB != models.STDB {
		t.Errorf("Expected InternalTokens.STDB to be same as STDB")
	}
}
