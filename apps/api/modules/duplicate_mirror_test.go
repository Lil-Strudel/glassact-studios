package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func duplicateInlay(t *testing.T, ctx *testContext, token string, inlay *data.Inlay) (*testResponse, *data.Inlay) {
	t.Helper()
	resp := ctx.request(testRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("/api/inlay/%s/duplicate", inlay.UUID),
		token:  token,
	})
	if resp.statusCode != http.StatusCreated {
		return resp, nil
	}

	var created data.Inlay
	require.NoError(t, json.Unmarshal(resp.body, &created))
	reloaded, found, err := ctx.db.Inlays.GetByUUID(created.UUID)
	require.NoError(t, err)
	require.True(t, found)
	return resp, reloaded
}

// approveInlay records proofID as the inlay's approval. Inlays.Update leaves
// approved_proof_id alone, so this goes through TxUpdateFields like the
// approve handler does.
func approveInlay(t *testing.T, ctx *testContext, inlay *data.Inlay, proofID int) {
	t.Helper()
	tx, err := ctx.db.STDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback() //nolint:errcheck

	inlay.ApprovedProofID = &proofID
	require.NoError(t, ctx.db.Inlays.TxUpdateFields(tx, inlay))
	require.NoError(t, tx.Commit())
}

func TestMirrorInlay(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-BUN-0001")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Mirror")

	t.Run("mirroring an approved inlay keeps it approved", func(t *testing.T) {
		inlay, proof := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)
		approveInlay(t, ctx, inlay, proof.ID)

		resp := ctx.request(testRequest{
			method: http.MethodPatch,
			path:   fmt.Sprintf("/api/inlay/%s", inlay.UUID),
			token:  dealershipToken,
			body:   map[string]any{"mirror_x": true},
		})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		reloaded, _, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
		require.NoError(t, err)
		assert.True(t, reloaded.MirrorX)
		assert.False(t, reloaded.MirrorY)
		require.NotNil(t, reloaded.ApprovedProofID, "mirroring is an orientation, not a new design")
		assert.Equal(t, proof.ID, *reloaded.ApprovedProofID)
		assert.Equal(t, inlay.PreviewURL, reloaded.PreviewURL)
	})

	t.Run("each axis can be set without touching the other", func(t *testing.T) {
		inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Bunny")
		path := fmt.Sprintf("/api/inlay/%s", inlay.UUID)

		resp := ctx.request(testRequest{method: http.MethodPatch, path: path, token: dealershipToken, body: map[string]any{"mirror_x": true}})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))
		resp = ctx.request(testRequest{method: http.MethodPatch, path: path, token: dealershipToken, body: map[string]any{"mirror_y": true}})
		require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

		reloaded, _, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
		require.NoError(t, err)
		assert.True(t, reloaded.MirrorX)
		assert.True(t, reloaded.MirrorY)
	})

	t.Run("mirroring outside draft is rejected", func(t *testing.T) {
		_, inlay := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

		resp := ctx.request(testRequest{
			method: http.MethodPatch,
			path:   fmt.Sprintf("/api/inlay/%s", inlay.UUID),
			token:  dealershipToken,
			body:   map[string]any{"mirror_x": true},
		})
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
	})

	t.Run("another dealership cannot mirror the inlay", func(t *testing.T) {
		inlay := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Bunny")
		_, otherToken := seedOtherDealershipToken(t, ctx)

		resp := ctx.request(testRequest{
			method: http.MethodPatch,
			path:   fmt.Sprintf("/api/inlay/%s", inlay.UUID),
			token:  otherToken,
			body:   map[string]any{"mirror_x": true},
		})
		assert.Equal(t, http.StatusForbidden, resp.statusCode, string(resp.body))
	})
}

