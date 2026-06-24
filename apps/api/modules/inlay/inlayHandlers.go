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

// InlayWithProofStatus is the API response shape for inlays. The frontend uses
// `is_ready` + price fields directly so the user can see what's blocking the
// order and what each line will cost.
type InlayWithProofStatus struct {
	*data.Inlay
	HasPendingProof      bool                     `json:"has_pending_proof"`
	LatestProofStatus    *string                  `json:"latest_proof_status"`
	IsReady              bool                     `json:"is_ready"`
	PriceGroupID         *int                     `json:"price_group_id"`
	PriceGroupName       *string                  `json:"price_group_name"`
	PriceCents           *int                     `json:"price_cents"`
	PriceAdjustmentType  data.PriceAdjustmentType `json:"price_adjustment_type"`
	PriceAdjustmentValue float64                  `json:"price_adjustment_value"`
}

// inlayPricing carries the resolved pricing for an inlay: which price group
// applies, its display name, the final per-unit price after adjustment, and the
// adjustment itself (so the frontend can render the "PG1 + 20%" formula).
type inlayPricing struct {
	PriceGroupID    *int
	PriceGroupName  *string
	PriceCents      *int
	AdjustmentType  data.PriceAdjustmentType
	AdjustmentValue float64
}

// inlayIsReady is the single source of truth for "can this inlay be ordered?"
// It mirrors the rule in projectHandlers.inlayIsReady.
func inlayIsReady(inlay *data.Inlay) bool {
	if inlay.Type == data.InlayTypes.Catalog && !inlay.IsCustomized {
		return true
	}
	return inlay.ApprovedProofID != nil
}

// buildInlayPricing resolves the price group, final per-unit price, and the
// adjustment formula for an inlay. Stock catalog inlays use the catalog's
// default price group (no adjustment); otherwise we look at the approved proof
// (or, for customized catalog inlays still pending internal review, the latest
// proof's proposed pricing).
func (m InlayModule) buildInlayPricing(inlay *data.Inlay) (inlayPricing, error) {
	if inlay.ApprovedProofID != nil {
		proof, found, err := m.Db.InlayProofs.GetByID(*inlay.ApprovedProofID)
		if err != nil {
			return inlayPricing{}, err
		}
		if found && proof.PriceGroupID != nil {
			return m.pricingFromProof(proof)
		}
	}

	// Customized catalog inlay still awaiting internal review: pull the latest
	// pending internal-authority proof to surface the proposed price.
	if inlay.Type == data.InlayTypes.Catalog && inlay.IsCustomized && inlay.ApprovedProofID == nil {
		latest, found, err := m.Db.InlayProofs.GetLatestByInlayID(inlay.ID)
		if err != nil {
			return inlayPricing{}, err
		}
		if found && latest.PriceGroupID != nil {
			return m.pricingFromProof(latest)
		}
	}

	// Stock catalog inlay: catalog defaults, no adjustment.
	if inlay.Type == data.InlayTypes.Catalog && inlay.CatalogInfo != nil {
		catalogItem, found, err := m.Db.CatalogItems.GetByID(inlay.CatalogInfo.CatalogItemID)
		if err != nil {
			return inlayPricing{}, err
		}
		if found {
			pg, pgFound, pgErr := m.Db.PriceGroups.GetByID(catalogItem.DefaultPriceGroupID)
			if pgErr != nil {
				return inlayPricing{}, pgErr
			}
			priceGroupID := catalogItem.DefaultPriceGroupID
			pricing := inlayPricing{
				PriceGroupID:   &priceGroupID,
				AdjustmentType: data.PriceAdjustmentTypes.None,
			}
			if pgFound {
				name := pg.Name
				base := pg.BasePriceCents
				pricing.PriceGroupName = &name
				pricing.PriceCents = &base
			}
			return pricing, nil
		}
	}

	return inlayPricing{AdjustmentType: data.PriceAdjustmentTypes.None}, nil
}

