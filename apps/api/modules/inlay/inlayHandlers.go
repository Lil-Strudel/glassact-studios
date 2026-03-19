package inlay

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

type InlayModule struct {
	*app.Application
}

func NewInlayModule(app *app.Application) *InlayModule {
	return &InlayModule{
		app,
	}
}

func (m InlayModule) getProjectForInlayAccess(w http.ResponseWriter, r *http.Request, projectUUID string) (*data.Project, bool) {
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

func (m InlayModule) validateInlayOwnership(w http.ResponseWriter, r *http.Request, inlay *data.Inlay) (*data.Project, bool) {
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

	return project, true
}

type InlayWithProofStatus struct {
	*data.Inlay
	HasPendingProof             bool    `json:"has_pending_proof"`
	LatestProofStatus           *string `json:"latest_proof_status"`
	ApprovedProofPriceGroupID   *int    `json:"approved_proof_price_group_id"`
	ApprovedProofPriceGroupName *string `json:"approved_proof_price_group_name"`
	ApprovedProofPriceCents     *int    `json:"approved_proof_price_cents"`
	HasActiveBlocker            bool    `json:"has_active_blocker"`
}

func (m InlayModule) HandleGetInlaysByProject(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectForInlayAccess(w, r, projectUUID)
	if !ok {
		return
	}

	inlays, err := m.Db.Inlays.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	result := make([]InlayWithProofStatus, len(inlays))
	for i, inlay := range inlays {
		result[i] = InlayWithProofStatus{
			Inlay:             inlay,
			HasPendingProof:   false,
			LatestProofStatus: nil,
		}

		latestProof, found, err := m.Db.InlayProofs.GetLatestByInlayID(inlay.ID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}

		if found {
			status := string(latestProof.Status)
			result[i].LatestProofStatus = &status
			result[i].HasPendingProof = latestProof.Status == data.ProofStatuses.Pending
		}

		if inlay.ApprovedProofID != nil {
			approvedProof, found, err := m.Db.InlayProofs.GetByID(*inlay.ApprovedProofID)
			if err != nil {
				m.WriteError(w, r, m.Err.ServerError, err)
				return
			}

			if found && approvedProof.PriceGroupID != nil {
				result[i].ApprovedProofPriceGroupID = approvedProof.PriceGroupID
				result[i].ApprovedProofPriceCents = approvedProof.PriceCents

				priceGroup, found, err := m.Db.PriceGroups.GetByID(*approvedProof.PriceGroupID)
				if err != nil {
					m.WriteError(w, r, m.Err.ServerError, err)
					return
				}
				if found {
					result[i].ApprovedProofPriceGroupName = &priceGroup.Name
					if result[i].ApprovedProofPriceCents == nil {
						result[i].ApprovedProofPriceCents = &priceGroup.BasePriceCents
					}
				}
			}
		}

		unresolvedBlockers, err := m.Db.InlayBlockers.GetUnresolved(inlay.ID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		result[i].HasActiveBlocker = len(unresolvedBlockers) > 0
	}

	m.WriteJSON(w, r, http.StatusOK, result)
}

func (m InlayModule) HandleGetInlayByUUID(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if _, ok := m.validateInlayOwnership(w, r, inlay); !ok {
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlay)
}

func (m InlayModule) HandlePostCatalogInlay(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name               string `json:"name" validate:"required"`
		CatalogItemID      int    `json:"catalog_item_id" validate:"required,gt=0"`
		CustomizationNotes string `json:"customization_notes"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectForInlayAccess(w, r, projectUUID)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft or designing status, current status: %s", project.Status))
		return
	}

	catalogItem, found, err := m.Db.CatalogItems.GetByID(body.CatalogItemID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("catalog item with id %d not found", body.CatalogItemID))
		return
	}

	inlay := data.Inlay{
		ProjectID:  project.ID,
		Name:       body.Name,
		Type:       data.InlayTypes.Catalog,
		PreviewURL: catalogItem.SvgURL,
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID:      body.CatalogItemID,
			CustomizationNotes: body.CustomizationNotes,
		},
	}

	err = m.Db.Inlays.Insert(&inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, inlay)
}

func (m InlayModule) HandlePostCustomInlay(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name            string  `json:"name" validate:"required"`
		Description     string  `json:"description" validate:"required"`
		RequestedWidth  float64 `json:"requested_width"`
		RequestedHeight float64 `json:"requested_height"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectForInlayAccess(w, r, projectUUID)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft or designing status, current status: %s", project.Status))
		return
	}

	inlay := data.Inlay{
		ProjectID:  project.ID,
		Name:       body.Name,
		Type:       data.InlayTypes.Custom,
		PreviewURL: "",
		CustomInfo: &data.InlayCustomInfo{
			Description:     body.Description,
			RequestedWidth:  body.RequestedWidth,
			RequestedHeight: body.RequestedHeight,
		},
	}

	err = m.Db.Inlays.Insert(&inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, inlay)
}

func (m InlayModule) HandlePatchInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Name *string `json:"name"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	project, ok := m.validateInlayOwnership(w, r, inlay)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only modify inlays on projects in draft or designing status"))
		return
	}

	if body.Name != nil {
		inlay.Name = *body.Name
	}

	err = m.Db.Inlays.Update(inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlay)
}

func (m InlayModule) HandleDeleteInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	project, ok := m.validateInlayOwnership(w, r, inlay)
	if !ok {
		return
	}

	if project.Status != data.ProjectStatuses.Draft && project.Status != data.ProjectStatuses.Designing {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only remove inlays from projects in draft or designing status"))
		return
	}

	err = m.Db.Inlays.Delete(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

var preOrderStatuses = map[data.ProjectStatus]bool{
	data.ProjectStatuses.Draft:           true,
	data.ProjectStatuses.Designing:       true,
	data.ProjectStatuses.PendingApproval: true,
	data.ProjectStatuses.Approved:        true,
}

var manufacturingStepOrder = []data.ManufacturingStep{
	data.ManufacturingSteps.Ordered,
	data.ManufacturingSteps.MaterialsPrep,
	data.ManufacturingSteps.Cutting,
	data.ManufacturingSteps.FirePolish,
	data.ManufacturingSteps.Packaging,
	data.ManufacturingSteps.Shipped,
	data.ManufacturingSteps.Delivered,
}

func manufacturingStepIndex(step data.ManufacturingStep) int {
	for i, s := range manufacturingStepOrder {
		if s == step {
			return i
		}
	}
	return -1
}

type KanbanInlay struct {
	*data.Inlay
	ProjectName    string `json:"project_name"`
	DealershipName string `json:"dealership_name"`
	HasHardBlocker bool   `json:"has_hard_blocker"`
}

func (m InlayModule) HandleGetKanbanInlays(w http.ResponseWriter, r *http.Request) {
	query := postgres.SELECT(
		table.Inlays.AllColumns,
		table.Projects.Name.AS("project_name"),
		table.Dealerships.Name.AS("dealership_name"),
	).FROM(
		table.Inlays.
			INNER_JOIN(table.Projects, table.Projects.ID.EQ(table.Inlays.ProjectID)).
			INNER_JOIN(table.Dealerships, table.Dealerships.ID.EQ(table.Projects.DealershipID)),
	).WHERE(
		table.Inlays.ManufacturingStep.IS_NOT_NULL(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []struct {
		model.Inlays
		ProjectName    string `alias:"project_name"`
		DealershipName string `alias:"dealership_name"`
	}
	err := query.QueryContext(ctx, m.Db.STDB, &dest)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to query kanban inlays: %w", err))
		return
	}

	result := make([]KanbanInlay, len(dest))
	for i, d := range dest {
		inlay := data.Inlay{
			StandardTable: data.StandardTable{
				ID:        int(d.Inlays.ID),
				UUID:      d.Inlays.UUID.String(),
				CreatedAt: d.Inlays.CreatedAt,
				UpdatedAt: d.Inlays.UpdatedAt,
				Version:   int(d.Inlays.Version),
			},
			ProjectID:         int(d.Inlays.ProjectID),
			Name:              d.Inlays.Name,
			Type:              data.InlayType(d.Inlays.Type),
			PreviewURL:        d.Inlays.PreviewURL,
			ExcludedFromOrder: d.Inlays.ExcludedFromOrder,
			ManufacturingStep: d.Inlays.ManufacturingStep,
		}

		if d.Inlays.ApprovedProofID != nil {
			id := int(*d.Inlays.ApprovedProofID)
			inlay.ApprovedProofID = &id
		}

		hasHardBlocker := false
		unresolvedBlockers, err := m.Db.InlayBlockers.GetUnresolved(inlay.ID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get blockers for inlay %d: %w", inlay.ID, err))
			return
		}
		for _, b := range unresolvedBlockers {
			if b.BlockerType == data.BlockerTypes.Hard {
				hasHardBlocker = true
				break
			}
		}

		result[i] = KanbanInlay{
			Inlay:          &inlay,
			ProjectName:    d.ProjectName,
			DealershipName: d.DealershipName,
			HasHardBlocker: hasHardBlocker,
		}
	}

	m.WriteJSON(w, r, http.StatusOK, result)
}

func (m InlayModule) HandlePatchInlayStep(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Step data.ManufacturingStep `json:"step" validate:"required"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if manufacturingStepIndex(body.Step) == -1 {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("invalid manufacturing step: %s", body.Step))
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	unresolvedBlockers, err := m.Db.InlayBlockers.GetUnresolved(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to check blockers: %w", err))
		return
	}
	for _, b := range unresolvedBlockers {
		if b.BlockerType == data.BlockerTypes.Hard {
			m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("inlay has unresolved hard blockers that must be resolved before moving to the next step"))
			return
		}
	}

	user := m.ContextGetUser(r)
	userID := user.GetID()
	now := time.Now()

	var currentStepIdx int
	var currentStep data.ManufacturingStep
	if inlay.ManufacturingStep != nil {
		currentStep = data.ManufacturingStep(*inlay.ManufacturingStep)
		currentStepIdx = manufacturingStepIndex(currentStep)
	} else {
		currentStepIdx = -1
	}

	destStepIdx := manufacturingStepIndex(body.Step)

	var newEventType data.MilestoneEventType
	if destStepIdx >= currentStepIdx {
		newEventType = data.MilestoneEventTypes.Entered
	} else {
		newEventType = data.MilestoneEventTypes.Reverted
	}

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	if inlay.ManufacturingStep != nil {
		exitedMilestone := data.InlayMilestone{
			InlayID:     inlay.ID,
			Step:        currentStep,
			EventType:   data.MilestoneEventTypes.Exited,
			PerformedBy: userID,
			EventTime:   now,
		}
		if err := m.Db.InlayMilestones.TxInsert(tx, &exitedMilestone); err != nil {
			m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to insert exited milestone: %w", err))
			return
		}
	}

	enteredMilestone := data.InlayMilestone{
		InlayID:     inlay.ID,
		Step:        body.Step,
		EventType:   newEventType,
		PerformedBy: userID,
		EventTime:   now,
	}
	if err := m.Db.InlayMilestones.TxInsert(tx, &enteredMilestone); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to insert entered milestone: %w", err))
		return
	}

	newStep := string(body.Step)
	inlay.ManufacturingStep = &newStep
	if err := m.Db.Inlays.TxUpdateFields(tx, inlay); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to update inlay manufacturing step: %w", err))
		return
	}

	if err := tx.Commit(); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.SendNotificationToAllDealershipUsersForProject(
		inlay.ProjectID,
		data.NotificationEventTypes.InlayStepChanged,
		fmt.Sprintf("Inlay moved to %s: %s", body.Step, inlay.Name),
		fmt.Sprintf("Inlay %q has moved to the %s step.", inlay.Name, body.Step),
		&inlay.ID,
	)

	m.tryAdvanceProjectStatus(w, r, inlay.ProjectID)

	m.WriteJSON(w, r, http.StatusOK, inlay)
}

