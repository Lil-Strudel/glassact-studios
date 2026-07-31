package modules

import (
	"fmt"
	"net/http"
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedDraftCatalogInlay creates a stock (non-customized) catalog inlay on the
// given project. Stock catalog inlays are ready to order immediately.
func seedDraftCatalogInlay(t *testing.T, ctx *testContext, projectID, catalogItemID int, name string) *data.Inlay {
	inlay := &data.Inlay{
		ProjectID: projectID,
		Name:      name,
		Type:      data.InlayTypes.Catalog,
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID:      catalogItemID,
			CustomizationNotes: "",
		},
	}
	err := ctx.db.Inlays.Insert(inlay)
	require.NoError(t, err)
	return inlay
}

// One kit covers the whole project, so the charge is flat no matter how many
// inlays are on the order.
func TestPlaceOrder_LocksSingleInstallationKitPriceOnProject(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)

	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-KIT-0001")

	project := &data.Project{
		Name:         "Kit Project",
		Status:       data.ProjectStatuses.Draft,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	firstInlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "First")
	secondInlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Second")

	patchResp := ctx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/project/%s", project.UUID),
		token:  dealershipToken,
		body:   map[string]any{"installation_kit": true},
	})
	require.Equal(t, http.StatusOK, patchResp.statusCode, string(patchResp.body))

	reloaded, found, err := ctx.db.Projects.GetByUUID(project.UUID)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, reloaded.InstallationKit, "PATCH should persist installation_kit")
	assert.Nil(t, reloaded.InstallationKitPriceCents, "price should not lock before the order is placed")

	orderResp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/project/%s/place-order", project.UUID),
		token:  dealershipToken,
		body: map[string]any{
			"inlay_uuids": []string{firstInlay.UUID, secondInlay.UUID},
		},
	})
	require.Equal(t, http.StatusOK, orderResp.statusCode, string(orderResp.body))

	ordered, found, err := ctx.db.Projects.GetByUUID(project.UUID)
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, ordered.InstallationKitPriceCents)
	assert.Equal(t, data.InstallationKitPriceCents, *ordered.InstallationKitPriceCents,
		"two inlays must still be charged for exactly one kit")
}

func TestPlaceOrder_WithoutInstallationKit_LocksZero(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)

	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-KIT-0003")

	project := &data.Project{
		Name:         "No Kit Project",
		Status:       data.ProjectStatuses.Draft,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Plain")

	orderResp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/project/%s/place-order", project.UUID),
		token:  dealershipToken,
		body:   map[string]any{"inlay_uuids": []string{inlay.UUID}},
	})
	require.Equal(t, http.StatusOK, orderResp.statusCode, string(orderResp.body))

	ordered, found, err := ctx.db.Projects.GetByUUID(project.UUID)
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, ordered.InstallationKitPriceCents)
	assert.Equal(t, 0, *ordered.InstallationKitPriceCents)
}

func TestPatchProject_InstallationKit_RejectedWhenNotDraft(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)

	project := &data.Project{
		Name:         "Ordered Project",
		Status:       data.ProjectStatuses.Ordered,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	resp := ctx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/project/%s", project.UUID),
		token:  dealershipToken,
		body:   map[string]any{"installation_kit": true},
	})
	assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
}
