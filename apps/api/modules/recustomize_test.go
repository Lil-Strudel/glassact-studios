package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedCustomizedCatalogInlay mirrors what POST /project/{uuid}/inlays/catalog
// produces for a customizer-baked inlay: the inlay plus a pending
// internal-authority proof at v1.
func seedCustomizedCatalogInlay(
	t *testing.T,
	ctx *testContext,
	projectID, catalogItemID, priceGroupID int,
) (*data.Inlay, *data.InlayProof) {
	inlay := &data.Inlay{
		ProjectID:    projectID,
		Name:         "Customized Dove",
		Type:         data.InlayTypes.Catalog,
		IsCustomized: true,
		PreviewURL:   "/file/baked/v1.svg",
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID:      catalogItemID,
			CustomizationNotes: "",
		},
	}
	require.NoError(t, ctx.db.Inlays.Insert(inlay))

	proof := &data.InlayProof{
		InlayID:           inlay.ID,
		VersionNumber:     1,
		DesignAssetURL:    "/file/baked/v1.svg",
		Width:             10,
		Height:            10,
		PriceGroupID:      &priceGroupID,
		ScaleFactor:       1,
		ColorOverrides:    map[string]interface{}{},
		ApprovalAuthority: data.ProofApprovalAuthorities.Internal,
		Status:            data.ProofStatuses.Pending,
	}
	require.NoError(t, ctx.db.InlayProofs.Insert(proof))

	return inlay, proof
}

func recustomizeBody() map[string]any {
	return map[string]any{
		"baked_design_asset_url": "/file/baked/v2.svg",
		"scale_factor":           1.5,
		"width":                  12.0,
		"height":                 12.0,
		"color_overrides": map[string]any{
			"groups": map[string]any{
				"group-0": map[string]any{"glass_color_id": 3},
			},
		},
	}
}

func seedDraftProject(t *testing.T, ctx *testContext, dealershipID int, name string) *data.Project {
	project := &data.Project{
		Name:         name,
		Status:       data.ProjectStatuses.Draft,
		DealershipID: dealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))
	return project
}

func TestRecustomizeInlay_SupersedesPreviousProofAndResetsApproval(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-RC-0001")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Recustomize Project")

	inlay, firstProof := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

	// The first proof was already approved, so re-customizing has to undo that.
	inlay.ApprovedProofID = &firstProof.ID
	require.NoError(t, ctx.db.Inlays.Update(inlay))

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/recustomize", inlay.UUID),
		token:  dealershipToken,
		body:   recustomizeBody(),
	})
	require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

	proofs, err := ctx.db.InlayProofs.GetByInlayID(inlay.ID)
	require.NoError(t, err)
	require.Len(t, proofs, 2)

	byVersion := map[int]*data.InlayProof{}
	for _, proof := range proofs {
		byVersion[proof.VersionNumber] = proof
	}

	require.Contains(t, byVersion, 1)
	require.Contains(t, byVersion, 2)

	assert.Equal(t, data.ProofStatuses.Superseded, byVersion[1].Status)

	newProof := byVersion[2]
	assert.Equal(t, data.ProofStatuses.Pending, newProof.Status)
	assert.Equal(t, data.ProofApprovalAuthorities.Internal, newProof.ApprovalAuthority)
	assert.Equal(t, "/file/baked/v2.svg", newProof.DesignAssetURL)
	assert.Equal(t, 12.0, newProof.Width)
	assert.Equal(t, 1.5, newProof.ScaleFactor)
	require.NotNil(t, newProof.PriceGroupID)
	assert.Equal(t, priceGroup.ID, *newProof.PriceGroupID)

	reloaded, found, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Nil(t, reloaded.ApprovedProofID, "re-coloring must send the inlay back for pricing review")
	assert.Equal(t, "/file/baked/v2.svg", reloaded.PreviewURL)

	// The response is the enriched detail shape the inlay page consumes.
	var detail struct {
		IsReady     bool `json:"is_ready"`
		LatestProof *struct {
			VersionNumber int `json:"version_number"`
		} `json:"latest_proof"`
		CatalogItem *struct {
			CatalogCode string `json:"catalog_code"`
		} `json:"catalog_item"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &detail))
	assert.False(t, detail.IsReady)
	require.NotNil(t, detail.LatestProof)
	assert.Equal(t, 2, detail.LatestProof.VersionNumber)
	require.NotNil(t, detail.CatalogItem)
	assert.Equal(t, "A-RC-0001", detail.CatalogItem.CatalogCode)
}

func TestRecustomizeInlay_RejectedWhenProjectNotDraft(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-RC-0002")

	project := &data.Project{
		Name:         "Ordered Project",
		Status:       data.ProjectStatuses.Ordered,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	inlay, _ := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/recustomize", inlay.UUID),
		token:  dealershipToken,
		body:   recustomizeBody(),
	})
	assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
}

// Customizing a stock inlay makes it customized, which drops it out of "ready"
// until internal prices the coloring — otherwise inlayIsReady would keep
// treating it as an uncustomized catalog item and let it be ordered unpriced.
func TestRecustomizeInlay_OnStockCatalogInlay_MarksCustomizedAndNotReady(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-RC-0003")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Stock Project")

	inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Stock Dove")

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/recustomize", inlay.UUID),
		token:  dealershipToken,
		body:   recustomizeBody(),
	})
	require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

	updated, found, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, updated.IsCustomized)
	assert.Nil(t, updated.ApprovedProofID)

	proofs, err := ctx.db.InlayProofs.GetByInlayID(inlay.ID)
	require.NoError(t, err)
	require.Len(t, proofs, 1, "first customization starts proof versioning at v1")
	assert.Equal(t, 1, proofs[0].VersionNumber)
	assert.Equal(t, data.ProofStatuses.Pending, proofs[0].Status)
	assert.Equal(t, data.ProofApprovalAuthorities.Internal, proofs[0].ApprovalAuthority)
}

func TestRecustomizeInlay_RejectedForCustomInlay(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Custom Project")

	inlay := &data.Inlay{
		ProjectID: project.ID,
		Name:      "Hand-drawn Rose",
		Type:      data.InlayTypes.Custom,
		CustomInfo: &data.InlayCustomInfo{
			Description: "A rose, from the sketch attached.",
		},
	}
	require.NoError(t, ctx.db.Inlays.Insert(inlay))

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/recustomize", inlay.UUID),
		token:  dealershipToken,
		body:   recustomizeBody(),
	})
	assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
}

func TestRecustomizeInlay_OtherDealershipForbidden(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, _, _, _ := seedTestData(t, ctx)
	_, otherToken, _, _ := seedTestData(t, ctx)

	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-RC-0004")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Owned Project")

	inlay, _ := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/recustomize", inlay.UUID),
		token:  otherToken,
		body:   recustomizeBody(),
	})
	assert.NotEqual(t, http.StatusCreated, resp.statusCode, string(resp.body))
	assert.Contains(
		t,
		[]int{http.StatusForbidden, http.StatusNotFound},
		resp.statusCode,
		string(resp.body),
	)

	proofs, err := ctx.db.InlayProofs.GetByInlayID(inlay.ID)
	require.NoError(t, err)
	assert.Len(t, proofs, 1, "a foreign dealership must not be able to add a proof")
}
