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

func seedInternalUser(t *testing.T, ctx *testContext, role data.InternalUserRole) (*data.InternalUser, string) {
	t.Helper()

	user := &data.InternalUser{
		Name:     "Internal " + string(role),
		Email:    fmt.Sprintf("%s%d@example.com", role, time.Now().UnixNano()),
		Avatar:   "https://example.com/avatar.jpg",
		Role:     role,
		IsActive: true,
	}
	require.NoError(t, ctx.db.InternalUsers.Insert(user))

	token, err := ctx.db.InternalTokens.New(user.ID, 2*time.Hour, data.InternalScopeAccess)
	require.NoError(t, err)

	return user, token.Plaintext
}

func TestSupportModule(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	_, dealershipToken, _, internalAdminToken := seedTestData(t, testCtx)
	_, designerToken := seedInternalUser(t, testCtx, data.InternalUserRoles.Designer)

	articleBody := map[string]interface{}{
		"category":     "installation",
		"title":        "How to install",
		"body":         "Some **markdown**.",
		"youtube_url":  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"sort_order":   10,
		"is_published": true,
	}

	t.Run("admin can create an article", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "POST",
			path:   "/api/support/articles",
			body:   articleBody,
			token:  internalAdminToken,
		})
		assert.Equal(t, http.StatusCreated, resp.statusCode)
	})

	t.Run("dealership user is forbidden from creating", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "POST",
			path:   "/api/support/articles",
			body:   articleBody,
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
	})

	t.Run("non-admin internal user is forbidden from creating", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "POST",
			path:   "/api/support/articles",
			body:   articleBody,
			token:  designerToken,
		})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
	})

	t.Run("invalid category is rejected", func(t *testing.T) {
		bad := map[string]interface{}{
			"category": "bogus",
			"title":    "Bad",
		}
		resp := testCtx.request(testRequest{
			method: "POST",
			path:   "/api/support/articles",
			body:   bad,
			token:  internalAdminToken,
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode)
	})

	t.Run("dealership user can read published articles", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "GET",
			path:   "/api/support/articles",
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusOK, resp.statusCode)

		var articles []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.body, &articles))
		assert.GreaterOrEqual(t, len(articles), 1)
	})

	t.Run("unauthenticated read is rejected", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "GET",
			path:   "/api/support/articles",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.statusCode)
	})

	t.Run("admin can update and dealership cannot", func(t *testing.T) {
		created := testCtx.request(testRequest{
			method: "POST",
			path:   "/api/support/articles",
			body: map[string]interface{}{
				"category":     "ordering",
				"title":        "Editable",
				"body":         "",
				"is_published": true,
			},
			token: internalAdminToken,
		})
		require.Equal(t, http.StatusCreated, created.statusCode)

		var article map[string]interface{}
		require.NoError(t, json.Unmarshal(created.body, &article))
		uuid, ok := article["uuid"].(string)
		require.True(t, ok)

		patch := map[string]interface{}{"title": "Edited title"}

		denied := testCtx.request(testRequest{
			method: "PATCH",
			path:   "/api/support/articles/" + uuid,
			body:   patch,
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusForbidden, denied.statusCode)

		allowed := testCtx.request(testRequest{
			method: "PATCH",
			path:   "/api/support/articles/" + uuid,
			body:   patch,
			token:  internalAdminToken,
		})
		assert.Equal(t, http.StatusOK, allowed.statusCode)
	})

	t.Run("price groups are readable by dealership users", func(t *testing.T) {
		seedPriceGroup(t, testCtx, "Support PG")

		resp := testCtx.request(testRequest{
			method: "GET",
			path:   "/api/support/price-groups",
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusOK, resp.statusCode)

		var priceGroups []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.body, &priceGroups))
		assert.GreaterOrEqual(t, len(priceGroups), 1)
	})
}
