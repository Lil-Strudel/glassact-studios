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

func seedBackgroundItem(t *testing.T, ctx *testContext, priceGroupID int, catalogCode string) *data.CatalogItem {
	item := &data.CatalogItem{
		CatalogCode:         catalogCode,
		Name:                "Background " + catalogCode,
		Category:            data.BackgroundCategory,
		DefaultWidth:        4,
		DefaultHeight:       2,
		MinWidth:            4,
		MinHeight:           2,
		DefaultPriceGroupID: priceGroupID,
		SvgURL:              "/file/catalog-items/" + catalogCode + ".svg",
		IsActive:            true,
	}
	require.NoError(t, ctx.db.CatalogItems.Insert(item))
	return item
}

func browseCodes(t *testing.T, ctx *testContext, token, query string) []string {
	resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/catalog/browse" + query, token: token})
	require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

	var page struct {
		Items []data.CatalogItem `json:"items"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &page))

	codes := make([]string, len(page.Items))
	for i, item := range page.Items {
		codes[i] = item.CatalogCode
	}
	return codes
}

func TestCatalogBackgrounds(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	inlayItem := seedCatalogItem(t, ctx, priceGroup.ID, "A-DOV-0001")
	background := seedBackgroundItem(t, ctx, priceGroup.ID, "SB-BOR-0001")

	t.Run("browse defaults to inlays only", func(t *testing.T) {
		assert.Equal(t, []string{inlayItem.CatalogCode}, browseCodes(t, ctx, dealershipToken, ""))
	})

	t.Run("browse kind=background returns only backgrounds", func(t *testing.T) {
		assert.Equal(t, []string{background.CatalogCode}, browseCodes(t, ctx, dealershipToken, "?kind=background"))
	})

	t.Run("browse kind=all returns both", func(t *testing.T) {
		assert.ElementsMatch(t, []string{inlayItem.CatalogCode, background.CatalogCode}, browseCodes(t, ctx, dealershipToken, "?kind=all"))
	})

	t.Run("browse rejects an unknown kind", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/catalog/browse?kind=stone", token: dealershipToken})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode)
	})

	t.Run("a background cannot be added to a project as an inlay", func(t *testing.T) {
		project := &data.Project{Name: "Draft", Status: data.ProjectStatuses.Draft, DealershipID: dealershipUser.DealershipID}
		require.NoError(t, ctx.db.Projects.Insert(project))

		resp := ctx.request(testRequest{
			method: http.MethodPost,
			path:   fmt.Sprintf("/api/project/%s/inlays/catalog", project.UUID),
			token:  dealershipToken,
			body:   map[string]any{"name": "Border", "catalog_item_id": background.ID},
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))

		inlays, err := ctx.db.Inlays.GetByProjectID(project.ID)
		require.NoError(t, err)
		assert.Empty(t, inlays)
	})

	t.Run("sandblast download is not offered for inlay items", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/catalog/%s/sandblast", inlayItem.UUID),
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusNotFound, resp.statusCode)
	})

	t.Run("sandblast download requires authentication", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/catalog/%s/sandblast", background.UUID),
		})
		assert.Equal(t, http.StatusUnauthorized, resp.statusCode)
	})
}
