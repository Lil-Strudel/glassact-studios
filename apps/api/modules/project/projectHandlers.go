package project

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type ProjectModule struct {
	*app.Application
}

func NewProjectModule(app *app.Application) *ProjectModule {
	return &ProjectModule{
		app,
	}
}

func (m ProjectModule) HandleGetProjects(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var projects []*data.Project
	var err error

	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
		projects, err = m.Db.Projects.GetByDealershipID(*dealershipID)
	} else {
		projects, err = m.Db.Projects.GetAll()
	}

	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	// Dealership users get the bare project list. Internal users additionally
	// get a per-project action summary so the project list can flag, at a
	// glance, which projects need their attention.
	if user.IsDealership() {
		m.WriteJSON(w, r, http.StatusOK, projects)
		return
	}

	summaries, err := m.Db.Projects.GetActionSummaries()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	result := make([]projectListItem, len(projects))
	for i, project := range projects {
		result[i] = projectListItem{Project: project}
		if summary, ok := summaries[project.ID]; ok {
			result[i].ActionSummary = &summary
		}
	}

	m.WriteJSON(w, r, http.StatusOK, result)
}

// projectListItem embeds the project and, for internal users, the action
// summary. The summary is omitted entirely for projects with no outstanding
// internal action.
type projectListItem struct {
	*data.Project
	ActionSummary *data.ProjectActionSummary `json:"action_summary,omitempty"`
}

func (m ProjectModule) HandleGetProjectByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

