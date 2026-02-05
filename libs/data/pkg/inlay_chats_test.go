package data

import (
	"testing"
)

func TestInlayChat_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}

	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	chat := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Test message",
	}

	err = models.InlayChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	if chat.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", chat.ID)
	}
	if chat.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInlayChat_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	chat := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Test message",
	}
	err = models.InlayChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	retrieved, found, err := models.InlayChats.GetByID(chat.ID)
	if err != nil {
		t.Fatalf("Failed to get chat: %v", err)
	}

	if !found {
		t.Errorf("Expected chat to be found")
	}

	if retrieved.ID != chat.ID {
		t.Errorf("Expected ID %d, got %d", chat.ID, retrieved.ID)
	}
	if retrieved.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", retrieved.Message)
	}
}

func TestInlayChat_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	chat := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Test message",
	}
	err = models.InlayChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	retrieved, found, err := models.InlayChats.GetByUUID(chat.UUID)
	if err != nil {
		t.Fatalf("Failed to get chat by UUID: %v", err)
	}

	if !found {
		t.Errorf("Expected chat to be found")
	}

	if retrieved.UUID != chat.UUID {
		t.Errorf("Expected UUID %s, got %s", chat.UUID, retrieved.UUID)
	}
}

func TestInlayChat_GetByInlayID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	chat1 := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Message 1",
	}
	err = models.InlayChats.Insert(chat1)
	if err != nil {
		t.Fatalf("Failed to insert chat 1: %v", err)
	}

	chat2 := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Message 2",
	}
	err = models.InlayChats.Insert(chat2)
	if err != nil {
		t.Fatalf("Failed to insert chat 2: %v", err)
	}

	chats, err := models.InlayChats.GetByInlayID(inlay.ID)
	if err != nil {
		t.Fatalf("Failed to get chats by inlay ID: %v", err)
	}

	if len(chats) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(chats))
	}
}

func TestInlayChat_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	chat := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Test message",
	}
	err = models.InlayChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	chat.Message = "Updated message"
	err = models.InlayChats.Update(chat)
	if err != nil {
		t.Fatalf("Failed to update chat: %v", err)
	}

	retrieved, _, err := models.InlayChats.GetByID(chat.ID)
	if err != nil {
		t.Fatalf("Failed to get chat: %v", err)
	}

	if retrieved.Message != "Updated message" {
		t.Errorf("Expected message to be updated, got %s", retrieved.Message)
	}
}

func TestInlayChat_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	priceGroup := createTestPriceGroup(t, models)
	catalogItem := createTestCatalogItem(t, models, priceGroup.ID)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Test Inlay",
		Type:       InlayTypes.Catalog,
		PreviewURL: "https://example.com/preview.png",
		CatalogInfo: &InlayCatalogInfo{
			CatalogItemID:      catalogItem.ID,
			CustomizationNotes: "Test notes",
		},
	}
	err := models.Inlays.Insert(inlay)
	if err != nil {
		t.Fatalf("Failed to insert inlay: %v", err)
	}

	chat := &InlayChat{
		InlayID:          inlay.ID,
		DealershipUserID: &dealershipUser.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Test message",
	}
	err = models.InlayChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	err = models.InlayChats.Delete(chat.ID)
	if err != nil {
		t.Fatalf("Failed to delete chat: %v", err)
	}

	_, found, err := models.InlayChats.GetByID(chat.ID)
	if err != nil {
		t.Fatalf("Failed to get chat: %v", err)
	}

	if found {
		t.Errorf("Expected chat to be deleted")
	}
}
