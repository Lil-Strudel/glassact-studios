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

func TestPlaceOrder_SnapshotsInstallationKitFromInlay(t *testing.T) {
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

	kitInlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "With Kit")
	plainInlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "No Kit")

	// Toggle the installation kit on via the inlay PATCH endpoint.
	patchResp := ctx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/inlay/%s", kitInlay.UUID),
		token:  dealershipToken,
		body:   map[string]any{"installation_kit": true},
	})
	require.Equal(t, http.StatusOK, patchResp.statusCode, string(patchResp.body))

	reloaded, found, err := ctx.db.Inlays.GetByUUID(kitInlay.UUID)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, reloaded.InstallationKit, "PATCH should persist installation_kit")

	// Place the order for both inlays.
	orderResp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/project/%s/place-order", project.UUID),
		token:  dealershipToken,
		body: map[string]any{
			"inlay_uuids": []string{kitInlay.UUID, plainInlay.UUID},
		},
	})
	require.Equal(t, http.StatusOK, orderResp.statusCode, string(orderResp.body))

	// The kit inlay's snapshot locks in the kit + its price.
	kitSnapshot, found, err := ctx.db.OrderSnapshots.GetByInlayID(kitInlay.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, kitSnapshot.InstallationKit)
	assert.Equal(t, data.InstallationKitPriceCents, kitSnapshot.InstallationKitPriceCents)

	// The plain inlay's snapshot carries no kit.
	plainSnapshot, found, err := ctx.db.OrderSnapshots.GetByInlayID(plainInlay.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.False(t, plainSnapshot.InstallationKit)
	assert.Equal(t, 0, plainSnapshot.InstallationKitPriceCents)
}

func TestPatchInlay_InstallationKit_RejectedWhenNotDraft(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)

	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-KIT-0002")

	project := &data.Project{
		Name:         "Ordered Project",
		Status:       data.ProjectStatuses.Ordered,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Locked")

	resp := ctx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/inlay/%s", inlay.UUID),
		token:  dealershipToken,
		body:   map[string]any{"installation_kit": true},
	})
	assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
}
