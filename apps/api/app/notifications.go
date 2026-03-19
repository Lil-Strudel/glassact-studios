package app

import (
	"fmt"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// SendNotificationToUser creates an in-app notification and, if the user has email enabled
// for this event type, sends an email in a background goroutine.
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

// SendNotificationToAllInternalUsers fans out notifications to every active internal user.
func (app *Application) SendNotificationToAllInternalUsers(
	eventType data.NotificationEventType,
	title, body string,
	projectID, inlayID *int,
) {
	users, err := app.Db.InternalUsers.GetAll()
	if err != nil {
		app.Log.Error("failed to get internal users for notification", "error", err, "event_type", eventType)
		return
	}

	for _, user := range users {
		if !user.IsActive {
			continue
		}
		app.SendNotificationToUser(user.ID, "internal", user.Email, eventType, title, body, projectID, inlayID)
	}
}

// SendNotificationToAllDealershipUsersForProject fans out notifications to all dealership
// users associated with the project's dealership.
func (app *Application) SendNotificationToAllDealershipUsersForProject(
	projectID int,
	eventType data.NotificationEventType,
	title, body string,
	inlayID *int,
) {
	project, found, err := app.Db.Projects.GetByID(projectID)
	if err != nil || !found {
		app.Log.Error("failed to get project for notification fan-out", "error", err, "project_id", projectID)
		return
	}

	users, err := app.Db.DealershipUsers.GetByDealershipID(project.DealershipID)
	if err != nil {
		app.Log.Error("failed to get dealership users for notification", "error", err, "project_id", projectID)
		return
	}

	for _, user := range users {
		if !user.IsActive {
			continue
		}
		app.SendNotificationToUser(user.ID, "dealership", user.Email, eventType, title, body, &projectID, inlayID)
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
