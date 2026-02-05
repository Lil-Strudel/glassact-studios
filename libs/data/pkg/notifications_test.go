package data

import (
	"testing"
	"time"
)

func TestNotification_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "notif@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	notification := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.ProofReady,
		Title:            "Proof Ready",
		Body:             "Your proof is ready for review",
		ProjectID:        intPtr(project.ID),
	}

	err = models.Notifications.Insert(notification)
	if err != nil {
		t.Fatalf("Failed to insert notification: %v", err)
	}

	if notification.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", notification.ID)
	}
	if notification.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if notification.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestNotification_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "notif2@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	original := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.OrderPlaced,
		Title:            "Order Placed",
		Body:             "Your order has been placed",
		ProjectID:        intPtr(project.ID),
	}

	err = models.Notifications.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.Notifications.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Notification not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.EventType != original.EventType {
		t.Errorf("Expected event type %s, got %s", original.EventType, retrieved.EventType)
	}
	if retrieved.Title != original.Title {
		t.Errorf("Expected title %s, got %s", original.Title, retrieved.Title)
	}
}

func TestNotification_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "notif3@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	original := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.InvoiceSent,
		Title:            "Invoice Sent",
		Body:             "Your invoice has been sent",
		ProjectID:        intPtr(project.ID),
	}

	err = models.Notifications.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.Notifications.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Notification not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestNotification_GetForDealershipUser(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project1 := createTestProject(t, models, dealership.ID)
	project2 := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "notif4@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	notif1 := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.ProofApproved,
		Title:            "Proof Approved",
		Body:             "Your proof has been approved",
		ProjectID:        intPtr(project1.ID),
	}

	notif2 := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.PaymentReceived,
		Title:            "Payment Received",
		Body:             "Payment has been received",
		ProjectID:        intPtr(project2.ID),
	}

	err = models.Notifications.Insert(notif1)
	if err != nil {
		t.Fatalf("Failed to insert notif1: %v", err)
	}
	err = models.Notifications.Insert(notif2)
	if err != nil {
		t.Fatalf("Failed to insert notif2: %v", err)
	}

	notifications, err := models.Notifications.GetForDealershipUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get for user: %v", err)
	}
	if len(notifications) < 2 {
		t.Errorf("Expected at least 2 notifications, got %d", len(notifications))
	}
}

func TestNotification_MarkRead(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "notif5@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	notification := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.ProofDeclined,
		Title:            "Proof Declined",
		Body:             "Your proof has been declined",
		ProjectID:        intPtr(project.ID),
	}

	err = models.Notifications.Insert(notification)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Mark as read
	err = models.Notifications.MarkRead(notification.ID)
	if err != nil {
		t.Fatalf("Failed to mark as read: %v", err)
	}

	retrieved, found, err := models.Notifications.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Notification not found after update")
	}
	if retrieved.ReadAt == nil {
		t.Errorf("Expected ReadAt to be set")
	}
}

func TestNotification_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	user := &DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        "notif6@example.com",
		Avatar:       "https://example.com/avatar.png",
		Role:         DealershipUserRoles.Admin,
		IsActive:     true,
	}

	err := models.DealershipUsers.Insert(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	notification := &Notification{
		DealershipUserID: &user.ID,
		EventType:        NotificationEventTypes.ChatMessage,
		Title:            "New Message",
		Body:             "You have a new message",
		ProjectID:        intPtr(project.ID),
	}

	err = models.Notifications.Insert(notification)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.Notifications.Delete(notification.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.Notifications.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected notification to be deleted")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr() *time.Time {
	t := time.Now()
	return &t
}
