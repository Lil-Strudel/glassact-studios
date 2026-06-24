package data

import (
	"testing"
)

func createCustomizedCatalogInlay(t *testing.T, models Models, projectID, catalogItemID int) *Inlay {
	t.Helper()

	inlay := &Inlay{
		ProjectID:    projectID,
		Name:         "Customized Catalog Inlay",
		Type:         InlayTypes.Catalog,
		IsCustomized: true,
		PreviewURL:   "https://example.com/preview.svg",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItemID,
			CustomizationNotes: "needs review",
		},
	}

	if err := models.Inlays.Insert(inlay); err != nil {
		t.Fatalf("Failed to create customized catalog inlay: %v", err)
	}

	return inlay
}

func createPendingProof(t *testing.T, models Models, inlayID, priceGroupID int, authority ProofApprovalAuthority) *InlayProof {
	t.Helper()

	proof := &InlayProof{
		InlayID:           inlayID,
		VersionNumber:     1,
		DesignAssetURL:    "https://example.com/design.svg",
		Width:             100.0,
		Height:            150.0,
		PriceGroupID:      &priceGroupID,
		ScaleFactor:       1.0,
		ColorOverrides:    map[string]interface{}{},
		ApprovalAuthority: authority,
		Status:            ProofStatuses.Pending,
	}

	if err := models.InlayProofs.Insert(proof); err != nil {
		t.Fatalf("Failed to create pending proof: %v", err)
	}

	return proof
}

func TestInlayModel_GetNeedingInternalApproval(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	// Qualifies: customized catalog inlay with a pending internal proof.
	target := createCustomizedCatalogInlay(t, models, project.ID, catalogItem.ID)
	createPendingProof(t, models, target.ID, priceGroup.ID, ProofApprovalAuthorities.Internal)

	// Does not qualify: custom inlay (different flow).
	custom := createTestInlay(t, models, project.ID)
	createPendingProof(t, models, custom.ID, priceGroup.ID, ProofApprovalAuthorities.Dealership)

	// Does not qualify: customized catalog inlay that already has an approved
	// proof (approved_proof_id is set, so it is ready).
	approved := createCustomizedCatalogInlay(t, models, project.ID, catalogItem.ID)
	approvedProof := createPendingProof(t, models, approved.ID, priceGroup.ID, ProofApprovalAuthorities.Internal)
	if _, err := testDB.STDB.Exec(
		"UPDATE inlays SET approved_proof_id = $1 WHERE id = $2",
		approvedProof.ID, approved.ID,
	); err != nil {
		t.Fatalf("Failed to set approved proof id: %v", err)
	}

	result, err := models.Inlays.GetNeedingInternalApproval()
	if err != nil {
		t.Fatalf("GetNeedingInternalApproval failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 inlay needing internal approval, got %d", len(result))
	}
	if result[0].ID != target.ID {
		t.Fatalf("expected inlay %d, got %d", target.ID, result[0].ID)
	}
}

func TestInlayModel_GetCustomNeedingProof(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)

	// Qualifies: custom inlay with no proof yet.
	target := createTestInlay(t, models, project.ID)

	// Does not qualify: custom inlay that already has a pending proof.
	withProof := createTestInlay(t, models, project.ID)
	createPendingProof(t, models, withProof.ID, priceGroup.ID, ProofApprovalAuthorities.Dealership)

	result, err := models.Inlays.GetCustomNeedingProof()
	if err != nil {
		t.Fatalf("GetCustomNeedingProof failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 custom inlay needing a proof, got %d", len(result))
	}
	if result[0].ID != target.ID {
		t.Fatalf("expected inlay %d, got %d", target.ID, result[0].ID)
	}
}

func TestProjectModel_GetActionSummaries(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	// Project A: one customized catalog inlay needing approval, one custom
	// inlay needing a proof, and an inlay whose latest chat is from the
	// dealership (awaiting reply).
	projectA := createTestProject(t, models, dealership.ID)

	approvalInlay := createCustomizedCatalogInlay(t, models, projectA.ID, catalogItem.ID)
	createPendingProof(t, models, approvalInlay.ID, priceGroup.ID, ProofApprovalAuthorities.Internal)

	createTestInlay(t, models, projectA.ID) // custom, no proof -> needs proof

	chatInlay := createTestInlay(t, models, projectA.ID)
	// System message first, then a later dealership message -> awaiting reply.
	internalChat := &InlayChat{InlayID: chatInlay.ID, MessageType: ChatMessageTypes.System, Message: "system note"}
	if err := models.InlayChats.Insert(internalChat); err != nil {
		t.Fatalf("Failed to insert system chat: %v", err)
	}
	dealershipChat := &InlayChat{
		InlayID:          chatInlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "dealership reply please",
	}
	if err := models.InlayChats.Insert(dealershipChat); err != nil {
		t.Fatalf("Failed to insert dealership chat: %v", err)
	}

	// Project B: an inlay whose latest chat is internal -> NOT awaiting reply.
	projectB := createTestProject(t, models, dealership.ID)
	bInlay := createTestInlay(t, models, projectB.ID)
	bDealershipChat := &InlayChat{
		InlayID:          bInlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "older dealership msg",
	}
	if err := models.InlayChats.Insert(bDealershipChat); err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}
	bInternalChat := &InlayChat{InlayID: bInlay.ID, MessageType: ChatMessageTypes.System, Message: "latest system note"}
	if err := models.InlayChats.Insert(bInternalChat); err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	summaries, err := models.Projects.GetActionSummaries()
	if err != nil {
		t.Fatalf("GetActionSummaries failed: %v", err)
	}

	a := summaries[projectA.ID]
	if a.NeedsInternalApproval != 1 {
		t.Errorf("project A: expected NeedsInternalApproval 1, got %d", a.NeedsInternalApproval)
	}
	if a.NeedsProof != 2 {
		// chatInlay is also a custom inlay with no proof, so it counts too.
		t.Errorf("project A: expected NeedsProof 2, got %d", a.NeedsProof)
	}
	if a.AwaitingReply != 1 {
		t.Errorf("project A: expected AwaitingReply 1, got %d", a.AwaitingReply)
	}

	b := summaries[projectB.ID]
	if b.AwaitingReply != 0 {
		t.Errorf("project B: expected AwaitingReply 0, got %d", b.AwaitingReply)
	}
	if b.NeedsProof != 1 {
		t.Errorf("project B: expected NeedsProof 1, got %d", b.NeedsProof)
	}
}
