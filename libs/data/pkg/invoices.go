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
	ProjectID  int           `json:"project_id"`
	InvoiceURL *string       `json:"invoice_url"`
	Status     InvoiceStatus `json:"status"`
	PaidAt     *time.Time    `json:"paid_at"`
}

type InvoiceModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func invoiceFromGen(gen model.Invoices) *Invoice {
	invoice := &Invoice{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		ProjectID:  int(gen.ProjectID),
		InvoiceURL: gen.InvoiceURL,
		Status:     InvoiceStatus(gen.Status),
		PaidAt:     gen.PaidAt,
	}

	return invoice
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

	gen := &model.Invoices{
		ID:         int32(i.ID),
		UUID:       invoiceUUID,
		ProjectID:  int32(i.ProjectID),
		InvoiceURL: i.InvoiceURL,
		Status:     string(i.Status),
		PaidAt:     i.PaidAt,
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
		Version:    int32(i.Version),
	}

	return gen, nil
}

func (m InvoiceModel) Insert(invoice *Invoice) error {
	gen, err := invoiceToGen(invoice)
	if err != nil {
		return err
	}

	query := table.Invoices.INSERT(
		table.Invoices.ProjectID,
		table.Invoices.InvoiceURL,
		table.Invoices.Status,
		table.Invoices.PaidAt,
	).MODEL(gen).RETURNING(
		table.Invoices.ID,
		table.Invoices.UUID,
		table.Invoices.CreatedAt,
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

	invoice.ID = int(dest.ID)
	invoice.UUID = dest.UUID.String()
	invoice.CreatedAt = dest.CreatedAt
	invoice.UpdatedAt = dest.UpdatedAt
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
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return invoiceFromGen(dest), true, nil
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
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return invoiceFromGen(dest), true, nil
}

func (m InvoiceModel) GetActiveByProjectID(projectID int) (*Invoice, bool, error) {
	query := postgres.SELECT(
		table.Invoices.AllColumns,
	).FROM(
		table.Invoices,
	).WHERE(
		postgres.AND(
			table.Invoices.ProjectID.EQ(postgres.Int(int64(projectID))),
			table.Invoices.Status.NOT_EQ(postgres.String("void")),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Invoices
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return invoiceFromGen(dest), true, nil
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
		invoices[i] = invoiceFromGen(d)
	}

	return invoices, nil
}

func (m InvoiceModel) Update(invoice *Invoice) error {
	gen, err := invoiceToGen(invoice)
	if err != nil {
		return err
	}

	query := table.Invoices.UPDATE(
		table.Invoices.InvoiceURL,
		table.Invoices.Status,
		table.Invoices.PaidAt,
		table.Invoices.Version,
	).MODEL(gen).WHERE(
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
	return err
}