func (m InlayModule) tryAdvanceProjectStatus(w http.ResponseWriter, r *http.Request, projectID int) {
	project, found, err := m.Db.Projects.GetByID(projectID)
	if err != nil || !found {
		return
	}

	advanceable := map[data.ProjectStatus]bool{
		data.ProjectStatuses.Ordered:      true,
		data.ProjectStatuses.InProduction: true,
		data.ProjectStatuses.Shipped:      true,
	}
	if !advanceable[project.Status] {
		return
	}

	projectInlays, err := m.Db.Inlays.GetByProjectID(projectID)
	if err != nil {
		return
	}

	var activeInlays []*data.Inlay
	for _, inlay := range projectInlays {
		if !inlay.ExcludedFromOrder {
			activeInlays = append(activeInlays, inlay)
		}
	}

	if len(activeInlays) == 0 {
		return
	}

	allDelivered := true
	allShipped := true
	for _, inlay := range activeInlays {
		if inlay.ManufacturingStep == nil {
			allDelivered = false
			allShipped = false
			break
		}
		step := data.ManufacturingStep(*inlay.ManufacturingStep)
		if step != data.ManufacturingSteps.Delivered {
			allDelivered = false
		}
		if step != data.ManufacturingSteps.Delivered && step != data.ManufacturingSteps.Shipped {
			allShipped = false
		}
	}

	var newStatus data.ProjectStatus
	if allDelivered {
		newStatus = data.ProjectStatuses.Delivered
	} else if allShipped {
		newStatus = data.ProjectStatuses.Shipped
	} else {
		return
	}

	if project.Status == newStatus {
		return
	}

	project.Status = newStatus
	if err := m.Db.Projects.Update(project); err != nil {
		return
	}

	if newStatus == data.ProjectStatuses.Shipped {
		m.SendNotificationToAllDealershipUsersForProject(
			projectID,
			data.NotificationEventTypes.ProjectShipped,
			fmt.Sprintf("Project shipped: %s", project.Name),
			fmt.Sprintf("Your project %q has been shipped.", project.Name),
			nil,
		)
	} else if newStatus == data.ProjectStatuses.Delivered {
		m.SendNotificationToAllDealershipUsersForProject(
			projectID,
			data.NotificationEventTypes.ProjectDelivered,
			fmt.Sprintf("Project delivered: %s", project.Name),
			fmt.Sprintf("Your project %q has been delivered.", project.Name),
			nil,
		)
		m.SendNotificationToAllInternalUsers(
			data.NotificationEventTypes.ProjectDelivered,
			fmt.Sprintf("Project delivered: %s", project.Name),
			fmt.Sprintf("Project %q has been delivered and is ready for invoicing.", project.Name),
			&projectID, nil,
		)
	}
}

