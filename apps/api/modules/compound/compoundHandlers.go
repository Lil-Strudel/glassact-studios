package compound

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/google/uuid"
)

type CompoundModule struct {
	*app.Application
}

func NewCompoundModule(app *app.Application) *CompoundModule {
	return &CompoundModule{app}
}

type compoundMemberBody struct {
	CatalogItemID int     `json:"catalog_item_id" validate:"required,gt=0"`
	CenterX       float64 `json:"center_x"`
	CenterY       float64 `json:"center_y"`
	RotationDeg   float64 `json:"rotation_deg"`
}

type compoundWriteBody struct {
	Name         string               `json:"name" validate:"required,min=1,max=255"`
	Description  *string              `json:"description"`
	IsActive     bool                 `json:"is_active"`
	DisplayOrder *int                 `json:"display_order"`
	Members      []compoundMemberBody `json:"members" validate:"min=2,max=50,dive"`
}

const maxMemberOffsetInches = 120

func (m CompoundModule) membersFromBody(members []compoundMemberBody) ([]data.CatalogCompoundMember, error) {
	result := make([]data.CatalogCompoundMember, len(members))
	for i, member := range members {
		for _, v := range []float64{member.CenterX, member.CenterY} {
			if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > maxMemberOffsetInches {
				return nil, fmt.Errorf("member %d is placed out of range", i+1)
			}
		}
		if math.IsNaN(member.RotationDeg) || math.Abs(member.RotationDeg) > 360 {
			return nil, fmt.Errorf("member %d has an out-of-range rotation", i+1)
		}

		item, found, err := m.Db.CatalogItems.GetByID(member.CatalogItemID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("catalog item %d not found", member.CatalogItemID)
		}
		if item.IsBackground() {
			return nil, fmt.Errorf("%s is a sandblast background and cannot be part of a compound", item.CatalogCode)
		}
		if !item.IsActive {
			return nil, fmt.Errorf("%s is disabled in the catalog", item.CatalogCode)
		}

		result[i] = data.CatalogCompoundMember{
			CatalogItemID: member.CatalogItemID,
			CenterX:       member.CenterX,
			CenterY:       member.CenterY,
			RotationDeg:   member.RotationDeg,
		}
	}
	return result, nil
}

func (m CompoundModule) getCompound(w http.ResponseWriter, r *http.Request) (*data.CatalogCompound, bool) {
	compoundUUID := r.PathValue("uuid")
	if err := m.Validate.Var(compoundUUID, "required,uuid4"); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return nil, false
	}

	compound, found, err := m.Db.CatalogCompounds.GetByUUID(compoundUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return nil, false
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return nil, false
	}
	return compound, true
}

func (m CompoundModule) writeSaveError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, data.ErrEditConflict) {
		m.WriteError(w, r, m.Err.Conflict, errors.New("the compound was changed elsewhere; reload to see the latest version"))
		return
	}
	m.WriteError(w, r, m.Err.ServerError, err)
}

func (m CompoundModule) HandleGetCompounds(w http.ResponseWriter, r *http.Request) {
	compounds, err := m.Db.CatalogCompounds.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	m.WriteJSON(w, r, http.StatusOK, compounds)
}

func (m CompoundModule) HandleBrowseCompounds(w http.ResponseWriter, r *http.Request) {
	compounds, err := m.Db.CatalogCompounds.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	offerable := make([]*data.CatalogCompound, 0, len(compounds))
	for _, compound := range compounds {
		if compound.IsOfferable() {
			offerable = append(offerable, compound)
		}
	}
	m.WriteJSON(w, r, http.StatusOK, offerable)
}

func (m CompoundModule) HandleGetCompound(w http.ResponseWriter, r *http.Request) {
	compound, ok := m.getCompound(w, r)
	if !ok {
		return
	}
	m.WriteJSON(w, r, http.StatusOK, compound)
}

