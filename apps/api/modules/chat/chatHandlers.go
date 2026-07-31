package chat

import (
	"fmt"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type ChatModule struct {
	*app.Application
}

func NewChatModule(app *app.Application) *ChatModule {
	return &ChatModule{app}
}

func (m ChatModule) getProjectWithAccessCheck(w http.ResponseWriter, r *http.Request) (*data.Project, bool) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return nil, false
	}

	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, false
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, false
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return nil, false
		}
	}

	return project, true
}

func (m ChatModule) HandleGetProjectChats(w http.ResponseWriter, r *http.Request) {
	project, ok := m.getProjectWithAccessCheck(w, r)
	if !ok {
		return
	}

	chats, err := m.Db.ProjectChats.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, chats)
}

func (m ChatModule) HandlePostProjectChat(w http.ResponseWriter, r *http.Request) {
	project, ok := m.getProjectWithAccessCheck(w, r)
	if !ok {
		return
	}

	var body struct {
		Message       string  `json:"message" validate:"required"`
		MessageType   string  `json:"message_type" validate:"required,oneof=text image"`
		AttachmentURL *string `json:"attachment_url"`
		// Optional: tags the message with the inlay it is about so the thread
		// can label it and link there. Not a scope.
		InlayUUID *string `json:"inlay_uuid" validate:"omitempty,uuid4"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.MessageType == string(data.ChatMessageTypes.Image) && body.AttachmentURL == nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("attachment_url is required for image messages"))
		return
	}

	var taggedInlay *data.Inlay
	if body.InlayUUID != nil {
		inlay, found, err := m.Db.Inlays.GetByUUID(*body.InlayUUID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}

		// Never trust the client's inlay/project pairing.
		if !found || inlay.ProjectID != project.ID {
			m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("inlay %s does not belong to project %s", *body.InlayUUID, project.UUID))
			return
		}

		taggedInlay = inlay
	}

	user := m.ContextGetUser(r)
	chat := data.ProjectChat{
		ProjectID:     project.ID,
		MessageType:   data.ChatMessageType(body.MessageType),
		Message:       body.Message,
		AttachmentURL: body.AttachmentURL,
	}

	if taggedInlay != nil {
		chat.InlayID = &taggedInlay.ID
	}

	if user.IsDealership() {
		userID := user.GetID()
		chat.DealershipUserID = &userID
	} else {
		userID := user.GetID()
		chat.InternalUserID = &userID
	}

	err = m.Db.ProjectChats.Insert(&chat)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.AutoWatchProject(project.ID, user)

	// Tagged messages deep-link the notification to the inlay; untagged ones
	// land on the project page.
	title := fmt.Sprintf("New message on project: %s", project.Name)
	var notifyInlayID *int
	if taggedInlay != nil {
		title = fmt.Sprintf("New message on inlay: %s", taggedInlay.Name)
		notifyInlayID = &taggedInlay.ID
	}

	if user.IsInternal() {
		m.NotifyDealership(
			project.ID,
			user,
			data.NotificationEventTypes.ChatMessage,
			title,
			body.Message,
			notifyInlayID,
		)
	} else {
		m.NotifyInternal(
			project.ID,
			user,
			data.NotificationEventTypes.ChatMessage,
			title,
			body.Message,
			notifyInlayID,
		)
	}

	m.WriteJSON(w, r, http.StatusCreated, chat)
}
