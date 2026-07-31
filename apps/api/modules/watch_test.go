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

// seedDealershipUser adds another user to an existing dealership so tests can
// distinguish "watching" from "merely belongs to the same dealership".
func seedDealershipUser(t *testing.T, ctx *testContext, dealershipID int, role data.DealershipUserRole) *data.DealershipUser {
	t.Helper()

	user := &data.DealershipUser{
		DealershipID: dealershipID,
		Name:         fmt.Sprintf("Dealership %s", role),
		Email:        fmt.Sprintf("dealer-%s-%d@example.com", role, time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.jpg",
		Role:         role,
		IsActive:     true,
	}
	require.NoError(t, ctx.db.DealershipUsers.Insert(user))
	return user
}

func notificationEventTypesFor(t *testing.T, ctx *testContext, userID int, isDealership bool) []data.NotificationEventType {
	t.Helper()

	var notifs []*data.Notification
	var err error
	if isDealership {
		notifs, err = ctx.db.Notifications.GetForDealershipUser(userID)
	} else {
		notifs, err = ctx.db.Notifications.GetForInternalUser(userID)
	}
	require.NoError(t, err)

	eventTypes := make([]data.NotificationEventType, len(notifs))
	for i, notif := range notifs {
		eventTypes[i] = notif.EventType
	}
	return eventTypes
}

func TestPutProjectWatch_TogglesAndPersists(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, testCtx)

	project := &data.Project{
		DealershipID: dealershipUser.DealershipID,
		Name:         "Watch Toggle Project",
		Status:       data.ProjectStatuses.Draft,
	}
	require.NoError(t, testCtx.db.Projects.Insert(project))

	// Creating the project through the API auto-subscribes, but this project was
	// inserted directly, so the user starts unsubscribed.
	isWatching, err := testCtx.db.ProjectWatchers.IsWatching(project.ID, dealershipUser)
	require.NoError(t, err)
	assert.False(t, isWatching)

	res := testCtx.request(testRequest{
		method: http.MethodPut,
		path:   fmt.Sprintf("/api/project/%s/watch", project.UUID),
		body:   map[string]any{"is_watching": true},
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, res.statusCode, string(res.body))

	var watchState struct {
		IsWatching   bool `json:"is_watching"`
		WatcherCount int  `json:"watcher_count"`
	}
	require.NoError(t, json.Unmarshal(res.body, &watchState))
	assert.True(t, watchState.IsWatching)
	assert.Equal(t, 1, watchState.WatcherCount)

	res = testCtx.request(testRequest{
		method: http.MethodPut,
		path:   fmt.Sprintf("/api/project/%s/watch", project.UUID),
		body:   map[string]any{"is_watching": false},
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, res.statusCode, string(res.body))
	require.NoError(t, json.Unmarshal(res.body, &watchState))
	assert.False(t, watchState.IsWatching)
	assert.Equal(t, 0, watchState.WatcherCount)

	isWatching, err = testCtx.db.ProjectWatchers.IsWatching(project.ID, dealershipUser)
	require.NoError(t, err)
	assert.False(t, isWatching)
}

func TestProjectWatch_DealershipUserCannotWatchOtherDealershipsProject(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	_, dealershipToken, _, _ := seedTestData(t, testCtx)

	otherDealership := &data.Dealership{
		Name: "Other Dealership",
		Address: data.Address{
			Street:     "999 Other St",
			City:       "Other City",
			State:      "OS",
			PostalCode: "99999",
			Country:    "US",
			Latitude:   41.0,
			Longitude:  -75.0,
		},
	}
	require.NoError(t, testCtx.db.Dealerships.Insert(otherDealership))

	foreignProject := &data.Project{
		DealershipID: otherDealership.ID,
		Name:         "Someone Else's Project",
		Status:       data.ProjectStatuses.Draft,
	}
	require.NoError(t, testCtx.db.Projects.Insert(foreignProject))

	res := testCtx.request(testRequest{
		method: http.MethodPut,
		path:   fmt.Sprintf("/api/project/%s/watch", foreignProject.UUID),
		body:   map[string]any{"is_watching": true},
		token:  dealershipToken,
	})
	assert.Equal(t, http.StatusForbidden, res.statusCode, string(res.body))

	res = testCtx.request(testRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/project/%s/watchers", foreignProject.UUID),
		token:  dealershipToken,
	})
	assert.Equal(t, http.StatusForbidden, res.statusCode, string(res.body))
}

func TestPostProject_AutoSubscribesCreator(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, testCtx)

	res := testCtx.request(testRequest{
		method: http.MethodPost,
		path:   "/api/project",
		body:   map[string]any{"name": "Auto Watched Project"},
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusCreated, res.statusCode, string(res.body))

	var created data.Project
	require.NoError(t, json.Unmarshal(res.body, &created))

	isWatching, err := testCtx.db.ProjectWatchers.IsWatching(created.ID, dealershipUser)
	require.NoError(t, err)
	assert.True(t, isWatching, "the project creator should be watching it")
}

func TestNotifyDealership_SkipsActorAndNonWatchers(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, _, internalUser, internalToken := seedTestData(t, testCtx)

	// A colleague at the same dealership who never touches the project. Under
	// the old fan-out they would have been emailed anyway.
	bystander := seedDealershipUser(t, testCtx, dealershipUser.DealershipID, data.DealershipUserRoles.Submitter)

	project := &data.Project{
		DealershipID: dealershipUser.DealershipID,
		Name:         "Chat Project",
		Status:       data.ProjectStatuses.Draft,
	}
	require.NoError(t, testCtx.db.Projects.Insert(project))

	require.NoError(t, testCtx.db.ProjectWatchers.SetWatching(project.ID, dealershipUser, true))

	priceGroup := seedPriceGroup(t, testCtx, "Watch PG")
	catalogItem := seedCatalogItem(t, testCtx, priceGroup.ID, "WATCH-001")

	inlay := &data.Inlay{
		ProjectID:  project.ID,
		Name:       "Chat Inlay",
		Type:       data.InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.svg",
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID: catalogItem.ID,
		},
	}
	require.NoError(t, testCtx.db.Inlays.Insert(inlay))

	res := testCtx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/project/%s/chats", project.UUID),
		body: map[string]any{
			"message":      "Starting on this today",
			"message_type": "text",
			"inlay_uuid":   inlay.UUID,
		},
		token: internalToken,
	})
	require.Equal(t, http.StatusCreated, res.statusCode, string(res.body))

	assert.Contains(t,
		notificationEventTypesFor(t, testCtx, dealershipUser.ID, true),
		data.NotificationEventTypes.ChatMessage,
		"the watching dealership user should be notified")

	assert.Empty(t,
		notificationEventTypesFor(t, testCtx, bystander.ID, true),
		"a dealership user who is not watching should get nothing")

	assert.Empty(t,
		notificationEventTypesFor(t, testCtx, internalUser.ID, false),
		"the internal user who posted the message should not notify themselves")
}

