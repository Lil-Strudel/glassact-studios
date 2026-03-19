package data

import (
	"testing"
)

func TestInvoice_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID: project.ID,
		Status:    InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert invoice: %v", err)
	}

	if invoice.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", invoice.ID)
	}
	if invoice.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInvoice_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoiceURL := "https://invoice.example.com/inv-001"
	original := &Invoice{
		ProjectID:  project.ID,
		InvoiceURL: &invoiceURL,
		Status:     InvoiceStatuses.Sent,
	}

	err := models.Invoices.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.Invoices.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Errorf("Invoice not found")
	}
	if retrieved.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
	}
	if retrieved.InvoiceURL == nil || *retrieved.InvoiceURL != invoiceURL {
		t.Errorf("Expected invoice URL %s, got %v", invoiceURL, retrieved.InvoiceURL)
	}
}

func TestInvoice_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	original := &Invoice{
		ProjectID: project.ID,
		Status:    InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.Invoices.GetByUUID(original.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Errorf("Invoice not found by UUID")
	}
	if retrieved.UUID != original.UUID {
		t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
	}
}

func TestInvoice_GetByProjectID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID: project.ID,
		Status:    InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.Invoices.GetByProjectID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get by project ID: %v", err)
	}
	if !found {
		t.Errorf("Invoice not found by project ID")
	}
	if retrieved.ProjectID != project.ID {
		t.Errorf("Expected project ID %d, got %d", project.ID, retrieved.ProjectID)
	}
}

func TestInvoice_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	original := &Invoice{
		ProjectID: project.ID,
		Status:    InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	invoiceURL := "https://invoice.example.com/inv-updated"
	original.Status = InvoiceStatuses.Sent
	original.InvoiceURL = &invoiceURL

	err = models.Invoices.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.Invoices.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Errorf("Invoice not found after update")
	}
	if retrieved.Status != InvoiceStatuses.Sent {
		t.Errorf("Expected status Sent, got %s", retrieved.Status)
	}
	if retrieved.InvoiceURL == nil || *retrieved.InvoiceURL != invoiceURL {
		t.Errorf("Expected invoice URL to be set")
	}
}

func TestInvoice_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID: project.ID,
		Status:    InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.Invoices.Delete(invoice.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.Invoices.GetByID(invoice.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected invoice to be deleted")
	}
}

func TestInvoice_GetAll(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project1 := createTestProject(t, models, dealership.ID)
	project2 := createTestProject(t, models, dealership.ID)

	invoice1 := &Invoice{
		ProjectID: project1.ID,
		Status:    InvoiceStatuses.Draft,
	}

	invoice2 := &Invoice{
		ProjectID: project2.ID,
		Status:    InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice1)
	if err != nil {
		t.Fatalf("Failed to insert invoice1: %v", err)
	}
	err = models.Invoices.Insert(invoice2)
	if err != nil {
		t.Fatalf("Failed to insert invoice2: %v", err)
	}

	invoices, err := models.Invoices.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}
	if len(invoices) < 2 {
		t.Errorf("Expected at least 2 invoices, got %d", len(invoices))
	}
}
