package stonedesign

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type StoneDesignModule struct {
	*app.Application
}

func NewStoneDesignModule(app *app.Application) *StoneDesignModule {
	return &StoneDesignModule{app}
}

func (m StoneDesignModule) getAccessibleProject(w http.ResponseWriter, r *http.Request) (*data.Project, bool) {
	projectUUID := r.PathValue("uuid")
	if err := m.Validate.Var(projectUUID, "required,uuid4"); err != nil {
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

func (m StoneDesignModule) HandleGetStoneDesign(w http.ResponseWriter, r *http.Request) {
	project, ok := m.getAccessibleProject(w, r)
	if !ok {
		return
	}

	design, found, err := m.Db.StoneDesigns.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	inlayUUIDs, err := m.projectInlayUUIDs(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	design.Layout = withoutMissingInlays(design.Layout, inlayUUIDs)

	m.WriteJSON(w, r, http.StatusOK, design)
}

func (m StoneDesignModule) HandlePutStoneDesign(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version          int              `json:"version" validate:"gte=0"`
		StoneShape       data.StoneShape  `json:"stone_shape" validate:"required"`
		StoneWidth       float64          `json:"stone_width" validate:"required"`
		StoneHeight      float64          `json:"stone_height" validate:"required"`
		GraniteKey       string           `json:"granite_key" validate:"required"`
		GraniteCustomHex *string          `json:"granite_custom_hex"`
		Layout           data.StoneLayout `json:"layout"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getAccessibleProject(w, r)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.Conflict, fmt.Errorf("the stone design is frozen once the project leaves draft (status: %s)", project.Status))
		return
	}

	inlayUUIDs, err := m.projectInlayUUIDs(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	design := &data.StoneDesign{
		StandardTable:    data.StandardTable{Version: body.Version},
		ProjectID:        project.ID,
		StoneShape:       body.StoneShape,
		StoneWidth:       body.StoneWidth,
		StoneHeight:      body.StoneHeight,
		GraniteKey:       body.GraniteKey,
		GraniteCustomHex: body.GraniteCustomHex,
		Layout:           withoutEmptyGroups(withoutMissingInlays(body.Layout, inlayUUIDs)),
	}

	if err := design.Validate(); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}
	if err := m.validateBackgrounds(design.Layout); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	existing, found, err := m.Db.StoneDesigns.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	switch {
	case found && body.Version != existing.Version, !found && body.Version != 0:
		err = data.ErrEditConflict
	case found:
		design.ID = existing.ID
		design.UUID = existing.UUID
		design.CreatedAt = existing.CreatedAt
		err = m.Db.StoneDesigns.Update(design)
	default:
		err = m.Db.StoneDesigns.Insert(design)
	}
	if err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			m.WriteError(w, r, m.Err.Conflict, errors.New("the stone design was changed elsewhere; reload to see the latest version"))
			return
		}
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.AutoWatchProject(project.ID, m.ContextGetUser(r))

	m.WriteJSON(w, r, http.StatusOK, design)
}

func (m StoneDesignModule) projectInlayUUIDs(projectID int) (map[string]bool, error) {
	inlays, err := m.Db.Inlays.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load inlays for project %d: %w", projectID, err)
	}
	uuids := make(map[string]bool, len(inlays))
	for _, inlay := range inlays {
		uuids[inlay.UUID] = true
	}
	return uuids, nil
}

func (m StoneDesignModule) validateBackgrounds(layout data.StoneLayout) error {
	checked := map[string]bool{}
	for _, el := range layout.Elements {
		if el.Type != data.StoneElementTypes.Background || checked[el.BackgroundUUID] {
			continue
		}
		_, found, err := m.Db.SandblastBackgrounds.GetByUUID(el.BackgroundUUID)
		if err != nil {
			return fmt.Errorf("failed to look up background %s: %w", el.BackgroundUUID, err)
		}
		if !found {
			return fmt.Errorf("%s is not a sandblast background", el.BackgroundUUID)
		}
		checked[el.BackgroundUUID] = true
	}
	return nil
}

// withoutMissingInlays drops inlay elements whose inlay is not on the project:
// deleted since the layout was saved, or never this project's to place.
func withoutMissingInlays(layout data.StoneLayout, projectInlays map[string]bool) data.StoneLayout {
	kept := make([]data.StoneElement, 0, len(layout.Elements))
	for _, el := range layout.Elements {
		if el.Type == data.StoneElementTypes.Inlay && !projectInlays[el.InlayUUID] {
			continue
		}
		kept = append(kept, el)
	}
	layout.Elements = kept
	if layout.Groups == nil {
		layout.Groups = []data.StoneGroup{}
	}
	return layout
}

func withoutEmptyGroups(layout data.StoneLayout) data.StoneLayout {
	used := map[string]bool{}
	for _, el := range layout.Elements {
		if el.GroupID != nil {
			used[*el.GroupID] = true
		}
	}
	kept := make([]data.StoneGroup, 0, len(layout.Groups))
	for _, group := range layout.Groups {
		if used[group.ID] {
			kept = append(kept, group)
		}
	}
	layout.Groups = kept
	return layout
}