func (m ProjectModule) HandlePostProject(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	var body struct {
		Name              string  `json:"name" validate:"required"`
		InternalReference *string `json:"internal_reference"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if !user.IsDealership() {
		m.WriteError(w, r, m.Err.Forbidden, fmt.Errorf("only dealership users can create projects"))
		return
	}

	dealershipID := user.GetDealershipID()
	if dealershipID == nil {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	internalRef := normalizeInternalReference(body.InternalReference)

	project := data.Project{
		Name:              body.Name,
		InternalReference: internalRef,
		Status:            data.ProjectStatuses.Draft,
		DealershipID:      *dealershipID,
	}

	err = m.Db.Projects.Insert(&project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, project)
}

func (m ProjectModule) HandlePatchProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name              *string `json:"name"`
		InternalReference *string `json:"internal_reference"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	if body.Name != nil {
		project.Name = *body.Name
	}

	if body.InternalReference != nil {
		project.InternalReference = normalizeInternalReference(body.InternalReference)
	}

	err = m.Db.Projects.Update(project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

// Projects can only be cancelled before manufacturing starts.
var cancellableStatuses = map[data.ProjectStatus]bool{
	data.ProjectStatuses.Draft:   true,
	data.ProjectStatuses.Ordered: true,
}

func (m ProjectModule) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	if !cancellableStatuses[project.Status] {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("cannot cancel a project in %s status", project.Status))
		return
	}

	project.Status = data.ProjectStatuses.Cancelled

	err = m.Db.Projects.Update(project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, project)
}

// inlayIsReady captures the same predicate the frontend uses:
//   - Stock catalog inlays are ready immediately (no proof needed).
//   - Customized catalog inlays and custom inlays are ready only when an
//     approved proof has been attached.
func inlayIsReady(inlay *data.Inlay) bool {
	if inlay.Type == data.InlayTypes.Catalog && !inlay.IsCustomized {
		return true
	}
	return inlay.ApprovedProofID != nil
}

func (m ProjectModule) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		InlayUUIDs []string `json:"inlay_uuids" validate:"required,min=1,dive,uuid4"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
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
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || *dealershipID != project.DealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("project must be in draft status to place order, currently: %s", project.Status))
		return
	}

	allInlays, err := m.Db.Inlays.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	selectedUUIDs := make(map[string]bool, len(body.InlayUUIDs))
	for _, uuid := range body.InlayUUIDs {
		selectedUUIDs[uuid] = true
	}

	selected := make([]*data.Inlay, 0, len(allInlays))
	for _, inlayItem := range allInlays {
		if selectedUUIDs[inlayItem.UUID] {
			selected = append(selected, inlayItem)
		}
	}

	if len(selected) == 0 {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("no valid inlays selected for order"))
		return
	}

	for _, inlayItem := range selected {
		if !inlayIsReady(inlayItem) {
			m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("inlay %q is not ready to order", inlayItem.Name))
			return
		}
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	orderedStep := string(data.ManufacturingSteps.Ordered)
	for _, inlayItem := range selected {
		snapshot, snapshotErr := m.buildOrderSnapshot(project.ID, inlayItem)
		if snapshotErr != nil {
			m.WriteError(w, r, m.Err.ServerError, snapshotErr)
			return
		}

		if snapErr := m.Db.OrderSnapshots.TxInsert(tx, snapshot); snapErr != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create order snapshot for inlay %q: %w", inlayItem.Name, snapErr))
			return
		}

		inlayItem.ManufacturingStep = &orderedStep
		if updateErr := m.Db.Inlays.TxUpdateFields(tx, inlayItem); updateErr != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update inlay manufacturing step: %w", updateErr))
			return
		}
	}

	now := time.Now()
	userID := user.GetID()
	project.Status = data.ProjectStatuses.Ordered
	project.OrderedAt = &now
	project.OrderedBy = &userID

	err = m.Db.Projects.TxUpdate(tx, project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update project status: %w", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.SendNotificationToAllInternalUsers(
		data.NotificationEventTypes.OrderPlaced,
		fmt.Sprintf("Order placed: %s", project.Name),
		fmt.Sprintf("A new order has been placed for project %q.", project.Name),
		&project.ID, nil,
	)

	m.WriteJSON(w, r, http.StatusOK, project)
}

// buildOrderSnapshot assembles the immutable pricing/dimension record for an
// inlay at order time. Stock catalog inlays pull from the catalog defaults;
// approved-proof inlays pull from the proof.
func (m ProjectModule) buildOrderSnapshot(projectID int, inlay *data.Inlay) (*data.OrderSnapshot, error) {
	if inlay.ApprovedProofID != nil {
		approvedProof, proofFound, err := m.Db.InlayProofs.GetByID(*inlay.ApprovedProofID)
		if err != nil {
			return nil, fmt.Errorf("failed to load approved proof for inlay %q: %w", inlay.Name, err)
		}
		if !proofFound {
			return nil, fmt.Errorf("approved proof not found for inlay %q", inlay.Name)
		}

		priceGroupID := 0
		baseCents := 0
		if approvedProof.PriceGroupID != nil {
			priceGroupID = *approvedProof.PriceGroupID
			pg, pgFound, pgErr := m.Db.PriceGroups.GetByID(*approvedProof.PriceGroupID)
			if pgErr != nil {
				return nil, fmt.Errorf("failed to load price group for inlay %q: %w", inlay.Name, pgErr)
			}
			if pgFound {
				baseCents = pg.BasePriceCents
			}
		}

		priceCents := data.ComputeAdjustedPriceCents(baseCents, approvedProof.PriceAdjustmentType, approvedProof.PriceAdjustmentValue)

		proofID := approvedProof.ID
		return &data.OrderSnapshot{
			ProjectID:            projectID,
			InlayID:              inlay.ID,
			ProofID:              &proofID,
			PriceGroupID:         priceGroupID,
			PriceCents:           priceCents,
			PriceAdjustmentType:  approvedProof.PriceAdjustmentType,
			PriceAdjustmentValue: approvedProof.PriceAdjustmentValue,
			Width:                approvedProof.Width,
			Height:               approvedProof.Height,
		}, nil
	}

	// Stock catalog inlay: derive from the catalog item.
	if inlay.Type != data.InlayTypes.Catalog || inlay.CatalogInfo == nil {
		return nil, fmt.Errorf("inlay %q has no approved proof and is not a stock catalog inlay", inlay.Name)
	}

	catalogItem, found, err := m.Db.CatalogItems.GetByID(inlay.CatalogInfo.CatalogItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to load catalog item for inlay %q: %w", inlay.Name, err)
	}
	if !found {
		return nil, fmt.Errorf("catalog item not found for inlay %q", inlay.Name)
	}

	priceGroup, pgFound, pgErr := m.Db.PriceGroups.GetByID(catalogItem.DefaultPriceGroupID)
	if pgErr != nil {
		return nil, fmt.Errorf("failed to load price group for catalog item %q: %w", catalogItem.Name, pgErr)
	}
	if !pgFound {
		return nil, fmt.Errorf("default price group not found for catalog item %q", catalogItem.Name)
	}

	return &data.OrderSnapshot{
		ProjectID:            projectID,
		InlayID:              inlay.ID,
		ProofID:              nil,
		PriceGroupID:         catalogItem.DefaultPriceGroupID,
		PriceCents:           priceGroup.BasePriceCents,
		PriceAdjustmentType:  data.PriceAdjustmentTypes.None,
		PriceAdjustmentValue: 0,
		Width:                catalogItem.DefaultWidth,
		Height:               catalogItem.DefaultHeight,
	}, nil
}

func normalizeInternalReference(in *string) *string {
	if in == nil {
		return nil
	}
	trimmed := *in
	for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t') {
		trimmed = trimmed[1:]
	}
	for len(trimmed) > 0 && (trimmed[len(trimmed)-1] == ' ' || trimmed[len(trimmed)-1] == '\t') {
		trimmed = trimmed[:len(trimmed)-1]
	}
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
