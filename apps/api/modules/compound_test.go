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

func compoundBody(name string, itemIDs ...int) map[string]any {
	members := make([]map[string]any, len(itemIDs))
	for i, id := range itemIDs {
		members[i] = map[string]any{"catalog_item_id": id, "center_x": float64(i*4) - 2, "center_y": 0, "rotation_deg": 0}
	}
	return map[string]any{"name": name, "is_active": true, "members": members}
}

func TestCompounds(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, adminToken := seedTestData(t, ctx)
	_, designerToken := seedInternalUser(t, ctx, data.InternalUserRoles.Designer)
	_, productionToken := seedInternalUser(t, ctx, data.InternalUserRoles.Production)
	_, otherToken := seedOtherDealershipToken(t, ctx)

	priceGroup := seedPriceGroup(t, ctx, "Standard")
	left := seedCatalogItem(t, ctx, priceGroup.ID, "A-DOV-0001L")
	right := seedCatalogItem(t, ctx, priceGroup.ID, "A-DOV-0001R")
	background := seedBackgroundItem(t, ctx, priceGroup.ID, "SB-BOR-0001")

	var compound data.CatalogCompound

	t.Run("only catalog managers can create compounds", func(t *testing.T) {
		for _, token := range []string{dealershipToken, productionToken} {
			resp := ctx.request(testRequest{method: http.MethodPost, path: "/api/compound", token: token, body: compoundBody("Doves", left.ID, right.ID)})
			assert.Equal(t, http.StatusForbidden, resp.statusCode)
		}

		resp := ctx.request(testRequest{method: http.MethodPost, path: "/api/compound", token: designerToken, body: compoundBody("Doves", left.ID, right.ID)})
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))
		require.NoError(t, json.Unmarshal(resp.body, &compound))
		require.Len(t, compound.Members, 2)
		assert.Equal(t, "A-DOV-0001L", compound.Members[0].CatalogItem.CatalogCode)
	})

	t.Run("a compound needs at least two members", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodPost, path: "/api/compound", token: adminToken, body: compoundBody("Lonely", left.ID)})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode)
	})

	t.Run("a background cannot be a member", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodPost, path: "/api/compound", token: adminToken, body: compoundBody("Framed", left.ID, background.ID)})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode)
	})

	t.Run("dealerships can browse offerable compounds", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/compound/browse", token: dealershipToken})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))
		var compounds []data.CatalogCompound
		require.NoError(t, json.Unmarshal(resp.body, &compounds))
		require.Len(t, compounds, 1)
		assert.Equal(t, compound.UUID, compounds[0].UUID)
	})

	t.Run("adding a compound creates one inlay per member, grouped on the stone", func(t *testing.T) {
		project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Smith")

		resp := ctx.request(testRequest{
			method: http.MethodPost, path: fmt.Sprintf("/api/project/%s/inlays/compound", project.UUID),
			token: dealershipToken, body: map[string]any{"compound_uuid": compound.UUID},
		})
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		inlays, err := ctx.db.Inlays.GetByProjectID(project.ID)
		require.NoError(t, err)
		require.Len(t, inlays, 2)
		for _, inlay := range inlays {
			assert.Equal(t, data.InlayTypes.Catalog, inlay.Type)
			assert.False(t, inlay.IsCustomized)
		}

		design, found, err := ctx.db.StoneDesigns.GetByProjectID(project.ID)
		require.NoError(t, err)
		require.True(t, found, "adding a compound should create the stone design")
		require.Len(t, design.Layout.Groups, 1)
		require.Len(t, design.Layout.Elements, 2)
		group := design.Layout.Groups[0]
		assert.Equal(t, compound.UUID, *group.CompoundUUID)
		assert.Equal(t, "Doves", group.Name)
		for _, el := range design.Layout.Elements {
			assert.Equal(t, group.ID, *el.GroupID)
		}
		assert.InDelta(t, design.StoneWidth/2-2, design.Layout.Elements[0].X, 0.001)
		assert.InDelta(t, design.StoneWidth/2+2, design.Layout.Elements[1].X, 0.001)
	})

	t.Run("adding to an ordered project is rejected", func(t *testing.T) {
		project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Ordered")
		project.Status = data.ProjectStatuses.Ordered
		require.NoError(t, ctx.db.Projects.Update(project))

		resp := ctx.request(testRequest{
			method: http.MethodPost, path: fmt.Sprintf("/api/project/%s/inlays/compound", project.UUID),
			token: dealershipToken, body: map[string]any{"compound_uuid": compound.UUID},
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode)
	})

	t.Run("another dealership cannot add to the project", func(t *testing.T) {
		project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Mine")
		resp := ctx.request(testRequest{
			method: http.MethodPost, path: fmt.Sprintf("/api/project/%s/inlays/compound", project.UUID),
			token: otherToken, body: map[string]any{"compound_uuid": compound.UUID},
		})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
	})

	t.Run("a compound with a retired member is hidden and cannot be added", func(t *testing.T) {
		right.IsActive = false
		require.NoError(t, ctx.db.CatalogItems.Update(right))

		resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/compound/browse", token: dealershipToken})
		require.Equal(t, http.StatusOK, resp.statusCode)
		var compounds []data.CatalogCompound
		require.NoError(t, json.Unmarshal(resp.body, &compounds))
		assert.Empty(t, compounds)

		project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Retired")
		resp = ctx.request(testRequest{
			method: http.MethodPost, path: fmt.Sprintf("/api/project/%s/inlays/compound", project.UUID),
			token: dealershipToken, body: map[string]any{"compound_uuid": compound.UUID},
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode)
		inlays, err := ctx.db.Inlays.GetByProjectID(project.ID)
		require.NoError(t, err)
		assert.Empty(t, inlays)
	})

	t.Run("update with a stale version conflicts", func(t *testing.T) {
		right.IsActive = true
		require.NoError(t, ctx.db.CatalogItems.Update(right))

		body := compoundBody("Doves renamed", left.ID, right.ID)
		body["version"] = compound.Version
		resp := ctx.request(testRequest{method: http.MethodPut, path: "/api/compound/" + compound.UUID, token: adminToken, body: body})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		resp = ctx.request(testRequest{method: http.MethodPut, path: "/api/compound/" + compound.UUID, token: adminToken, body: body})
		assert.Equal(t, http.StatusConflict, resp.statusCode, string(resp.body))
	})
}