func (m CompoundModule) HandlePostCompound(w http.ResponseWriter, r *http.Request) {
	var body compoundWriteBody
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	members, err := m.membersFromBody(body.Members)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	compound := &data.CatalogCompound{
		Name:         body.Name,
		Description:  body.Description,
		IsActive:     body.IsActive,
		DisplayOrder: body.DisplayOrder,
		Members:      members,
	}
	if err := m.Db.CatalogCompounds.Insert(compound); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	saved, _, err := m.Db.CatalogCompounds.GetByUUID(compound.UUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	m.WriteJSON(w, r, http.StatusCreated, saved)
}

func (m CompoundModule) HandlePutCompound(w http.ResponseWriter, r *http.Request) {
	var body struct {
		compoundWriteBody
		Version int `json:"version" validate:"required,gt=0"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	compound, ok := m.getCompound(w, r)
	if !ok {
		return
	}

	members, err := m.membersFromBody(body.Members)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	compound.Name = body.Name
	compound.Description = body.Description
	compound.IsActive = body.IsActive
	compound.DisplayOrder = body.DisplayOrder
	compound.Members = members
	compound.Version = body.Version
	if err := m.Db.CatalogCompounds.Update(compound); err != nil {
		m.writeSaveError(w, r, err)
		return
	}

	saved, _, err := m.Db.CatalogCompounds.GetByUUID(compound.UUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	m.WriteJSON(w, r, http.StatusOK, saved)
}

func (m CompoundModule) HandlePatchCompound(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IsActive *bool `json:"is_active" validate:"required"`
		Version  int   `json:"version" validate:"required,gt=0"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	compound, ok := m.getCompound(w, r)
	if !ok {
		return
	}

	compound.IsActive = *body.IsActive
	compound.Version = body.Version
	if err := m.Db.CatalogCompounds.Update(compound); err != nil {
		m.writeSaveError(w, r, err)
		return
	}
	m.WriteJSON(w, r, http.StatusOK, compound)
}

func (m CompoundModule) HandleDeleteCompound(w http.ResponseWriter, r *http.Request) {
	compound, ok := m.getCompound(w, r)
	if !ok {
		return
	}
	if err := m.Db.CatalogCompounds.Delete(compound.ID); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandlePostProjectCompound adds every member of a compound to a draft project
// as an ordinary stock catalog inlay and drops them onto the project's stone as
// one group, all or nothing.
func (m CompoundModule) HandlePostProjectCompound(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CompoundUUID string `json:"compound_uuid" validate:"required,uuid4"`
	}
	if err := m.ReadJSONBody(w, r, &body); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	projectUUID := r.PathValue("uuid")
	if err := m.Validate.Var(projectUUID, "required,uuid4"); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	dealershipID := user.GetDealershipID()
	if !user.IsDealership() || dealershipID == nil || *dealershipID != project.DealershipID {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft status, current status: %s", project.Status))
		return
	}

	compound, found, err := m.Db.CatalogCompounds.GetByUUID(body.CompoundUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found || !compound.IsOfferable() {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("that compound is not available"))
		return
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	inlays := make([]*data.Inlay, len(compound.Members))
	elements := make([]data.StoneElement, len(compound.Members))
	for i, member := range compound.Members {
		inlay := &data.Inlay{
			ProjectID:  project.ID,
			Name:       member.CatalogItem.Name,
			Type:       data.InlayTypes.Catalog,
			PreviewURL: member.CatalogItem.SvgURL,
			CatalogInfo: &data.InlayCatalogInfo{
				CatalogItemID: member.CatalogItemID,
			},
		}
		if err := m.Db.Inlays.TxInsert(tx, inlay); err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to add %s from compound %s: %w", member.CatalogItem.CatalogCode, compound.UUID, err))
			return
		}
		inlays[i] = inlay
		elements[i] = data.StoneElement{
			ID:        uuid.NewString(),
			Type:      data.StoneElementTypes.Inlay,
			X:         member.CenterX,
			Y:         member.CenterY,
			Rotation:  member.RotationDeg,
			InlayUUID: inlay.UUID,
		}
	}

	compoundUUID := compound.UUID
	group := data.StoneGroup{ID: uuid.NewString(), Name: compound.Name, CompoundUUID: &compoundUUID}
	if _, err := m.Db.StoneDesigns.TxAppendGroup(tx, project.ID, group, elements); err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			m.WriteError(w, r, m.Err.Conflict, errors.New("the stone design was changed at the same moment; try again"))
			return
		}
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.AutoWatchProject(project.ID, user)

	m.WriteJSON(w, r, http.StatusCreated, inlays)
}