func TestPostProjectChat_AutoSubscribesPoster(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, _, internalUser, internalToken := seedTestData(t, testCtx)

	project := &data.Project{
		DealershipID: dealershipUser.DealershipID,
		Name:         "Auto Subscribe Chat Project",
		Status:       data.ProjectStatuses.Draft,
	}
	require.NoError(t, testCtx.db.Projects.Insert(project))

	priceGroup := seedPriceGroup(t, testCtx, "Chat PG")
	catalogItem := seedCatalogItem(t, testCtx, priceGroup.ID, "CHAT-001")

	inlay := &data.Inlay{
		ProjectID:  project.ID,
		Name:       "Chat Inlay",
		Type:       data.InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.svg",
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID: catalogItem.ID,
		},
	}
	require.NoError(t, testCtx.db.Inlays.Insert(inlay))

	isWatching, err := testCtx.db.ProjectWatchers.IsWatching(project.ID, internalUser)
	require.NoError(t, err)
	require.False(t, isWatching)

	res := testCtx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/project/%s/chats", project.UUID),
		body: map[string]any{
			"message":      "Taking a look",
			"message_type": "text",
		},
		token: internalToken,
	})
	require.Equal(t, http.StatusCreated, res.statusCode, string(res.body))

	isWatching, err = testCtx.db.ProjectWatchers.IsWatching(project.ID, internalUser)
	require.NoError(t, err)
	assert.True(t, isWatching, "posting a message should subscribe the poster")
}

func TestNotifyInternal_RoleFallbackReachesNonWatchers(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, testCtx)

	// Neither of these two watches the project. order_placed falls back to
	// production (and admin), so only the production user should hear about it.
	production, _ := seedInternalUser(t, testCtx, data.InternalUserRoles.Production)
	designer, _ := seedInternalUser(t, testCtx, data.InternalUserRoles.Designer)

	project := &data.Project{
		DealershipID: dealershipUser.DealershipID,
		Name:         "Order Fallback Project",
		Status:       data.ProjectStatuses.Draft,
	}
	require.NoError(t, testCtx.db.Projects.Insert(project))

	priceGroup := seedPriceGroup(t, testCtx, "Order PG")
	catalogItem := seedCatalogItem(t, testCtx, priceGroup.ID, "ORDER-001")

	// A stock catalog inlay is ready to order immediately.
	inlay := &data.Inlay{
		ProjectID:  project.ID,
		Name:       "Stock Inlay",
		Type:       data.InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.svg",
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID: catalogItem.ID,
		},
	}
	require.NoError(t, testCtx.db.Inlays.Insert(inlay))

	res := testCtx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/project/%s/place-order", project.UUID),
		body:   map[string]any{"inlay_uuids": []string{inlay.UUID}},
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, res.statusCode, string(res.body))

	assert.Contains(t,
		notificationEventTypesFor(t, testCtx, production.ID, false),
		data.NotificationEventTypes.OrderPlaced,
		"production should hear about a new order without watching")

	assert.Empty(t,
		notificationEventTypesFor(t, testCtx, designer.ID, false),
		"a designer who is not watching should not hear about a new order")

	isWatching, err := testCtx.db.ProjectWatchers.IsWatching(project.ID, dealershipUser)
	require.NoError(t, err)
	assert.True(t, isWatching, "placing an order should subscribe the order placer")
}