// pricingFromProof resolves a proof's price group base and applies its
// adjustment to produce the final per-unit price plus the formula components.
func (m InlayModule) pricingFromProof(proof *data.InlayProof) (inlayPricing, error) {
	pg, pgFound, pgErr := m.Db.PriceGroups.GetByID(*proof.PriceGroupID)
	if pgErr != nil {
		return inlayPricing{}, pgErr
	}

	pricing := inlayPricing{
		PriceGroupID:    proof.PriceGroupID,
		AdjustmentType:  proof.PriceAdjustmentType,
		AdjustmentValue: proof.PriceAdjustmentValue,
	}
	if pgFound {
		name := pg.Name
		pricing.PriceGroupName = &name
		priceCents := data.ComputeAdjustedPriceCents(pg.BasePriceCents, proof.PriceAdjustmentType, proof.PriceAdjustmentValue)
		pricing.PriceCents = &priceCents
	}

	return pricing, nil
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
			Inlay:   inlay,
			IsReady: inlayIsReady(inlay),
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

		pricing, err := m.buildInlayPricing(inlay)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		result[i].PriceGroupID = pricing.PriceGroupID
		result[i].PriceGroupName = pricing.PriceGroupName
		result[i].PriceCents = pricing.PriceCents
		result[i].PriceAdjustmentType = pricing.AdjustmentType
		result[i].PriceAdjustmentValue = pricing.AdjustmentValue
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

	// The Customization sub-object is optional. When provided, the inlay was
	// produced via the customizer and needs internal review.
	var body struct {
		Name               string `json:"name" validate:"required"`
		CatalogItemID      int    `json:"catalog_item_id" validate:"required,gt=0"`
		CustomizationNotes string `json:"customization_notes"`
		Customization      *struct {
			BakedDesignAssetURL string                 `json:"baked_design_asset_url" validate:"required"`
			ScaleFactor         float64                `json:"scale_factor" validate:"required,gt=0"`
			Width               float64                `json:"width" validate:"required,gt=0"`
			Height              float64                `json:"height" validate:"required,gt=0"`
			ColorOverrides      map[string]interface{} `json:"color_overrides"`
		} `json:"customization"`
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

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft status, current status: %s", project.Status))
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

	if body.Customization == nil {
		// Stock catalog inlay — ready immediately, no proof needed.
		inlay := data.Inlay{
			ProjectID:    project.ID,
			Name:         body.Name,
			Type:         data.InlayTypes.Catalog,
			IsCustomized: false,
			PreviewURL:   catalogItem.SvgURL,
			CatalogInfo: &data.InlayCatalogInfo{
				CatalogItemID:      body.CatalogItemID,
				CustomizationNotes: body.CustomizationNotes,
			},
		}

		if err := m.Db.Inlays.Insert(&inlay); err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}

		m.WriteJSON(w, r, http.StatusCreated, inlay)
		return
	}

	// Customized catalog inlay — bake the SVG was already uploaded; we now
	// persist the inlay and a pending internal-authority proof in one tx.
	tx, err := m.Db.STDB.Begin()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	defer tx.Rollback()

	inlay := data.Inlay{
		ProjectID:    project.ID,
		Name:         body.Name,
		Type:         data.InlayTypes.Catalog,
		IsCustomized: true,
		PreviewURL:   body.Customization.BakedDesignAssetURL,
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID:      body.CatalogItemID,
			CustomizationNotes: body.CustomizationNotes,
		},
	}

	if err := m.Db.Inlays.TxInsert(tx, &inlay); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	defaultPriceGroupID := catalogItem.DefaultPriceGroupID
	colorOverrides := map[string]interface{}{}
	if body.Customization.ColorOverrides != nil {
		colorOverrides = body.Customization.ColorOverrides
	}

	proof := data.InlayProof{
		InlayID:           inlay.ID,
		VersionNumber:     1,
		DesignAssetURL:    body.Customization.BakedDesignAssetURL,
		Width:             body.Customization.Width,
		Height:            body.Customization.Height,
		PriceGroupID:      &defaultPriceGroupID,
		ScaleFactor:       body.Customization.ScaleFactor,
		ColorOverrides:    colorOverrides,
		ApprovalAuthority: data.ProofApprovalAuthorities.Internal,
		Status:            data.ProofStatuses.Pending,
		SentInChatID:      nil,
	}

	if err := m.Db.InlayProofs.TxInsert(tx, &proof); err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create customizer proof: %w", err))
		return
	}

	if err := tx.Commit(); err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.SendNotificationToAllInternalUsers(
		data.NotificationEventTypes.InternalReviewRequired,
		fmt.Sprintf("Customized inlay needs review: %s", inlay.Name),
		fmt.Sprintf("A customized inlay %q (from catalog %s) is ready for internal pricing review.", inlay.Name, catalogItem.CatalogCode),
		&project.ID, &inlay.ID,
	)

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

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only add inlays to projects in draft status, current status: %s", project.Status))
		return
	}

	inlay := data.Inlay{
		ProjectID:    project.ID,
		Name:         body.Name,
		Type:         data.InlayTypes.Custom,
		IsCustomized: false,
		PreviewURL:   "",
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

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only modify inlays on projects in draft status"))
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

	if project.Status != data.ProjectStatuses.Draft {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("can only remove inlays from projects in draft status"))
		return
	}

	err = m.Db.Inlays.Delete(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]bool{"success": true})
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
			IsCustomized:      d.Inlays.IsCustomized,
			PreviewURL:        d.Inlays.PreviewURL,
			ManufacturingStep: d.Inlays.ManufacturingStep,
		}

		if d.Inlays.ApprovedProofID != nil {
			id := int(*d.Inlays.ApprovedProofID)
			inlay.ApprovedProofID = &id
		}

		result[i] = KanbanInlay{
			Inlay:          &inlay,
			ProjectName:    d.ProjectName,
			DealershipName: d.DealershipName,
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

	// Only inlays that actually went into production (have a manufacturing
	// step) participate in project-level advancement. Inlays that the user
	// removed from the cart never enter production and would otherwise hold
	// the project back forever.
	var activeInlays []*data.Inlay
	for _, inlay := range projectInlays {
		if inlay.ManufacturingStep != nil {
			activeInlays = append(activeInlays, inlay)
		}
	}

	if len(activeInlays) == 0 {
		return
	}

	allDelivered := true
	allShipped := true
	for _, inlay := range activeInlays {
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

func (m InlayModule) HandleGetInlayMilestones(w http.ResponseWriter, r *http.Request) {
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

	milestones, err := m.Db.InlayMilestones.GetByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, milestones)
}

func (m InlayModule) HandleGetInlayUpdates(w http.ResponseWriter, r *http.Request) {
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

	updates, err := m.Db.InlayUpdates.GetByInlayID(inlay.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, updates)
}

func (m InlayModule) HandlePostInlayUpdate(w http.ResponseWriter, r *http.Request) {
	inlayUUID := r.PathValue("uuid")

	err := m.Validate.Var(inlayUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		UpdateType data.InlayUpdateType `json:"update_type" validate:"required"`
		Message    string               `json:"message" validate:"required"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.UpdateType != data.InlayUpdateTypes.Info && body.UpdateType != data.InlayUpdateTypes.Issue {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("invalid update_type: must be 'info' or 'issue'"))
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

	update := data.InlayUpdate{
		InlayID:    inlay.ID,
		UpdateType: body.UpdateType,
		Message:    body.Message,
		Step:       inlay.ManufacturingStep,
		CreatedBy:  &userID,
	}

	err = m.Db.InlayUpdates.Insert(&update)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to insert inlay update: %w", err))
		return
	}

	m.SendNotificationToAllDealershipUsersForProject(
		inlay.ProjectID,
		data.NotificationEventTypes.InlayUpdate,
		fmt.Sprintf("Update on inlay: %s", inlay.Name),
		fmt.Sprintf("New update on inlay %q: %s", inlay.Name, body.Message),
		&inlay.ID,
	)

	m.WriteJSON(w, r, http.StatusCreated, update)
}
