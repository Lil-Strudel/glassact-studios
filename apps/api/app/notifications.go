package app

import (
	"fmt"
	"slices"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

func (app *Application) SendNotificationToUser(
	userID int,
	userType string,
	email string,
	eventType data.NotificationEventType,
	title, body string,
	projectID, inlayID *int,
) {
	notif := data.Notification{
		EventType: eventType,
		Title:     title,
		Body:      body,
		ProjectID: projectID,
		InlayID:   inlayID,
	}

	if userType == "dealership" {
		notif.DealershipUserID = &userID
	} else {
		notif.InternalUserID = &userID
	}

	if err := app.Db.Notifications.Insert(&notif); err != nil {
		app.Log.Error("failed to insert notification", "error", err, "event_type", eventType)
		return
	}

	var emailEnabled bool
	var emailCheckErr error

	if userType == "dealership" {
		emailEnabled, emailCheckErr = app.Db.NotificationPreferences.IsEmailEnabledForDealershipUser(userID, eventType)
	} else {
		emailEnabled, emailCheckErr = app.Db.NotificationPreferences.IsEmailEnabledForInternalUser(userID, eventType)
	}

	if emailCheckErr != nil {
		app.Log.Error("failed to check notification email preference", "error", emailCheckErr, "event_type", eventType)
		return
	}

	if !emailEnabled {
		return
	}

	notifID := notif.ID
	htmlBody := buildNotificationEmailHTML(title, body, app.Cfg.BaseURL)
	textBody := fmt.Sprintf("%s\n\n%s\n\nView in GlassAct Studios: %s", title, body, app.Cfg.BaseURL)

	app.Wg.Add(1)
	go func() {
		defer app.Wg.Done()
		if err := app.Mailer.Send(email, title, htmlBody, textBody); err != nil {
			app.Log.Error("failed to send notification email", "error", err, "event_type", eventType)
			return
		}
		if err := app.Db.Notifications.MarkEmailSent(notifID); err != nil {
			app.Log.Error("failed to mark email sent", "error", err)
		}
	}()
}

// internalRoleFallback lists the internal roles that must hear about an event
// whether or not they watch the project, because the event is a request for
// work that would otherwise sit unclaimed. Every other event reaches watchers
// only.
var internalRoleFallback = map[data.NotificationEventType][]data.InternalUserRole{
	data.NotificationEventTypes.OrderPlaced: {
		data.InternalUserRoles.Production, data.InternalUserRoles.Admin,
	},
	data.NotificationEventTypes.InternalReviewRequired: {
		data.InternalUserRoles.Designer, data.InternalUserRoles.Admin,
	},
	data.NotificationEventTypes.CustomInlaySubmitted: {
		data.InternalUserRoles.Designer, data.InternalUserRoles.Admin,
	},
	data.NotificationEventTypes.ProjectDelivered: {
		data.InternalUserRoles.Billing, data.InternalUserRoles.Admin,
	},
}

// dealershipRoleFallback is the customer-side equivalent: money and approvals
// have to reach whoever is allowed to act on them, even if that person never
// touched the project.
var dealershipRoleFallback = map[data.NotificationEventType][]data.DealershipUserRole{
	data.NotificationEventTypes.InvoiceSent: {
		data.DealershipUserRoles.Admin,
	},
	data.NotificationEventTypes.InvoiceVoided: {
		data.DealershipUserRoles.Admin,
	},
	data.NotificationEventTypes.PaymentReceived: {
		data.DealershipUserRoles.Admin,
	},
	data.NotificationEventTypes.ProofReady: {
		data.DealershipUserRoles.Approver, data.DealershipUserRoles.Admin,
	},
}

// AutoWatchProject subscribes a user to a project because they just did
// something meaningful on it. Failing to subscribe should never fail the
// request that triggered it, so the error is logged rather than returned.
func (app *Application) AutoWatchProject(projectID int, user data.AuthUser) {
	if user == nil {
		return
	}

	if err := app.Db.ProjectWatchers.AutoSubscribe(projectID, user); err != nil {
		app.Log.Error("failed to auto-subscribe project watcher",
			"error", err, "project_id", projectID, "user_id", user.GetID())
	}
}

// isActor reports whether a recipient is the person who triggered the event.
// Dealership and internal ids live in separate tables, so the side has to match
// before the ids are worth comparing.
func isActor(actor data.AuthUser, userID int, isDealership bool) bool {
	if actor == nil {
		return false
	}
	return actor.IsDealership() == isDealership && actor.GetID() == userID
}

// NotifyDealership notifies the dealership users watching a project, plus any
// role that must see this event regardless of watch state, minus the actor.
func (app *Application) NotifyDealership(
	projectID int,
	actor data.AuthUser,
	eventType data.NotificationEventType,
	title, body string,
	inlayID *int,
) {
	recipients := make(map[int]*data.DealershipUser)

	watchers, err := app.Db.ProjectWatchers.GetDealershipWatchers(projectID)
	if err != nil {
		app.Log.Error("failed to get dealership watchers for notification",
			"error", err, "project_id", projectID, "event_type", eventType)
		return
	}
	for _, user := range watchers {
		recipients[user.ID] = user
	}

	if roles, hasFallback := dealershipRoleFallback[eventType]; hasFallback {
		project, found, err := app.Db.Projects.GetByID(projectID)
		if err != nil || !found {
			app.Log.Error("failed to get project for notification fallback",
				"error", err, "project_id", projectID, "event_type", eventType)
			return
		}

		users, err := app.Db.DealershipUsers.GetByDealershipID(project.DealershipID)
		if err != nil {
			app.Log.Error("failed to get dealership users for notification fallback",
				"error", err, "project_id", projectID, "event_type", eventType)
			return
		}

		for _, user := range users {
			if !user.IsActive {
				continue
			}
			if slices.Contains(roles, user.Role) {
				recipients[user.ID] = user
			}
		}
	}

	for _, user := range recipients {
		if isActor(actor, user.ID, true) {
			continue
		}
		app.SendNotificationToUser(user.ID, "dealership", user.Email, eventType, title, body, &projectID, inlayID)
	}
}

// NotifyInternal is the GlassAct-staff counterpart of NotifyDealership.
func (app *Application) NotifyInternal(
	projectID int,
	actor data.AuthUser,
	eventType data.NotificationEventType,
	title, body string,
	inlayID *int,
) {
	recipients := make(map[int]*data.InternalUser)

	watchers, err := app.Db.ProjectWatchers.GetInternalWatchers(projectID)
	if err != nil {
		app.Log.Error("failed to get internal watchers for notification",
			"error", err, "project_id", projectID, "event_type", eventType)
		return
	}
	for _, user := range watchers {
		recipients[user.ID] = user
	}

	if roles, hasFallback := internalRoleFallback[eventType]; hasFallback {
		users, err := app.Db.InternalUsers.GetAll()
		if err != nil {
			app.Log.Error("failed to get internal users for notification fallback",
				"error", err, "project_id", projectID, "event_type", eventType)
			return
		}

		for _, user := range users {
			if !user.IsActive {
				continue
			}
			if slices.Contains(roles, user.Role) {
				recipients[user.ID] = user
			}
		}
	}

	for _, user := range recipients {
		if isActor(actor, user.ID, false) {
			continue
		}
		app.SendNotificationToUser(user.ID, "internal", user.Email, eventType, title, body, &projectID, inlayID)
	}
}

func buildNotificationEmailHTML(title, body, baseURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body style="margin:0; padding:0; background-color:#ffffff; font-family:Roboto, Arial, sans-serif; color:#0a0a0a;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
      <tr>
        <td align="center" style="padding: 40px 0;">
          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width:600px; background:#ffffff; border-radius:8px; box-shadow:0 2px 4px rgba(0,0,0,0.1); padding:40px;">
            <tr>
              <td style="text-align:center;">
                <h1 style="margin:0; font-size:24px; font-weight:600; color:#0a0a0a;">%s</h1>
                <p style="margin:20px 0; font-size:16px; color:#737373;">%s</p>
                <a href="%s" style="display:inline-block; padding:12px 24px; background-color:#8b0f24; color:#ffffff; text-decoration:none; border-radius:8px; font-size:16px; font-weight:500;">
                  View in GlassAct Studios
                </a>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, title, title, body, baseURL)
}
