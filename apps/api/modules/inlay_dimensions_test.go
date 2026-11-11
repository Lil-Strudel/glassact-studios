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

func seedProof(t *testing.T, ctx *testContext, inlayID, priceGroupID int, width, height float64, status data.ProofStatus) *data.InlayProof {
	t.Helper()
	proof := &data.InlayProof{
		InlayID:           inlayID,
		VersionNumber:     1,
		DesignAssetURL:    "/file/baked/proof.svg",
		Width:             width,
		Height:            height,
		PriceGroupID:      &priceGroupID,
		ScaleFactor:       1,
		ColorOverrides:    map[string]interface{}{},
		ApprovalAuthority: data.ProofApprovalAuthorities.Dealership,
		Status:            status,
	}
	require.NoError(t, ctx.db.InlayProofs.Insert(proof))
	return proof
}

func seedCustomInlay(t *testing.T, ctx *testContext, projectID int, width, height float64) *data.Inlay {
	t.Helper()
	inlay := &data.Inlay{
		ProjectID: projectID,
		Name:      "Custom",
		Type:      data.InlayTypes.Custom,
		CustomInfo: &data.InlayCustomInfo{
			Description:     "A heron",
			RequestedWidth:  width,
			RequestedHeight: height,
		},
	}
	require.NoError(t, ctx.db.Inlays.Insert(inlay))
	return inlay
}

func TestInlayList_ReportsTheSizeEachInlayWillBeMadeAt(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-DOV-0001")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Sizes")

	stock := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Stock")

	requested := seedCustomInlay(t, ctx, project.ID, 6, 9)

	pending := seedCustomInlay(t, ctx, project.ID, 6, 9)
	seedProof(t, ctx, pending.ID, priceGroup.ID, 5, 8, data.ProofStatuses.Pending)

	approved := seedCustomInlay(t, ctx, project.ID, 6, 9)
	approvedProof := seedProof(t, ctx, approved.ID, priceGroup.ID, 7, 5, data.ProofStatuses.Approved)
	approved.ApprovedProofID = &approvedProof.ID
	tx, err := ctx.db.STDB.Begin()
	require.NoError(t, err)
	require.NoError(t, ctx.db.Inlays.TxUpdateFields(tx, approved))
	require.NoError(t, tx.Commit())

	resp := ctx.request(testRequest{method: http.MethodGet, path: fmt.Sprintf("/api/project/%s/inlays", project.UUID), token: dealershipToken})
	require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

	var inlays []struct {
		UUID   string   `json:"uuid"`
		Width  *float64 `json:"width"`
		Height *float64 `json:"height"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &inlays))

	sizes := map[string][2]float64{}
	for _, inlay := range inlays {
		require.NotNil(t, inlay.Width, inlay.UUID)
		require.NotNil(t, inlay.Height, inlay.UUID)
		sizes[inlay.UUID] = [2]float64{*inlay.Width, *inlay.Height}
	}

	assert.Equal(t, [2]float64{item.DefaultWidth, item.DefaultHeight}, sizes[stock.UUID], "stock uses catalog defaults")
	assert.Equal(t, [2]float64{6, 9}, sizes[requested.UUID], "custom without a proof uses the request")
	assert.Equal(t, [2]float64{5, 8}, sizes[pending.UUID], "a pending proof beats the request")
	assert.Equal(t, [2]float64{7, 5}, sizes[approved.UUID], "the approved proof is binding")
}
