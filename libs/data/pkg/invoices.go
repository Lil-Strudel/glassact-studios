package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceStatus string

type invoiceStatuses struct {
	Draft InvoiceStatus
	Sent  InvoiceStatus
	Paid  InvoiceStatus
	Void  InvoiceStatus
}

var InvoiceStatuses = invoiceStatuses{
	Draft: InvoiceStatus("draft"),
	Sent:  InvoiceStatus("sent"),
	Paid:  InvoiceStatus("paid"),
	Void:  InvoiceStatus("void"),
}

type Invoice struct {
	StandardTable
	ProjectID     int                `json:"project_id"`
	InvoiceNumber string             `json:"invoice_number"`
	SubtotalCents int                `json:"subtotal_cents"`
	TaxCents      int                `json:"tax_cents"`
	TotalCents    int                `json:"total_cents"`
	Status        InvoiceStatus      `json:"status"`
	SentAt        *time.Time         `json:"sent_at"`
	SentToEmail   *string            `json:"sent_to_email"`
	PaidAt        *time.Time         `json:"paid_at"`
	Notes         *string            `json:"notes"`
	LineItems     []*InvoiceLineItem `json:"line_items,omitempty"`
}

type InvoiceLineItem struct {
	StandardTable
	InvoiceID      int    `json:"invoice_id"`
	InlayID        *int   `json:"inlay_id"`
	Description    string `json:"description"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int    `json:"unit_price_cents"`
	TotalCents     int    `json:"total_cents"`
	SortOrder      int    `json:"sort_order"`
}

type InvoiceModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func invoiceFromGen(genInvoice model.Invoices) *Invoice {
	invoice := Invoice{
		StandardTable: StandardTable{
			ID:        int(genInvoice.ID),
			UUID:      genInvoice.UUID.String(),
			CreatedAt: genInvoice.CreatedAt,
			UpdatedAt: genInvoice.UpdatedAt,
			Version:   int(genInvoice.Version),
		},
		ProjectID:     int(genInvoice.ProjectID),
		InvoiceNumber: genInvoice.InvoiceNumber,
		SubtotalCents: int(genInvoice.SubtotalCents),
		TaxCents:      int(genInvoice.TaxCents),
		TotalCents:    int(genInvoice.TotalCents),
		Status:        InvoiceStatus(genInvoice.Status),
		SentAt:        genInvoice.SentAt,
		SentToEmail:   genInvoice.SentToEmail,
		PaidAt:        genInvoice.PaidAt,
		Notes:         genInvoice.Notes,
	}

	return &invoice
}

func invoiceToGen(i *Invoice) (*model.Invoices, error) {
	var invoiceUUID uuid.UUID
	var err error

	if i.UUID != "" {
		invoiceUUID, err = uuid.Parse(i.UUID)
		if err != nil {
			return nil, err
		}
	}

	genInvoice := model.Invoices{
		ID:            int32(i.ID),
		UUID:          invoiceUUID,
		ProjectID:     int32(i.ProjectID),
		InvoiceNumber: i.InvoiceNumber,
		SubtotalCents: int32(i.SubtotalCents),
		TaxCents:      int32(i.TaxCents),
		TotalCents:    int32(i.TotalCents),
		Status:        string(i.Status),
		SentAt:        i.SentAt,
		SentToEmail:   i.SentToEmail,
		PaidAt:        i.PaidAt,
		Notes:         i.Notes,
		UpdatedAt:     i.UpdatedAt,
		CreatedAt:     i.CreatedAt,
		Version:       int32(i.Version),
	}

	return &genInvoice, nil
}

func invoiceLineItemFromGen(genLineItem model.InvoiceLineItems) *InvoiceLineItem {
	var inlayID *int
	if genLineItem.InlayID != nil {
		inlayIDVal := int(*genLineItem.InlayID)
		inlayID = &inlayIDVal
	}

	lineItem := InvoiceLineItem{
		StandardTable: StandardTable{
			ID:        int(genLineItem.ID),
			UUID:      genLineItem.UUID.String(),
			CreatedAt: genLineItem.CreatedAt,
			UpdatedAt: genLineItem.UpdatedAt,
			Version:   int(genLineItem.Version),
		},
		InvoiceID:      int(genLineItem.InvoiceID),
		InlayID:        inlayID,
		Description:    genLineItem.Description,
		Quantity:       int(genLineItem.Quantity),
		UnitPriceCents: int(genLineItem.UnitPriceCents),
		TotalCents:     int(genLineItem.TotalCents),
		SortOrder:      int(genLineItem.SortOrder),
	}

	return &lineItem
}

func invoiceLineItemToGen(li *InvoiceLineItem) (*model.InvoiceLineItems, error) {
	var lineItemUUID uuid.UUID
	var err error

	if li.UUID != "" {
		lineItemUUID, err = uuid.Parse(li.UUID)
		if err != nil {
			return nil, err
		}
	}

	var inlayID *int32
	if li.InlayID != nil {
		inlayIDVal := int32(*li.InlayID)
		inlayID = &inlayIDVal
	}

	genLineItem := model.InvoiceLineItems{
		ID:             int32(li.ID),
		UUID:           lineItemUUID,
		InvoiceID:      int32(li.InvoiceID),
		InlayID:        inlayID,
		Description:    li.Description,
		Quantity:       int32(li.Quantity),
		UnitPriceCents: int32(li.UnitPriceCents),
		TotalCents:     int32(li.TotalCents),
		SortOrder:      int32(li.SortOrder),
		UpdatedAt:      li.UpdatedAt,
		CreatedAt:      li.CreatedAt,
		Version:        int32(li.Version),
	}

	return &genLineItem, nil
}

// Invoice Operations

func (m InvoiceModel) Insert(invoice *Invoice) error {
	genInvoice, err := invoiceToGen(invoice)
	if err != nil {
		return err
	}

	query := table.Invoices.INSERT(
		table.Invoices.ProjectID,
		table.Invoices.InvoiceNumber,
		table.Invoices.SubtotalCents,
		table.Invoices.TaxCents,
		table.Invoices.TotalCents,
		table.Invoices.Status,
		table.Invoices.SentAt,
		table.Invoices.SentToEmail,
		table.Invoices.PaidAt,
		table.Invoices.Notes,
	).MODEL(
		genInvoice,
	).RETURNING(
		table.Invoices.ID,
		table.Invoices.UUID,
		table.Invoices.UpdatedAt,
		table.Invoices.CreatedAt,
		table.Invoices.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Invoices
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	invoice.ID = int(dest.ID)
	invoice.UUID = dest.UUID.String()
	invoice.UpdatedAt = dest.UpdatedAt
	invoice.CreatedAt = dest.CreatedAt
	invoice.Version = int(dest.Version)

	return nil
}

func (m InvoiceModel) GetByID(id int) (*Invoice, bool, error) {
	query := postgres.SELECT(
		table.Invoices.AllColumns,
	).FROM(
		table.Invoices,
	).WHERE(
		table.Invoices.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Invoices
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	invoice := invoiceFromGen(dest)

	// Load line items
	lineItems, err := m.GetLineItems(int(dest.ID))
	if err != nil {
		return nil, false, err
	}
	invoice.LineItems = lineItems

	return invoice, true, nil
}

func (m InvoiceModel) GetByUUID(uuidStr string) (*Invoice, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.Invoices.AllColumns,
	).FROM(
		table.Invoices,
	).WHERE(
		table.Invoices.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Invoices
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	invoice := invoiceFromGen(dest)

	// Load line items
	lineItems, err := m.GetLineItems(int(dest.ID))
	if err != nil {
		return nil, false, err
	}
	invoice.LineItems = lineItems

	return invoice, true, nil
}

func (m InvoiceModel) GetByProjectID(projectID int) (*Invoice, bool, error) {
	query := postgres.SELECT(
		table.Invoices.AllColumns,
	).FROM(
		table.Invoices,
	).WHERE(
		table.Invoices.ProjectID.EQ(postgres.Int(int64(projectID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Invoices
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	invoice := invoiceFromGen(dest)

	// Load line items
	lineItems, err := m.GetLineItems(int(dest.ID))
	if err != nil {
		return nil, false, err
	}
	invoice.LineItems = lineItems

	return invoice, true, nil
}

func (m InvoiceModel) GetAll() ([]*Invoice, error) {
	query := postgres.SELECT(
		table.Invoices.AllColumns,
	).FROM(
		table.Invoices,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Invoices
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	invoices := make([]*Invoice, len(dest))
	for i, d := range dest {
		invoice := invoiceFromGen(d)

		// Load line items for each invoice
		lineItems, err := m.GetLineItems(int(d.ID))
		if err != nil {
			return nil, err
		}
		invoice.LineItems = lineItems

		invoices[i] = invoice
	}

	return invoices, nil
}

func (m InvoiceModel) Update(invoice *Invoice) error {
	genInvoice, err := invoiceToGen(invoice)
	if err != nil {
		return err
	}

	query := table.Invoices.UPDATE(
		table.Invoices.ProjectID,
		table.Invoices.InvoiceNumber,
		table.Invoices.SubtotalCents,
		table.Invoices.TaxCents,
		table.Invoices.TotalCents,
		table.Invoices.Status,
		table.Invoices.SentAt,
		table.Invoices.SentToEmail,
		table.Invoices.PaidAt,
		table.Invoices.Notes,
		table.Invoices.Version,
	).MODEL(
		genInvoice,
	).WHERE(
		postgres.AND(
			table.Invoices.ID.EQ(postgres.Int(int64(invoice.ID))),
			table.Invoices.Version.EQ(postgres.Int(int64(invoice.Version))),
		),
	).RETURNING(
		table.Invoices.UpdatedAt,
		table.Invoices.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Invoices
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	invoice.UpdatedAt = dest.UpdatedAt
	invoice.Version = int(dest.Version)

	return nil
}

func (m InvoiceModel) Delete(id int) error {
	query := table.Invoices.DELETE().WHERE(
		table.Invoices.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}

// Line Item Operations

func (m InvoiceModel) InsertLineItem(lineItem *InvoiceLineItem) error {
	genLineItem, err := invoiceLineItemToGen(lineItem)
	if err != nil {
		return err
	}

	query := table.InvoiceLineItems.INSERT(
		table.InvoiceLineItems.InvoiceID,
		table.InvoiceLineItems.InlayID,
		table.InvoiceLineItems.Description,
		table.InvoiceLineItems.Quantity,
		table.InvoiceLineItems.UnitPriceCents,
		table.InvoiceLineItems.TotalCents,
		table.InvoiceLineItems.SortOrder,
	).MODEL(
		genLineItem,
	).RETURNING(
		table.InvoiceLineItems.ID,
		table.InvoiceLineItems.UUID,
		table.InvoiceLineItems.UpdatedAt,
		table.InvoiceLineItems.CreatedAt,
		table.InvoiceLineItems.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InvoiceLineItems
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	lineItem.ID = int(dest.ID)
	lineItem.UUID = dest.UUID.String()
	lineItem.UpdatedAt = dest.UpdatedAt
	lineItem.CreatedAt = dest.CreatedAt
	lineItem.Version = int(dest.Version)

	return nil
}

func (m InvoiceModel) GetLineItems(invoiceID int) ([]*InvoiceLineItem, error) {
	query := postgres.SELECT(
		table.InvoiceLineItems.AllColumns,
	).FROM(
		table.InvoiceLineItems,
	).WHERE(
		table.InvoiceLineItems.InvoiceID.EQ(postgres.Int(int64(invoiceID))),
	).ORDER_BY(
		table.InvoiceLineItems.SortOrder.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InvoiceLineItems
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	lineItems := make([]*InvoiceLineItem, len(dest))
	for i, d := range dest {
		lineItems[i] = invoiceLineItemFromGen(d)
	}

	return lineItems, nil
}

func (m InvoiceModel) UpdateLineItem(lineItem *InvoiceLineItem) error {
	genLineItem, err := invoiceLineItemToGen(lineItem)
	if err != nil {
		return err
	}

	query := table.InvoiceLineItems.UPDATE(
		table.InvoiceLineItems.InlayID,
		table.InvoiceLineItems.Description,
		table.InvoiceLineItems.Quantity,
		table.InvoiceLineItems.UnitPriceCents,
		table.InvoiceLineItems.TotalCents,
		table.InvoiceLineItems.SortOrder,
		table.InvoiceLineItems.Version,
	).MODEL(
		genLineItem,
	).WHERE(
		postgres.AND(
			table.InvoiceLineItems.ID.EQ(postgres.Int(int64(lineItem.ID))),
			table.InvoiceLineItems.Version.EQ(postgres.Int(int64(lineItem.Version))),
		),
	).RETURNING(
		table.InvoiceLineItems.UpdatedAt,
		table.InvoiceLineItems.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InvoiceLineItems
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	lineItem.UpdatedAt = dest.UpdatedAt
	lineItem.Version = int(dest.Version)

	return nil
}

func (m InvoiceModel) DeleteLineItem(id int) error {
	query := table.InvoiceLineItems.DELETE().WHERE(
		table.InvoiceLineItems.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
