package invoice

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type InvoiceModule struct {
	*app.Application
}

func NewInvoiceModule(app *app.Application) *InvoiceModule {
	return &InvoiceModule{app}
}

// HandlePostProjectInvoice creates a new invoice for a delivered project and immediately
// marks it as sent by attaching the provided external invoice URL.
func (m *InvoiceModule) HandlePostProjectInvoice(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		InvoiceURL string `json:"invoice_url" validate:"required,url"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if project.Status != data.ProjectStatuses.Delivered {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("invoice can only be attached to a delivered project"))
		return
	}

	existing, found, err := m.Db.Invoices.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if found && existing.Status != data.InvoiceStatuses.Void {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("project already has an active invoice"))
		return
	}

	invoice := &data.Invoice{
		ProjectID:  project.ID,
		InvoiceURL: &body.InvoiceURL,
		Status:     data.InvoiceStatuses.Sent,
	}

	err = m.Db.Invoices.Insert(invoice)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to create invoice: %w", err))
		return
	}

	project.Status = data.ProjectStatuses.Invoiced
	err = m.Db.Projects.Update(project)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to advance project to invoiced: %w", err))
		return
	}

	projectID := project.ID
	go m.SendNotificationToAllDealershipUsersForProject(
		projectID,
		data.NotificationEventTypes.InvoiceSent,
		fmt.Sprintf("Invoice ready: %s", project.Name),
		fmt.Sprintf("An invoice has been sent for project %q. You can view it using the link provided.", project.Name),
		nil,
	)

	m.WriteJSON(w, r, http.StatusCreated, invoice)
}

// HandleGetProjectInvoice returns the invoice for a given project.
func (m *InvoiceModule) HandleGetProjectInvoice(w http.ResponseWriter, r *http.Request) {
	projectUUID := r.PathValue("uuid")

	err := m.Validate.Var(projectUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, found, err := m.Db.Projects.GetByUUID(projectUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || project.DealershipID != *dealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	invoice, found, err := m.Db.Invoices.GetByProjectID(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, invoice)
}

// HandleGetInvoice returns a single invoice by UUID.
func (m *InvoiceModule) HandleGetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceUUID := r.PathValue("uuid")

	err := m.Validate.Var(invoiceUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	invoice, found, err := m.Db.Invoices.GetByUUID(invoiceUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user := m.ContextGetUser(r)
	if user.IsDealership() {
		project, projectFound, projectErr := m.Db.Projects.GetByID(invoice.ProjectID)
		if projectErr != nil || !projectFound {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
		dealershipID := user.GetDealershipID()
		if dealershipID == nil || project.DealershipID != *dealershipID {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
	}

	m.WriteJSON(w, r, http.StatusOK, invoice)
}

// HandleMarkInvoicePaid marks an invoice as paid and advances the project to completed.
func (m *InvoiceModule) HandleMarkInvoicePaid(w http.ResponseWriter, r *http.Request) {
	invoiceUUID := r.PathValue("uuid")

	err := m.Validate.Var(invoiceUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	invoice, found, err := m.Db.Invoices.GetByUUID(invoiceUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if invoice.Status != data.InvoiceStatuses.Sent {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("only sent invoices can be marked paid"))
		return
	}

	now := time.Now()
	invoice.Status = data.InvoiceStatuses.Paid
	invoice.PaidAt = &now

	err = m.Db.Invoices.Update(invoice)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to mark invoice paid: %w", err))
		return
	}

	project, found, err := m.Db.Projects.GetByID(invoice.ProjectID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if found {
		project.Status = data.ProjectStatuses.Completed
		if updateErr := m.Db.Projects.Update(project); updateErr != nil {
			m.Log.Error("failed to advance project to completed after payment", "error", updateErr, "project_id", project.ID)
		}

		go m.SendNotificationToAllDealershipUsersForProject(
			project.ID,
			data.NotificationEventTypes.PaymentReceived,
			fmt.Sprintf("Payment received: %s", project.Name),
			fmt.Sprintf("Payment has been received for project %q. The project is now complete.", project.Name),
			nil,
		)
	}

	m.WriteJSON(w, r, http.StatusOK, invoice)
}

// HandleVoidInvoice voids an invoice.
func (m *InvoiceModule) HandleVoidInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceUUID := r.PathValue("uuid")

	err := m.Validate.Var(invoiceUUID, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	invoice, found, err := m.Db.Invoices.GetByUUID(invoiceUUID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if invoice.Status == data.InvoiceStatuses.Paid {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("paid invoices cannot be voided"))
		return
	}

	invoice.Status = data.InvoiceStatuses.Void

	err = m.Db.Invoices.Update(invoice)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to void invoice: %w", err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, invoice)
}
