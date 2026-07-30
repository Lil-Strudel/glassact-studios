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

// A proof used to block removal outright, leaving the dealership stuck with an
// inlay it no longer wanted on a project it had not ordered yet.
func TestDeleteInlay_WithProofOnDraftProject_RemovesBoth(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-DEL-0001")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Delete Project")

	inlay, proof := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

	inlay.ApprovedProofID = &proof.ID
	require.NoError(t, ctx.db.Inlays.Update(inlay))

	resp := ctx.request(testRequest{
		method: http.MethodDelete,
		path:   fmt.Sprintf("/api/inlay/%s", inlay.UUID),
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

	_, found, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
	require.NoError(t, err)
	assert.False(t, found, "inlay should be gone")

	proofs, err := ctx.db.InlayProofs.GetByInlayID(inlay.ID)
	require.NoError(t, err)
	assert.Empty(t, proofs, "proofs should be removed with the inlay")
}

func TestDeleteInlay_ReportsProofWithoutBlocking(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-DEL-0002")
	project := seedDraftProject(t, ctx, dealershipUser.DealershipID, "Blocker Project")

	seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

	resp := ctx.request(testRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/project/%s/inlays", project.UUID),
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, resp.statusCode, string(resp.body))

	var inlays []struct {
		CanDelete      bool     `json:"can_delete"`
		DeleteBlockers []string `json:"delete_blockers"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &inlays))
	require.Len(t, inlays, 1)

	assert.True(t, inlays[0].CanDelete, "a proof alone should not block removal")
	assert.Contains(t, inlays[0].DeleteBlockers, "proof", "the proof is still reported")
}

func TestDeleteInlay_RejectedWhenProjectNotDraft(t *testing.T) {
	ctx, teardown := setupTestApp(t)
	defer teardown()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, ctx)
	priceGroup := seedPriceGroup(t, ctx, "Standard")
	item := seedCatalogItem(t, ctx, priceGroup.ID, "A-DEL-0003")

	project := &data.Project{
		Name:         "Ordered Project",
		Status:       data.ProjectStatuses.Ordered,
		DealershipID: dealershipUser.DealershipID,
	}
	require.NoError(t, ctx.db.Projects.Insert(project))

	inlay, _ := seedCustomizedCatalogInlay(t, ctx, project.ID, item.ID, priceGroup.ID)

	resp := ctx.request(testRequest{
		method: http.MethodDelete,
		path:   fmt.Sprintf("/api/inlay/%s", inlay.UUID),
		token:  dealershipToken,
	})
	assert.Equal(t, http.StatusBadRequest, resp.statusCode, string(resp.body))

	_, found, err := ctx.db.Inlays.GetByUUID(inlay.UUID)
	require.NoError(t, err)
	assert.True(t, found, "inlay should survive a rejected delete")
}
