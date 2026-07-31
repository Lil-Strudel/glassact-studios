package data

import (
	"testing"
)

func TestProjectChat_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	chat := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Hello, this is a test message",
	}

	err = models.ProjectChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert chat: %v", err)
	}

	if chat.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", chat.ID)
	}
	if chat.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if chat.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestProjectChat_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat2@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	original := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Test message for retrieval",
	}

	err = models.ProjectChats.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.ProjectChats.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Chat not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.Message != original.Message {
		t.Errorf("Expected message %s, got %s", original.Message, retrieved.Message)
	}
	if retrieved.ProjectID != original.ProjectID {
		t.Errorf("Expected project ID %d, got %d", original.ProjectID, retrieved.ProjectID)
	}
}

func TestProjectChat_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat3@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	original := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "UUID test message",
	}

	err = models.ProjectChats.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.ProjectChats.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Chat not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestProjectChat_GetByProjectID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat4@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	chat1 := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Message 1",
	}

	chat2 := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Message 2",
	}

	err = models.ProjectChats.Insert(chat1)
	if err != nil {
		t.Fatalf("Failed to insert chat1: %v", err)
	}
	err = models.ProjectChats.Insert(chat2)
	if err != nil {
		t.Fatalf("Failed to insert chat2: %v", err)
	}

	chats, err := models.ProjectChats.GetByProjectID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get by project ID: %v", err)
	}
	if len(chats) < 2 {
		t.Errorf("Expected at least 2 chats, got %d", len(chats))
	}
}

func TestProjectChat_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat5@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	original := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Original message",
	}

	err = models.ProjectChats.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.Message = "Updated message"

	err = models.ProjectChats.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.ProjectChats.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Chat not found after update")
	}
	if retrieved.Message != "Updated message" {
		t.Errorf("Expected updated message, got %s", retrieved.Message)
	}
}

func TestProjectChat_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat6@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	chat := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "Message to delete",
	}

	err = models.ProjectChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.ProjectChats.Delete(chat.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.ProjectChats.GetByID(chat.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected chat to be deleted")
	}
}

func TestProjectChat_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "chat7@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	chat1 := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "All test 1",
	}

	chat2 := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &user.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "All test 2",
	}

	err = models.ProjectChats.Insert(chat1)
	if err != nil {
		t.Fatalf("Failed to insert chat1: %v", err)
	}
	err = models.ProjectChats.Insert(chat2)
	if err != nil {
		t.Fatalf("Failed to insert chat2: %v", err)
	}

	chats, err := models.ProjectChats.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(chats) < 2 {
		t.Errorf("Expected at least 2 chats, got %d", len(chats))
	}
}

// The inlay tag is what lets a project-wide thread stay actionable: a message
// saying "this one is too dark" has to say which one.
func TestProjectChat_InlayTag_RoundTrips(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	author := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Tag Author",
		Email:        "tag-author@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}
	if err := models.DealershipUsers.Insert(author); err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}

	tagged := &ProjectChat{
		ProjectID:        project.ID,
		InlayID:          &inlay.ID,
		DealershipUserID: &author.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "this one is too dark",
	}
	if err := models.ProjectChats.Insert(tagged); err != nil {
		t.Fatalf("Failed to insert tagged chat: %v", err)
	}

	untagged := &ProjectChat{
		ProjectID:        project.ID,
		DealershipUserID: &author.ID,
		MessageType:      ChatMessageTypes.Text,
		Message:          "how is the whole order looking?",
	}
	if err := models.ProjectChats.Insert(untagged); err != nil {
		t.Fatalf("Failed to insert untagged chat: %v", err)
	}

	chats, err := models.ProjectChats.GetByProjectID(project.ID)
	if err != nil {
		t.Fatalf("GetByProjectID failed: %v", err)
	}
	if len(chats) != 2 {
		t.Fatalf("Expected 2 chats on the project thread, got %d", len(chats))
	}

	byID := make(map[int]*ProjectChat, len(chats))
	for _, c := range chats {
		byID[c.ID] = c
	}

	got := byID[tagged.ID]
	if got == nil || got.InlayID == nil {
		t.Fatalf("Expected the tagged message to keep its inlay_id")
	}
	if *got.InlayID != inlay.ID {
		t.Errorf("Expected inlay_id %d, got %d", inlay.ID, *got.InlayID)
	}

	if plain := byID[untagged.ID]; plain == nil || plain.InlayID != nil {
		t.Errorf("Expected the project-wide message to have no inlay_id")
	}
}

// Proof events now live in the project thread, so the widened CHECK has to
// accept them.
func TestProjectChat_AcceptsProofMessageTypes(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestInlay(t, models, project.ID)

	for _, messageType := range []ChatMessageType{
		ChatMessageTypes.ProofSent,
		ChatMessageTypes.ProofApproved,
		ChatMessageTypes.ProofDeclined,
	} {
		chat := &ProjectChat{
			ProjectID:   project.ID,
			InlayID:     &inlay.ID,
			MessageType: messageType,
			Message:     string(messageType),
		}
		if err := models.ProjectChats.Insert(chat); err != nil {
			t.Errorf("Failed to insert %q message: %v", messageType, err)
		}
	}
}
