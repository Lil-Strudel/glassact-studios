package modules

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedOrderedProjectWithInlay creates an ordered project owned by the given
// dealership plus a stock catalog inlay on it, ready for a sandblast file.
func seedOrderedProjectWithInlay(t *testing.T, ctx *testContext, dealershipID, catalogItemID int) (*data.Project, *data.Inlay) {
	project := &data.Project{
		Name:         "Ordered Project",
		Status:       data.ProjectStatuses.Ordered,
		DealershipID: dealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	inlay := seedDraftCatalogInlay(t, ctx, project.ID, catalogItemID, "Dove")
	return project, inlay
}

func TestPostSandblastFile_InternalUser_Persists(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, _, _, internalToken := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-SB-0001")

	_, inlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/sandblast", inlay.UUID),
		token:  internalToken,
		body:   map[string]any{"sandblast_file_url": "/file/sandblast/abc123.svg"},
	})
	require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

	reloaded, found, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, reloaded.SandblastFileURL)
	assert.Equal(t, "/file/sandblast/abc123.svg", *reloaded.SandblastFileURL)
}

func TestPostSandblastFile_RejectedWhenDraft(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, _, _, internalToken := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-SB-0002")

	project := &data.Project{
		Name:         "Draft Project",
		Status:       data.ProjectStatuses.Draft,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))
	inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Dove")

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/sandblast", inlay.UUID),
		token:  internalToken,
		body:   map[string]any{"sandblast_file_url": "/file/sandblast/abc123.svg"},
	})
	assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
}

func TestPostSandblastFile_DealershipUserForbidden(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-SB-0003")

	_, inlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/sandblast", inlay.UUID),
		token:  dealershipToken,
		body:   map[string]any{"sandblast_file_url": "/file/sandblast/abc123.svg"},
	})
	assert.Equal(t, http.StatusForbidden, resp.statusCode, string(resp.body))
}

func TestGetSandblastFile_NotFoundWhenEmpty(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-SB-0004")

	_, inlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

	resp := ctx.request(testRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/inlay/%s/sandblast", inlay.UUID),
		token:  dealershipToken,
	})
	assert.Equal(t, http.StatusNotFound, resp.statusCode, string(resp.body))
}

func TestGetSandblastFile_OtherDealershipForbidden(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, _, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-SB-0005")

	_, inlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

	// Attach a sandblast file so the ownership check is what blocks access,
	// not the empty-file 404.
	url := "/file/sandblast/secret.svg"
	inlay.SandblastFileURL = &url
	require.NoError(t, ctx.db.Inlays.UpdateSandblastFile(inlay))

	// Create a second dealership + user in a different tenant.
	otherDealership := &data.Dealership{
		Name: "Other Dealership",
		Address: data.Address{
			Street: "456 Side St", City: "Other City", State: "OS",
			PostalCode: "67890", Country: "US", Latitude: 41.0, Longitude: -73.0,
		},
	}
	require.NoError(t, ctx.db.Dealerships.Insert(otherDealership))

	otherUser := &data.DealershipUser{
		DealershipID: otherDealership.ID,
		Name:         "Other User",
		Email:        fmt.Sprintf("other%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.jpg",
		Role:         data.DealershipUserRoles.Admin,
		IsActive:     true,
	}
	require.NoError(t, ctx.db.DealershipUsers.Insert(otherUser))
	otherToken, err := ctx.db.DealershipTokens.New(otherUser.ID, 2*time.Hour, data.DealershipScopeAccess)
	require.NoError(t, err)

	resp := ctx.request(testRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/inlay/%s/sandblast", inlay.UUID),
		token:  otherToken.Plaintext,
	})
	assert.Equal(t, http.StatusForbidden, resp.statusCode, string(resp.body))
}
