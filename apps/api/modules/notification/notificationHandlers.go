package notification

import (
	"fmt"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type NotificationModule struct {
	*app.Application
}

func NewNotificationModule(app *app.Application) *NotificationModule {
	return &NotificationModule{app}
}

// dealershipEventTypes lists the notification event types relevant to dealership users.
var dealershipEventTypes = []data.NotificationEventType{
	data.NotificationEventTypes.ProofReady,
	data.NotificationEventTypes.ProofApproved,
	data.NotificationEventTypes.ProofDeclined,
	data.NotificationEventTypes.InlayStepChanged,
	data.NotificationEventTypes.InlayBlocked,
	data.NotificationEventTypes.InlayUnblocked,
	data.NotificationEventTypes.ProjectShipped,
	data.NotificationEventTypes.ProjectDelivered,
	data.NotificationEventTypes.InvoiceSent,
	data.NotificationEventTypes.PaymentReceived,
	data.NotificationEventTypes.ChatMessage,
}

// internalEventTypes lists the notification event types relevant to internal users.
var internalEventTypes = []data.NotificationEventType{
	data.NotificationEventTypes.ProjectSubmitted,
	data.NotificationEventTypes.OrderPlaced,
	data.NotificationEventTypes.ProofReady,
	data.NotificationEventTypes.ProofApproved,
	data.NotificationEventTypes.ProofDeclined,
	data.NotificationEventTypes.ProjectDelivered,
	data.NotificationEventTypes.ChatMessage,
}

func (m *NotificationModule) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var notifications []*data.Notification
	var err error

	if user.IsDealership() {
		notifications, err = m.Db.Notifications.GetForDealershipUser(user.GetID())
	} else {
		notifications, err = m.Db.Notifications.GetForInternalUser(user.GetID())
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get notifications: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, notifications)
}

func (m *NotificationModule) HandleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var unread []*data.Notification
	var err error

	if user.IsDealership() {
		unread, err = m.Db.Notifications.GetUnreadForDealershipUser(user.GetID())
	} else {
		unread, err = m.Db.Notifications.GetUnreadForInternalUser(user.GetID())
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get unread notifications: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]int{"count": len(unread)})
}

func (m *NotificationModule) HandleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	uuidStr := r.PathValue("uuid")

	err := m.Validate.Var(uuidStr, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	notif, found, err := m.Db.Notifications.GetByUUID(uuidStr)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get notification: %w", err))
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		if notif.DealershipUserID == nil || *notif.DealershipUserID != user.GetID() {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	} else {
		if notif.InternalUserID == nil || *notif.InternalUserID != user.GetID() {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	if err := m.Db.Notifications.MarkRead(notif.ID); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to mark notification read: %w", err))
		return
	}

	notif, _, err = m.Db.Notifications.GetByUUID(uuidStr)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to reload notification: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, notif)
}

func (m *NotificationModule) HandleMarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var err error
	if user.IsDealership() {
		err = m.Db.Notifications.MarkAllReadForDealershipUser(user.GetID())
	} else {
		err = m.Db.Notifications.MarkAllReadForInternalUser(user.GetID())
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to mark all notifications read: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

type notificationPreferenceResponse struct {
	EventType    data.NotificationEventType `json:"event_type"`
	EmailEnabled bool                       `json:"email_enabled"`
}

func (m *NotificationModule) HandleGetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var relevantTypes []data.NotificationEventType
	if user.IsDealership() {
		relevantTypes = dealershipEventTypes
	} else {
		relevantTypes = internalEventTypes
	}

	var storedPrefs []*data.NotificationPreference
	var err error

	if user.IsDealership() {
		storedPrefs, err = m.Db.NotificationPreferences.GetForDealershipUser(user.GetID())
	} else {
		storedPrefs, err = m.Db.NotificationPreferences.GetForInternalUser(user.GetID())
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get notification preferences: %w", err))
		return
	}

	// Build lookup map from stored prefs
	prefMap := make(map[data.NotificationEventType]bool)
	for _, p := range storedPrefs {
		prefMap[p.EventType] = p.EmailEnabled
	}

	// Build response filling in defaults (true = enabled) for missing rows
	result := make([]notificationPreferenceResponse, len(relevantTypes))
	for i, eventType := range relevantTypes {
		emailEnabled := true
		if val, ok := prefMap[eventType]; ok {
			emailEnabled = val
		}
		result[i] = notificationPreferenceResponse{
			EventType:    eventType,
			EmailEnabled: emailEnabled,
		}
	}

	m.WriteJSON(w, r, http.StatusOK, result)
}

func (m *NotificationModule) HandlePatchNotificationPreference(w http.ResponseWriter, r *http.Request) {
	eventTypeStr := r.PathValue("event_type")

	var body struct {
		EmailEnabled bool `json:"email_enabled"`
	}

	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	eventType := data.NotificationEventType(eventTypeStr)

	user := m.ContextGetUser(r)

	// Validate the event type is relevant for this user type
	var relevantTypes []data.NotificationEventType
	if user.IsDealership() {
		relevantTypes = dealershipEventTypes
	} else {
		relevantTypes = internalEventTypes
	}

	valid := false
	for _, t := range relevantTypes {
		if t == eventType {
			valid = true
			break
		}
	}
	if !valid {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("invalid event_type %q for this user type", eventTypeStr))
		return
	}

	var err error
	if user.IsDealership() {
		err = m.Db.NotificationPreferences.UpsertForDealershipUser(user.GetID(), eventType, body.EmailEnabled)
	} else {
		err = m.Db.NotificationPreferences.UpsertForInternalUser(user.GetID(), eventType, body.EmailEnabled)
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update notification preference: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, notificationPreferenceResponse{
		EventType:    eventType,
		EmailEnabled: body.EmailEnabled,
	})
}
