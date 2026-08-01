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

// Deactivating a user is a soft delete, so the PATCH route has to be able to put
// them back — otherwise an accidental deactivation is unrecoverable from the UI.
func TestUserDeactivateReactivateRoundTrip(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, internalAdminToken := seedTestData(t, testCtx)

	t.Run("dealership user can be deactivated then reactivated", func(t *testing.T) {
		target := &data.DealershipUser{
			DealershipID: dealershipUser.DealershipID,
			Name:         "Reactivate Me",
			Email:        fmt.Sprintf("reactivate%d@example.com", time.Now().UnixNano()),
			Avatar:       "https://example.com/avatar.jpg",
			Role:         data.DealershipUserRoles.Submitter,
			IsActive:     true,
		}
		require.NoError(t, testCtx.db.DealershipUsers.Insert(target))

		deleted := testCtx.request(testRequest{
			method: "DELETE",
			path:   "/api/dealership-user/" + target.UUID,
			token:  dealershipToken,
		})
		require.Equal(t, http.StatusOK, deleted.statusCode)

		stored, found, err := testCtx.db.DealershipUsers.GetByUUID(target.UUID)
		require.NoError(t, err)
		require.True(t, found)
		require.False(t, stored.IsActive)

		reactivated := testCtx.request(testRequest{
			method: "PATCH",
			path:   "/api/dealership-user/" + target.UUID,
			body:   map[string]interface{}{"is_active": true},
			token:  dealershipToken,
		})
		require.Equal(t, http.StatusOK, reactivated.statusCode)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(reactivated.body, &body))
		assert.Equal(t, true, body["is_active"])

		stored, found, err = testCtx.db.DealershipUsers.GetByUUID(target.UUID)
		require.NoError(t, err)
		require.True(t, found)
		assert.True(t, stored.IsActive)
	})

	t.Run("internal user can be deactivated then reactivated", func(t *testing.T) {
		target, _ := seedInternalUser(t, testCtx, data.InternalUserRoles.Production)

		deleted := testCtx.request(testRequest{
			method: "DELETE",
			path:   "/api/internal-user/" + target.UUID,
			token:  internalAdminToken,
		})
		require.Equal(t, http.StatusOK, deleted.statusCode)

		stored, found, err := testCtx.db.InternalUsers.GetByUUID(target.UUID)
		require.NoError(t, err)
		require.True(t, found)
		require.False(t, stored.IsActive)

		reactivated := testCtx.request(testRequest{
			method: "PATCH",
			path:   "/api/internal-user/" + target.UUID,
			body:   map[string]interface{}{"is_active": true},
			token:  internalAdminToken,
		})
		require.Equal(t, http.StatusOK, reactivated.statusCode)

		stored, found, err = testCtx.db.InternalUsers.GetByUUID(target.UUID)
		require.NoError(t, err)
		require.True(t, found)
		assert.True(t, stored.IsActive)
	})

	// A bool without an explicit false in the payload must not be clobbered by the
	// zero value the way the string fields on this handler are guarded by != "".
	t.Run("omitting is_active leaves it untouched", func(t *testing.T) {
		target, _ := seedInternalUser(t, testCtx, data.InternalUserRoles.Billing)

		resp := testCtx.request(testRequest{
			method: "PATCH",
			path:   "/api/internal-user/" + target.UUID,
			body:   map[string]interface{}{"name": "Renamed Only"},
			token:  internalAdminToken,
		})
		require.Equal(t, http.StatusOK, resp.statusCode)

		stored, found, err := testCtx.db.InternalUsers.GetByUUID(target.UUID)
		require.NoError(t, err)
		require.True(t, found)
		assert.True(t, stored.IsActive)
		assert.Equal(t, "Renamed Only", stored.Name)
	})

	t.Run("is_active false deactivates via PATCH", func(t *testing.T) {
		target, _ := seedInternalUser(t, testCtx, data.InternalUserRoles.Designer)

		resp := testCtx.request(testRequest{
			method: "PATCH",
			path:   "/api/internal-user/" + target.UUID,
			body:   map[string]interface{}{"is_active": false},
			token:  internalAdminToken,
		})
		require.Equal(t, http.StatusOK, resp.statusCode)

		stored, found, err := testCtx.db.InternalUsers.GetByUUID(target.UUID)
		require.NoError(t, err)
		require.True(t, found)
		assert.False(t, stored.IsActive)
	})
}