func (m InlayModule) HandleGetBlockersByInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
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
		if _, ok := m.validateInlayOwnership(w, r, inlay); !ok {
			return
		}
	}

	blockers, err := m.Db.InlayBlockers.GetByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, blockers)
}

func (m InlayModule) HandlePostBlocker(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		BlockerType data.BlockerType `json:"blocker_type" validate:"required"`
		Reason      string           `json:"reason" validate:"required"`
		StepBlocked string           `json:"step_blocked" validate:"required"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.BlockerType != data.BlockerTypes.Soft && body.BlockerType != data.BlockerTypes.Hard {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("invalid blocker_type: must be 'soft' or 'hard'"))
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	userID := user.GetID()

	blocker := data.InlayBlocker{
		InlayID:     inlay.ID,
		BlockerType: body.BlockerType,
		Reason:      body.Reason,
		StepBlocked: body.StepBlocked,
		CreatedBy:   &userID,
	}

	err = m.Db.InlayBlockers.Insert(&blocker)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to insert blocker: %w", err))
		return
	}

	m.SendNotificationToAllDealershipUsersForProject(
		inlay.ProjectID,
		data.NotificationEventTypes.InlayBlocked,
		fmt.Sprintf("Issue with inlay: %s", inlay.Name),
		fmt.Sprintf("Inlay %q has been blocked: %s", inlay.Name, body.Reason),
		&inlay.ID,
	)

	m.WriteJSON(w, r, http.StatusCreated, blocker)
}

func (m InlayModule) HandleExcludeInlay(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Excluded bool `json:"excluded"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	inlay, found, err := m.Db.Inlays.GetByUUID(inlayUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	project, ok := m.validateInlayOwnership(w, r, inlay)
	if !ok {
		return
	}

	if !preOrderStatuses[project.Status] {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only change inlay inclusion on projects before ordering"))
		return
	}

	inlay.ExcludedFromOrder = body.Excluded

	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	err = m.Db.Inlays.TxUpdateFields(tx, inlay)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, inlay)
}
