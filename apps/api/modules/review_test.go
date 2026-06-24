package modules

import (
	"encoding/json"
	"net/http"
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReviewQueueEndpoint(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, internalToken := seedTestData(t, testCtx)
	db := testCtx.db

	priceGroup := &data.PriceGroup{Name: "PG", BasePriceCents: 10000, IsActive: true}
	require.NoError(t, db.PriceGroups.Insert(priceGroup))

	catalogItem := &data.CatalogItem{
		CatalogCode:         "RQ-ITEM-001",
		Name:                "Review Queue Item",
		Category:            "test",
		DefaultWidth:        10.0,
		DefaultHeight:       15.0,
		MinWidth:            5.0,
		MinHeight:           8.0,
		DefaultPriceGroupID: priceGroup.ID,
		SvgURL:              "https://example.com/item.svg",
		IsActive:            true,
	}
	require.NoError(t, db.CatalogItems.Insert(catalogItem))

	project := &data.Project{
		DealershipID: dealershipUser.DealershipID,
		Name:         "Review Queue Project",
		Status:       data.ProjectStatuses.Draft,
	}
	require.NoError(t, db.Projects.Insert(project))

	// Customized catalog inlay with a pending internal proof -> needs approval.
	approvalInlay := &data.Inlay{
		ProjectID:    project.ID,
		Name:         "Needs Approval",
		Type:         data.InlayTypes.Catalog,
		IsCustomized: true,
		PreviewURL:   "https://example.com/preview.svg",
		CatalogInfo: &data.InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "review me",
		},
	}
	require.NoError(t, db.Inlays.Insert(approvalInlay))

	proof := &data.InlayProof{
		InlayID:           approvalInlay.ID,
		VersionNumber:     1,
		DesignAssetURL:    "https://example.com/design.svg",
		Width:             10.0,
		Height:            15.0,
		PriceGroupID:      &priceGroup.ID,
		ScaleFactor:       1.0,
		ColorOverrides:    map[string]interface{}{},
		ApprovalAuthority: data.ProofApprovalAuthorities.Internal,
		Status:            data.ProofStatuses.Pending,
	}
	require.NoError(t, db.InlayProofs.Insert(proof))

	// Custom inlay with no proof -> needs proof.
	customInlay := &data.Inlay{
		ProjectID:  project.ID,
		Name:       "Needs Proof",
		Type:       data.InlayTypes.Custom,
		PreviewURL: "https://example.com/preview.svg",
		CustomInfo: &data.InlayCustomInfo{
			Description:     "custom design",
			RequestedWidth:  100.0,
			RequestedHeight: 150.0,
		},
	}
	require.NoError(t, db.Inlays.Insert(customInlay))

	t.Run("internal user gets the queue", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "GET",
			path:   "/api/review-queue",
			token:  internalToken,
		})

		require.Equal(t, http.StatusOK, resp.statusCode)

		var queue struct {
			NeedsApproval []struct {
				ProjectUUID  string          `json:"project_uuid"`
				ProjectName  string          `json:"project_name"`
				Inlay        json.RawMessage `json:"inlay"`
				PendingProof json.RawMessage `json:"pending_proof"`
			} `json:"needs_approval"`
			NeedsProof []struct {
				ProjectUUID string `json:"project_uuid"`
			} `json:"needs_proof"`
		}
		require.NoError(t, json.Unmarshal(resp.body, &queue))

		require.Len(t, queue.NeedsApproval, 1)
		assert.Equal(t, project.UUID, queue.NeedsApproval[0].ProjectUUID)
		assert.Equal(t, project.Name, queue.NeedsApproval[0].ProjectName)
		assert.NotEmpty(t, queue.NeedsApproval[0].PendingProof)

		require.Len(t, queue.NeedsProof, 1)
		assert.Equal(t, project.UUID, queue.NeedsProof[0].ProjectUUID)
	})

	t.Run("dealership user is forbidden", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "GET",
			path:   "/api/review-queue",
			token:  dealershipToken,
		})
		assert.Equal(t, http.StatusForbidden, resp.statusCode)
	})

	t.Run("unauthenticated is rejected", func(t *testing.T) {
		resp := testCtx.request(testRequest{
			method: "GET",
			path:   "/api/review-queue",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.statusCode)
	})
}
