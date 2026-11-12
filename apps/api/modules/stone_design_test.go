package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedOtherDealershipToken(t *testing.T, ctx *testContext) (*data.Dealership, string) {
	t.Helper()

	dealership := &data.Dealership{
		Name:                "Other Dealership",
		PaymentTiming:       data.PaymentTimings.PostShipping,
		SandblastFileFormat: data.SandblastFileFormats.PDF,
		Address: data.Address{
			Street: "456 Side St", City: "Other City", State: "OS",
			PostalCode: "67890", Country: "US", Latitude: 41.0, Longitude: -73.0,
		},
	}
	require.NoError(t, ctx.db.Dealerships.Insert(dealership))

	user := seedDealershipUser(t, ctx, dealership.ID, data.DealershipUserRoles.Admin)
	token, err := ctx.db.DealershipTokens.New(user.ID, 2*time.Hour, data.DealershipScopeAccess)
	require.NoError(t, err)
	return dealership, token.Plaintext
}

func stoneDesignBody(version int, elements ...map[string]any) map[string]any {
	if elements == nil {
		elements = []map[string]any{}
	}
	return map[string]any{
		"version":      version,
		"stone_shape":  "serpentine",
		"stone_width":  36,
		"stone_height": 24,
		"granite_key":  "georgia-gray",
		"layout":       map[string]any{"elements": elements, "groups": []any{}},
	}
}

func inlayElement(id, inlayUUID string) map[string]any {
	return map[string]any{"id": id, "type": "inlay", "x": 10, "y": 10, "rotation": 0, "group_id": nil, "inlay_uuid": inlayUUID}
}

func decodeDesign(t *testing.T, body []byte) data.StoneDesign {
	t.Helper()
	var design data.StoneDesign
	require.NoError(t, json.Unmarshal(body, &design))
	return design
}

func TestStoneDesign(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, internalToken := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-DOV-0001")
	background := seedBackground(t, ctx, "SB-BOR-0001")

	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Smith")
	inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Dove")
	path := fmt.Sprintf("/api/project/%s/stone-design", project.UUID)

	_, otherToken := seedOtherDealershipToken(t, ctx)
	otherProject := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Jones")
	otherInlay := seedDraftCatalogInlay(t, ctx, otherProject.ID, item.ID, "Elsewhere")

	t.Run("no design yet returns 404", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: path, token: dealershipToken})
		assert.Equal(t, http.StatusNotFound, resp.statusCode)
	})

	t.Run("first save creates the design and drops inlays from other projects", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodPut, path: path, token: dealershipToken,
			body: stoneDesignBody(0,
				inlayElement("e1", inlay.UUID),
				inlayElement("e2", otherInlay.UUID),
				map[string]any{"id": "e3", "type": "background", "x": 18, "y": 12, "rotation": 0, "group_id": nil, "background_uuid": background.UUID, "width": 30},
				map[string]any{"id": "e4", "type": "text", "x": 18, "y": 4, "rotation": 0, "group_id": nil, "text": "SMITH", "font": "roman", "size": 2, "align": "center", "finish": "engraved"},
			),
		})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		design := decodeDesign(t, resp.body)
		assert.Equal(t, 1, design.Version)
		require.Len(t, design.Layout.Elements, 3)
		for _, el := range design.Layout.Elements {
			assert.NotEqual(t, otherInlay.UUID, el.InlayUUID)
		}
	})

	t.Run("save with a stale version conflicts", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodPut, path: path, token: dealershipToken, body: stoneDesignBody(0)})
		assert.Equal(t, http.StatusConflict, resp.statusCode, string(resp.body))
	})

	t.Run("save with the current version updates", func(t *testing.T) {
		body := stoneDesignBody(1, inlayElement("e1", inlay.UUID))
		body["stone_shape"] = "heart"
		resp := ctx.request(testRequest{method: http.MethodPut, path: path, token: dealershipToken, body: body})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))
		design := decodeDesign(t, resp.body)
		assert.Equal(t, data.StoneShapes.Heart, design.StoneShape)
		assert.Equal(t, 2, design.Version)
	})

	t.Run("a catalog item is rejected as a background", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodPut, path: path, token: dealershipToken,
			body: stoneDesignBody(2, map[string]any{"id": "b", "type": "background", "x": 1, "y": 1, "rotation": 0, "group_id": nil, "background_uuid": item.UUID, "width": 10}),
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
	})

	t.Run("an unknown stone shape is rejected", func(t *testing.T) {
		body := stoneDesignBody(2)
		body["stone_shape"] = "pyramid"
		resp := ctx.request(testRequest{method: http.MethodPut, path: path, token: dealershipToken, body: body})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
	})

	t.Run("another dealership can neither read nor write it", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: path, token: otherToken})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
		resp = ctx.request(testRequest{method: http.MethodPut, path: path, token: otherToken, body: stoneDesignBody(2)})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
	})

	t.Run("internal staff can read but not write", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: path, token: internalToken})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))
		resp = ctx.request(testRequest{method: http.MethodPut, path: path, token: internalToken, body: stoneDesignBody(2)})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
	})

	t.Run("a deleted inlay disappears from the design on read", func(t *testing.T) {
		require.NoError(t, ctx.db.Inlays.Delete(inlay.ID))
		resp := ctx.request(testRequest{method: http.MethodGet, path: path, token: dealershipToken})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))
		assert.Empty(t, decodeDesign(t, resp.body).Layout.Elements)
	})

	t.Run("the design is frozen once the project is ordered", func(t *testing.T) {
		project.Status = data.ProjectStatuses.Ordered
		require.NoError(t, ctx.db.Projects.Update(project))
		resp := ctx.request(testRequest{method: http.MethodPut, path: path, token: dealershipToken, body: stoneDesignBody(2)})
		assert.Equal(t, http.StatusConflict, resp.statusCode, string(resp.body))
	})
}
