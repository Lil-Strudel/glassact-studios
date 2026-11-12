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

func seedBackground(t *testing.T, ctx *testContext, code string, related ...*data.CatalogItem) *data.SandblastBackground {
	background := &data.SandblastBackground{
		Code:     code,
		Name:     "Background " + code,
		SvgURL:   "/file/sandblast-backgrounds/" + code + ".svg",
		IsActive: true,
	}
	for _, item := range related {
		background.RelatedItems = append(background.RelatedItems, data.RelatedCatalogItem{ID: item.ID})
	}
	require.NoError(t, ctx.db.SandblastBackgrounds.Insert(background))
	return background
}

func decodeBackgrounds(t *testing.T, body []byte) []data.SandblastBackground {
	var backgrounds []data.SandblastBackground
	require.NoError(t, json.Unmarshal(body, &backgrounds))
	return backgrounds
}

func relatedCodes(background data.SandblastBackground) []string {
	codes := make([]string, len(background.RelatedItems))
	for i, item := range background.RelatedItems {
		codes[i] = item.CatalogCode
	}
	return codes
}

func backgroundWriteBody(code string, relatedIDs ...int) map[string]any {
	if relatedIDs == nil {
		relatedIDs = []int{}
	}
	return map[string]any{
		"code":                     code,
		"name":                     "Heart Frame",
		"description":              nil,
		"tags":                     []string{" Heart ", "heart", "frame"},
		"is_active":                true,
		"svg_url":                  "/file/uploads/heart.svg",
		"content_bbox":             map[string]any{"x": 0, "y": 0, "width": 10, "height": 5},
		"related_catalog_item_ids": relatedIDs,
	}
}

