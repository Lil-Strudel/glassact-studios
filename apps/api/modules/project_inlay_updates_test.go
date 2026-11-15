package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectInlayUpdates_DealershipUser_ReturnsUpdatesForEveryInlayOnTheProject(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-UPD-0001")

	project, dove := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)
	rose := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Rose")
	_, otherProjectInlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

	for _, update := range []*data.InlayUpdate{
		{InlayID: dove.ID, UpdateType: data.InlayUpdateTypes.Issue, Message: "Cracked during manufacturing"},
		{InlayID: rose.ID, UpdateType: data.InlayUpdateTypes.Info, Message: "Ahead of schedule"},
		{InlayID: otherProjectInlay.ID, UpdateType: data.InlayUpdateTypes.Issue, Message: "Belongs to another project"},
	} {
		require.NoError(t, ctx.db.InlayUpdates.Insert(update))
	}

	resp := ctx.request(testRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/project/%s/inlay-updates", project.UUID),
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

	var updates []data.InlayUpdate
	require.NoError(t, json.Unmarshal(resp.body, &updates))

	messages := make([]string, len(updates))
	for i, update := range updates {
		messages[i] = update.Message
	}
	assert.ElementsMatch(t, []string{"Cracked during manufacturing", "Ahead of schedule"}, messages)
}

func TestGetProjectInlayUpdates_OtherDealership_Forbidden(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, _, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-UPD-0002")

	project, inlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)
	require.NoError(t, ctx.db.InlayUpdates.Insert(&data.InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: data.InlayUpdateTypes.Issue,
		Message:    "Private to the owning dealership",
	}))

	_, otherToken := seedOtherDealershipToken(t, ctx)

	resp := ctx.request(testRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/project/%s/inlay-updates", project.UUID),
		token:  otherToken,
	})
	assert.Equal(t, http.StatusForbidden, resp.statusCode, string(resp.body))
}
