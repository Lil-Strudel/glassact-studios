package data

import (
	"testing"
	"time"
)

func TestInvoice_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-001",
		SubtotalCents: 50000,
		TaxCents:      5000,
		TotalCents:    55000,
		Status:        InvoiceStatuses.Draft,
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

	original := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-002",
		SubtotalCents: 100000,
		TaxCents:      10000,
		TotalCents:    110000,
		Status:        InvoiceStatuses.Draft,
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
	if retrieved.InvoiceNumber != original.InvoiceNumber {
		t.Errorf("Expected invoice number %s, got %s", original.InvoiceNumber, retrieved.InvoiceNumber)
	}
	if retrieved.TotalCents != original.TotalCents {
		t.Errorf("Expected total %d, got %d", original.TotalCents, retrieved.TotalCents)
	}
}

func TestInvoice_GetByUUID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	original := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-003",
		SubtotalCents: 75000,
		TaxCents:      7500,
		TotalCents:    82500,
		Status:        InvoiceStatuses.Draft,
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
		ProjectID:     project.ID,
		InvoiceNumber: "INV-004",
		SubtotalCents: 50000,
		TaxCents:      5000,
		TotalCents:    55000,
		Status:        InvoiceStatuses.Draft,
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
		ProjectID:     project.ID,
		InvoiceNumber: "INV-005",
		SubtotalCents: 50000,
		TaxCents:      5000,
		TotalCents:    55000,
		Status:        InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.Status = InvoiceStatuses.Sent
	original.SentToEmail = stringPtr("test@example.com")
	original.SentAt = &time.Time{}
	*original.SentAt = time.Now()

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
	if retrieved.SentToEmail == nil {
		t.Errorf("Expected SentToEmail to be set")
	}
}

func TestInvoice_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-006",
		SubtotalCents: 50000,
		TaxCents:      5000,
		TotalCents:    55000,
		Status:        InvoiceStatuses.Draft,
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
		ProjectID:     project1.ID,
		InvoiceNumber: "INV-007",
		SubtotalCents: 50000,
		TaxCents:      5000,
		TotalCents:    55000,
		Status:        InvoiceStatuses.Draft,
	}

	invoice2 := &Invoice{
		ProjectID:     project2.ID,
		InvoiceNumber: "INV-008",
		SubtotalCents: 100000,
		TaxCents:      10000,
		TotalCents:    110000,
		Status:        InvoiceStatuses.Draft,
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

func TestInvoiceLineItem_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-009",
		SubtotalCents: 100000,
		TaxCents:      10000,
		TotalCents:    110000,
		Status:        InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert invoice: %v", err)
	}

	lineItem := &InvoiceLineItem{
		InvoiceID:      invoice.ID,
		Description:    "Glass Cut 1",
		Quantity:       1,
		UnitPriceCents: 100000,
		TotalCents:     100000,
		SortOrder:      1,
	}

	err = models.Invoices.InsertLineItem(lineItem)
	if err != nil {
		t.Fatalf("Failed to insert line item: %v", err)
	}

	if lineItem.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", lineItem.ID)
	}
	if lineItem.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
}

func TestInvoiceLineItem_GetLineItems(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-010",
		SubtotalCents: 200000,
		TaxCents:      20000,
		TotalCents:    220000,
		Status:        InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert invoice: %v", err)
	}

	lineItem1 := &InvoiceLineItem{
		InvoiceID:      invoice.ID,
		Description:    "Item 1",
		Quantity:       1,
		UnitPriceCents: 100000,
		TotalCents:     100000,
		SortOrder:      1,
	}

	lineItem2 := &InvoiceLineItem{
		InvoiceID:      invoice.ID,
		Description:    "Item 2",
		Quantity:       1,
		UnitPriceCents: 100000,
		TotalCents:     100000,
		SortOrder:      2,
	}

	err = models.Invoices.InsertLineItem(lineItem1)
	if err != nil {
		t.Fatalf("Failed to insert line item 1: %v", err)
	}
	err = models.Invoices.InsertLineItem(lineItem2)
	if err != nil {
		t.Fatalf("Failed to insert line item 2: %v", err)
	}

	lineItems, err := models.Invoices.GetLineItems(invoice.ID)
	if err != nil {
		t.Fatalf("Failed to get line items: %v", err)
	}
	if len(lineItems) != 2 {
		t.Errorf("Expected 2 line items, got %d", len(lineItems))
	}
	if lineItems[0].SortOrder != 1 || lineItems[1].SortOrder != 2 {
		t.Errorf("Line items not in correct sort order")
	}
}

func TestInvoiceLineItem_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-011",
		SubtotalCents: 100000,
		TaxCents:      10000,
		TotalCents:    110000,
		Status:        InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert invoice: %v", err)
	}

	lineItem := &InvoiceLineItem{
		InvoiceID:      invoice.ID,
		Description:    "Original Description",
		Quantity:       1,
		UnitPriceCents: 100000,
		TotalCents:     100000,
		SortOrder:      1,
	}

	err = models.Invoices.InsertLineItem(lineItem)
	if err != nil {
		t.Fatalf("Failed to insert line item: %v", err)
	}

	lineItem.Description = "Updated Description"
	lineItem.Quantity = 2
	lineItem.UnitPriceCents = 50000
	lineItem.TotalCents = 100000

	err = models.Invoices.UpdateLineItem(lineItem)
	if err != nil {
		t.Fatalf("Failed to update line item: %v", err)
	}

	lineItems, err := models.Invoices.GetLineItems(invoice.ID)
	if err != nil {
		t.Fatalf("Failed to get line items: %v", err)
	}
	if lineItems[0].Description != "Updated Description" {
		t.Errorf("Expected updated description, got %s", lineItems[0].Description)
	}
	if lineItems[0].Quantity != 2 {
		t.Errorf("Expected quantity 2, got %d", lineItems[0].Quantity)
	}
}

func TestInvoiceLineItem_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	invoice := &Invoice{
		ProjectID:     project.ID,
		InvoiceNumber: "INV-012",
		SubtotalCents: 100000,
		TaxCents:      10000,
		TotalCents:    110000,
		Status:        InvoiceStatuses.Draft,
	}

	err := models.Invoices.Insert(invoice)
	if err != nil {
		t.Fatalf("Failed to insert invoice: %v", err)
	}

	lineItem := &InvoiceLineItem{
		InvoiceID:      invoice.ID,
		Description:    "Item to Delete",
		Quantity:       1,
		UnitPriceCents: 100000,
		TotalCents:     100000,
		SortOrder:      1,
	}

	err = models.Invoices.InsertLineItem(lineItem)
	if err != nil {
		t.Fatalf("Failed to insert line item: %v", err)
	}

	err = models.Invoices.DeleteLineItem(lineItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete line item: %v", err)
	}

	lineItems, err := models.Invoices.GetLineItems(invoice.ID)
	if err != nil {
		t.Fatalf("Failed to get line items: %v", err)
	}
	if len(lineItems) != 0 {
		t.Errorf("Expected 0 line items after deletion, got %d", len(lineItems))
	}
}
