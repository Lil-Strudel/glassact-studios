package inlay

import (
	"fmt"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// InlayCatalogItemRef is the slice of the catalog item an inlay page needs:
// enough to identify the design to a human (code, name, category), link to it,
// and resolve the customizer's color overrides against the design's manifest.
type InlayCatalogItemRef struct {
	UUID          string                 `json:"uuid"`
	CatalogCode   string                 `json:"catalog_code"`
	Name          string                 `json:"name"`
	Category      string                 `json:"category"`
	SvgURL        string                 `json:"svg_url"`
	Manifest      map[string]interface{} `json:"manifest"`
	DefaultWidth  float64                `json:"default_width"`
	DefaultHeight float64                `json:"default_height"`
}

// InlayDetail is the response shape for a single inlay. It extends the list
// shape with everything the inlay page renders on its own: which catalog design
// this came from, the proofs that matter (approved and latest), and the
// immutable order snapshot once the project has been ordered.
type InlayDetail struct {
	InlayWithProofStatus
	CatalogItem   *InlayCatalogItemRef `json:"catalog_item"`
	ApprovedProof *data.InlayProof     `json:"approved_proof"`
	LatestProof   *data.InlayProof     `json:"latest_proof"`
	OrderSnapshot *data.OrderSnapshot  `json:"order_snapshot"`
}

// applyDeleteBlockers records which dependent rows block deletion. Callers that
// already have a batched lookup pass its result in; nil means nothing blocks.
//
// A proof is reported but does not clear CanDelete — the delete removes proofs
// along with the inlay, so design work having started is no longer a reason to
// refuse.
func (i *InlayWithProofStatus) applyDeleteBlockers(blockers []data.InlayDeleteBlocker) {
	if blockers == nil {
		blockers = []data.InlayDeleteBlocker{}
	}
	i.DeleteBlockers = blockers
	i.CanDelete = len(hardDeleteBlockers(blockers)) == 0
}

// buildInlayWithProofStatus resolves the readiness and pricing fields shared by
// the list and detail endpoints.
func (m InlayModule) buildInlayWithProofStatus(inlay *data.Inlay) (InlayWithProofStatus, *data.InlayProof, error) {
	result := InlayWithProofStatus{
		Inlay:   inlay,
		IsReady: inlayIsReady(inlay),
	}

	latestProof, found, err := m.Db.InlayProofs.GetLatestByInlayID(inlay.ID)
	if err != nil {
		return InlayWithProofStatus{}, nil, fmt.Errorf("failed to load latest proof for inlay %d: %w", inlay.ID, err)
	}
	if found {
		status := string(latestProof.Status)
		result.LatestProofStatus = &status
		result.HasPendingProof = latestProof.Status == data.ProofStatuses.Pending
	} else {
		latestProof = nil
	}

	pricing, err := m.buildInlayPricing(inlay)
	if err != nil {
		return InlayWithProofStatus{}, nil, fmt.Errorf("failed to resolve pricing for inlay %d: %w", inlay.ID, err)
	}
	result.PriceGroupID = pricing.PriceGroupID
	result.PriceGroupName = pricing.PriceGroupName
	result.PriceCents = pricing.PriceCents
	result.PriceAdjustmentType = pricing.AdjustmentType
	result.PriceAdjustmentValue = pricing.AdjustmentValue

	return result, latestProof, nil
}

// buildInlayDetail assembles the full single-inlay response.
func (m InlayModule) buildInlayDetail(inlay *data.Inlay) (*InlayDetail, error) {
	base, latestProof, err := m.buildInlayWithProofStatus(inlay)
	if err != nil {
		return nil, err
	}

	blockers, err := m.Db.Inlays.GetDeleteBlockers([]int{inlay.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve delete blockers for inlay %d: %w", inlay.ID, err)
	}
	base.applyDeleteBlockers(blockers[inlay.ID])

	detail := InlayDetail{
		InlayWithProofStatus: base,
		LatestProof:          latestProof,
	}

	if inlay.CatalogInfo != nil {
		catalogItem, found, err := m.Db.CatalogItems.GetByID(inlay.CatalogInfo.CatalogItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load catalog item %d for inlay %d: %w", inlay.CatalogInfo.CatalogItemID, inlay.ID, err)
		}
		if found {
			detail.CatalogItem = &InlayCatalogItemRef{
				UUID:          catalogItem.UUID,
				CatalogCode:   catalogItem.CatalogCode,
				Name:          catalogItem.Name,
				Category:      catalogItem.Category,
				SvgURL:        catalogItem.SvgURL,
				Manifest:      catalogItem.Manifest,
				DefaultWidth:  catalogItem.DefaultWidth,
				DefaultHeight: catalogItem.DefaultHeight,
			}
		}
	}

	if inlay.ApprovedProofID != nil {
		approvedProof, found, err := m.Db.InlayProofs.GetByID(*inlay.ApprovedProofID)
		if err != nil {
			return nil, fmt.Errorf("failed to load approved proof %d for inlay %d: %w", *inlay.ApprovedProofID, inlay.ID, err)
		}
		if found {
			detail.ApprovedProof = approvedProof
		}
	}

	snapshot, found, err := m.Db.OrderSnapshots.GetByInlayID(inlay.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load order snapshot for inlay %d: %w", inlay.ID, err)
	}
	if found {
		detail.OrderSnapshot = snapshot
	}

	return &detail, nil
}
