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

func (m ChatModule) getInlayWithAccessCheck(w http.ResponseWriter, r *http.Request) (*data.Inlay, bool) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return nil, false
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, false
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, false
	}

	project, found, err := m.Db.Projects.GetByID(inlay.ProjectID)
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

	return inlay, true
}

func (m ChatModule) HandleGetInlayChats(w http.ResponseWriter, r *http.Request) {
	inlay, ok := m.getInlayWithAccessCheck(w, r)
	if !ok {
		return
	}

	chats, err := m.Db.InlayChats.GetByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, chats)
}

func (m ChatModule) HandlePostInlayChat(w http.ResponseWriter, r *http.Request) {
	inlay, ok := m.getInlayWithAccessCheck(w, r)
	if !ok {
		return
	}

	var body struct {
		Message       string  `json:"message" validate:"required"`
		MessageType   string  `json:"message_type" validate:"required,oneof=text image"`
		AttachmentURL *string `json:"attachment_url"`
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

	user := m.ContextGetUser(r)
	chat := data.InlayChat{
		InlayID:       inlay.ID,
		MessageType:   data.ChatMessageType(body.MessageType),
		Message:       body.Message,
		AttachmentURL: body.AttachmentURL,
	}

	if user.IsDealership() {
		userID := user.GetID()
		chat.DealershipUserID = &userID
	} else {
		userID := user.GetID()
		chat.InternalUserID = &userID
	}

	err = m.Db.InlayChats.Insert(&chat)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, chat)
}