func TestSandblastBackgrounds(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	_, dealershipToken, _, internalToken := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	smallHeart := seedCatalogItem(t, ctx, priceGroup.ID, "E-HRT-0001")
	largeHeart := seedCatalogItem(t, ctx, priceGroup.ID, "E-HRT-0002")
	retiredHeart := seedCatalogItem(t, ctx, priceGroup.ID, "E-HRT-0003")

	shown := seedBackground(t, ctx, "SB-HRT-0001", smallHeart, retiredHeart)
	hidden := seedBackground(t, ctx, "SB-HRT-0002")
	hidden.IsActive = false
	require.NoError(t, ctx.db.SandblastBackgrounds.UpdateFields(hidden))
	retiredHeart.IsActive = false
	require.NoError(t, ctx.db.CatalogItems.Update(retiredHeart))

	t.Run("browse offers active backgrounds with only their active related items", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/background/browse", token: dealershipToken})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		backgrounds := decodeBackgrounds(t, resp.body)
		require.Len(t, backgrounds, 1)
		assert.Equal(t, shown.UUID, backgrounds[0].UUID)
		assert.Equal(t, []string{"E-HRT-0001"}, relatedCodes(backgrounds[0]))
	})

	t.Run("the admin list includes retired backgrounds and related items", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/background", token: internalToken})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		backgrounds := decodeBackgrounds(t, resp.body)
		require.Len(t, backgrounds, 2)
		assert.Equal(t, []string{"E-HRT-0001", "E-HRT-0003"}, relatedCodes(backgrounds[0]))
	})

	t.Run("dealership users cannot manage backgrounds", func(t *testing.T) {
		for _, req := range []testRequest{
			{method: http.MethodGet, path: "/api/background"},
			{method: http.MethodPost, path: "/api/background", body: backgroundWriteBody("SB-X-0001")},
			{method: http.MethodPost, path: "/api/background/prepare", body: map[string]any{"svg_url": "/x.svg"}},
			{method: http.MethodPatch, path: "/api/background/" + shown.UUID, body: map[string]any{"is_active": false, "version": 1}},
		} {
			req.token = dealershipToken
			resp := ctx.request(req)
			assert.Equal(t, http.StatusForbidden, resp.statusCode, "%s %s", req.method, req.path)
		}
	})

	var created data.SandblastBackground
	t.Run("create stores normalized tags and related items in order", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodPost, path: "/api/background", token: internalToken,
			body: backgroundWriteBody("SB-HRT-0010", largeHeart.ID, smallHeart.ID),
		})
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		require.NoError(t, json.Unmarshal(resp.body, &created))
		assert.Equal(t, []string{"heart", "frame"}, created.Tags)
		assert.Equal(t, []string{"E-HRT-0002", "E-HRT-0001"}, relatedCodes(created))
	})

	t.Run("create rejects a code already in use", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodPost, path: "/api/background", token: internalToken,
			body: backgroundWriteBody("sb-hrt-0001"),
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
	})

	t.Run("create rejects retired or repeated related items", func(t *testing.T) {
		for _, ids := range [][]int{{retiredHeart.ID}, {smallHeart.ID, smallHeart.ID}} {
			resp := ctx.request(testRequest{
				method: http.MethodPost, path: "/api/background", token: internalToken,
				body: backgroundWriteBody("SB-HRT-0099", ids...),
			})
			assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
		}
	})

	t.Run("a new svg needs its measured content box", func(t *testing.T) {
		body := backgroundWriteBody("SB-HRT-0098")
		body["content_bbox"] = nil
		resp := ctx.request(testRequest{method: http.MethodPost, path: "/api/background", token: internalToken, body: body})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
	})

	t.Run("update replaces related items and keeps the svg without a content box", func(t *testing.T) {
		body := backgroundWriteBody("SB-HRT-0010", smallHeart.ID)
		body["svg_url"] = created.SvgURL
		body["content_bbox"] = nil
		body["version"] = created.Version
		resp := ctx.request(testRequest{method: http.MethodPut, path: "/api/background/" + created.UUID, token: internalToken, body: body})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		var updated data.SandblastBackground
		require.NoError(t, json.Unmarshal(resp.body, &updated))
		assert.Equal(t, []string{"E-HRT-0001"}, relatedCodes(updated))
		assert.Equal(t, created.Version+1, updated.Version)
	})

	t.Run("update keeps a related item retired since it was linked", func(t *testing.T) {
		body := backgroundWriteBody("SB-HRT-0001", smallHeart.ID, retiredHeart.ID)
		body["svg_url"] = shown.SvgURL
		body["version"] = shown.Version
		resp := ctx.request(testRequest{method: http.MethodPut, path: "/api/background/" + shown.UUID, token: internalToken, body: body})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		var updated data.SandblastBackground
		require.NoError(t, json.Unmarshal(resp.body, &updated))
		assert.Equal(t, []string{"E-HRT-0001", "E-HRT-0003"}, relatedCodes(updated))
	})

	t.Run("update with a stale version conflicts", func(t *testing.T) {
		body := backgroundWriteBody("SB-HRT-0010")
		body["svg_url"] = created.SvgURL
		body["version"] = created.Version
		resp := ctx.request(testRequest{method: http.MethodPut, path: "/api/background/" + created.UUID, token: internalToken, body: body})
		assert.Equal(t, http.StatusConflict, resp.statusCode, string(resp.body))
	})

	t.Run("patch retires a background and keeps its related items", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodPatch, path: "/api/background/" + created.UUID, token: internalToken,
			body: map[string]any{"is_active": false, "version": created.Version + 1},
		})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		reloaded, _, err := ctx.db.SandblastBackgrounds.GetByUUID(created.UUID)
		require.NoError(t, err)
		assert.False(t, reloaded.IsActive)
		assert.Len(t, reloaded.RelatedItems, 1)
	})

	t.Run("sandblast download requires authentication", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: fmt.Sprintf("/api/background/%s/sandblast", shown.UUID)})
		assert.Equal(t, http.StatusUnauthorized, resp.statusCode)
	})

	t.Run("sandblast download of an unknown background is not found", func(t *testing.T) {
		resp := ctx.request(testRequest{
			method: http.MethodGet,
			path:   "/api/background/7ddea090-6db0-4156-95cd-831bcbbf5cc0/sandblast",
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusNotFound, resp.statusCode)
	})

	t.Run("catalog browse no longer takes a kind and lists inlays", func(t *testing.T) {
		resp := ctx.request(testRequest{method: http.MethodGet, path: "/api/catalog/browse", token: dealershipToken})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))
		var page struct {
			Total int `json:"total"`
		}
		require.NoError(t, json.Unmarshal(resp.body, &page))
		assert.Equal(t, 2, page.Total)
	})
}
