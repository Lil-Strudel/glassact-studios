package inlayChat

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type InlayChatModule struct {
	*app.Application
}

func NewInlayChatModule(app *app.Application) *InlayChatModule {
	return &InlayChatModule{
		app,
	}
}

func (m InlayChatModule) HandleGetInlayChats(w http.ResponseWriter, r *http.Request) {
	inlayChats, err := m.Db.InlayChats.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlayChats)
}

func (m InlayChatModule) HandleGetInlayChatsByInlayUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	inlayChats, err := m.Db.InlayChats.GetByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlayChats)
}

func (m InlayChatModule) HandleGetInlayChatByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlayChat, found, err := m.Db.InlayChats.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlayChat)
}

func (m InlayChatModule) HandlePostInlayChat(w http.ResponseWriter, r *http.Request) {
	var body struct {
		InlayID     int                  `json:"inlay_id" validate:"required"`
		MessageType data.ChatMessageType `json:"message_type" validate:"required"`
		Message     string               `json:"message" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user := m.ContextGetUser(r)

	inlayChat := data.InlayChat{
		InlayID:          body.InlayID,
		DealershipUserID: &user.ID,
		MessageType:      body.MessageType,
		Message:          body.Message,
	}

	err = m.Db.InlayChats.Insert(&inlayChat)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlayChat)
}
