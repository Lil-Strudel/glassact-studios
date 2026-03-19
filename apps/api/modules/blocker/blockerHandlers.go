package blocker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type BlockerModule struct {
	*app.Application
}

func NewBlockerModule(app *app.Application) *BlockerModule {
	return &BlockerModule{
		app,
	}
}

func (m BlockerModule) HandleResolveBlocker(w http.ResponseWriter, r *http.Request) {
	blockerUUID := r.PathValue("uuid")

	err := m.Validate.Var(blockerUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		ResolutionNotes string `json:"resolution_notes"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	blocker, found, err := m.Db.InlayBlockers.GetByUUID(blockerUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if blocker.ResolvedAt != nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("blocker is already resolved"))
		return
	}

	user := m.ContextGetUser(r)
	userID := user.GetID()
	now := time.Now()

	blocker.ResolvedAt = &now
	blocker.ResolvedBy = &userID

	if body.ResolutionNotes != "" {
		blocker.ResolutionNotes = &body.ResolutionNotes
	}

	err = m.Db.InlayBlockers.Update(blocker)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to resolve blocker: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, blocker)
}

func (m BlockerModule) HandleGetBlocker(w http.ResponseWriter, r *http.Request) {
	blockerUUID := r.PathValue("uuid")

	err := m.Validate.Var(blockerUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	blocker, found, err := m.Db.InlayBlockers.GetByUUID(blockerUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, blocker)
}