func TestDuplicateInlay(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, internalUser, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-DOV-0001")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Duplicate")

	t.Run("a stock inlay copies with its catalog item and mirror", func(t *testing.T) {
		source := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Dove")
		source.MirrorX = true
		require.NoError(t, ctx.db.Inlays.Update(source))

		resp, copied := duplicateInlay(t, ctx, dealershipToken, source)
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		assert.NotEqual(t, source.UUID, copied.UUID)
		assert.Equal(t, "Dove (copy)", copied.Name)
		assert.Equal(t, project.ID, copied.ProjectID)
		assert.True(t, copied.MirrorX)
		assert.False(t, copied.IsCustomized)
		require.NotNil(t, copied.CatalogInfo)
		assert.Equal(t, item.ID, copied.CatalogInfo.CatalogItemID)

		proofs, err := ctx.db.InlayProofs.GetByInlayID(copied.ID)
		require.NoError(t, err)
		assert.Empty(t, proofs)
	})

	t.Run("an approved customized inlay copies ready to order", func(t *testing.T) {
		source, proof := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)
		proof.Status = data.ProofStatuses.Approved
		proof.ApprovedByInternalUserID = &internalUser.ID
		require.NoError(t, ctx.db.InlayProofs.Update(proof))
		approveInlay(t, ctx, source, proof.ID)

		resp, copied := duplicateInlay(t, ctx, dealershipToken, source)
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		require.NotNil(t, copied.ApprovedProofID)
		assert.NotEqual(t, proof.ID, *copied.ApprovedProofID, "the copy gets its own proof")
		assert.True(t, copied.IsCustomized)

		copiedProof, found, err := ctx.db.InlayProofs.GetByID(*copied.ApprovedProofID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, copied.ID, copiedProof.InlayID)
		assert.Equal(t, 1, copiedProof.VersionNumber)
		assert.Equal(t, data.ProofStatuses.Approved, copiedProof.Status)
		assert.Equal(t, proof.DesignAssetURL, copiedProof.DesignAssetURL)
		require.NotNil(t, copiedProof.ApprovedByInternalUserID)
		assert.Equal(t, internalUser.ID, *copiedProof.ApprovedByInternalUserID)

		listResp := ctx.request(testRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/project/%s/inlays", project.UUID),
			token:  dealershipToken,
		})
		require.Equal(t, http.StatusOK, listResp.statusCode)
		var listed []struct {
			UUID    string `json:"uuid"`
			IsReady bool   `json:"is_ready"`
		}
		require.NoError(t, json.Unmarshal(listResp.body, &listed))
		idx := slices.IndexFunc(listed, func(i struct {
			UUID    string `json:"uuid"`
			IsReady bool   `json:"is_ready"`
		}) bool {
			return i.UUID == copied.UUID
		})
		require.GreaterOrEqual(t, idx, 0)
		assert.True(t, listed[idx].IsReady)
	})

	t.Run("a customized inlay awaiting review copies into review", func(t *testing.T) {
		source, _ := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

		resp, copied := duplicateInlay(t, ctx, dealershipToken, source)
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		assert.Nil(t, copied.ApprovedProofID)
		proofs, err := ctx.db.InlayProofs.GetByInlayID(copied.ID)
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		assert.Equal(t, data.ProofStatuses.Pending, proofs[0].Status)
		assert.Equal(t, data.ProofApprovalAuthorities.Internal, proofs[0].ApprovalAuthority)

		assert.Contains(t,
			notificationEventTypesFor(t, ctx, internalUser.ID, false),
			data.NotificationEventTypes.InternalReviewRequired)
	})

	t.Run("a custom inlay copies its description and reference images", func(t *testing.T) {
		source := &data.Inlay{
			ProjectID: project.ID,
			Name:      "Heron",
			Type:      data.InlayTypes.Custom,
			CustomInfo: &data.InlayCustomInfo{
				Description:     "A heron in reeds",
				RequestedWidth:  6,
				RequestedHeight: 9,
				ReferenceImages: []data.InlayCustomReferenceImage{
					{ImageURL: "/file/ref/a.png"},
					{ImageURL: "/file/ref/b.png"},
				},
			},
		}
		require.NoError(t, ctx.db.Inlays.Insert(source))

		resp, copied := duplicateInlay(t, ctx, dealershipToken, source)
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		require.NotNil(t, copied.CustomInfo)
		assert.Equal(t, "A heron in reeds", copied.CustomInfo.Description)
		assert.Equal(t, 6.0, copied.CustomInfo.RequestedWidth)
		require.Len(t, copied.CustomInfo.ReferenceImages, 2)
		assert.Equal(t, "/file/ref/a.png", copied.CustomInfo.ReferenceImages[0].ImageURL)
		assert.Equal(t, "/file/ref/b.png", copied.CustomInfo.ReferenceImages[1].ImageURL)
		assert.Nil(t, copied.ApprovedProofID)
	})

	t.Run("an inlay on the stone is placed beside its original", func(t *testing.T) {
		placedProject := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Placed")
		source := seedDraftCatalogInlay(t, ctx, placedProject.ID, item.ID, "Rose")

		saveResp := ctx.request(testRequest{
			method: http.MethodPut,
			path:   fmt.Sprintf("/api/project/%s/stone-design", placedProject.UUID),
			token:  dealershipToken,
			body:   stoneDesignBody(0, inlayElement("e1", source.UUID)),
		})
		require.Equal(t, http.StatusOK, saveResp.statusCode, string(saveResp.body))

		resp, copied := duplicateInlay(t, ctx, dealershipToken, source)
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		design, found, err := ctx.db.StoneDesigns.GetByProjectID(placedProject.ID)
		require.NoError(t, err)
		require.True(t, found)
		require.Len(t, design.Layout.Elements, 2)

		copyElement := design.Layout.Elements[1]
		assert.Equal(t, copied.UUID, copyElement.InlayUUID)
		assert.Equal(t, 10+data.DuplicateOffsetInches, copyElement.X)
		assert.Equal(t, 10+data.DuplicateOffsetInches, copyElement.Y)
	})

	t.Run("an inlay not on the stone leaves the design alone", func(t *testing.T) {
		source := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Lily")

		resp, _ := duplicateInlay(t, ctx, dealershipToken, source)
		require.Equal(t, http.StatusCreated, resp.statusCode, string(resp.body))

		_, found, err := ctx.db.StoneDesigns.GetByProjectID(project.ID)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("duplicating outside draft is rejected", func(t *testing.T) {
		_, source := seedOrderedProjectWithInlay(t, ctx, dealershipUser.DealershipID, item.ID)

		resp, _ := duplicateInlay(t, ctx, dealershipToken, source)
		assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))
	})

	t.Run("another dealership cannot duplicate the inlay", func(t *testing.T) {
		source := seedDraftCatalogInlay(t, ctx, project.ID, item.ID, "Dove")
		_, otherToken := seedOtherDealershipToken(t, ctx)

		resp, _ := duplicateInlay(t, ctx, otherToken, source)
		assert.Equal(t, http.StatusForbidden, resp.statusCode, string(resp.body))
	})
}
