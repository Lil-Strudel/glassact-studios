package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type ManufacturingStepCount struct {
	Step  string `json:"step"`
	Count int64  `json:"count"`
}

type DealershipDashboard struct {
	ProjectStatusCounts           []StatusCount `json:"project_status_counts"`
	PendingApprovalCount          int64         `json:"pending_approval_count"`
	OutstandingInvoiceCount       int64         `json:"outstanding_invoice_count"`
	OutstandingInvoiceAmountCents int64         `json:"outstanding_invoice_amount_cents"`
	RecentProjects                []*Project    `json:"recent_projects"`
}

type InternalDashboard struct {
	ProjectStatusCounts           []StatusCount            `json:"project_status_counts"`
	ManufacturingStepCounts       []ManufacturingStepCount `json:"manufacturing_step_counts"`
	PendingProofCount             int64                    `json:"pending_proof_count"`
	OutstandingInvoiceCount       int64                    `json:"outstanding_invoice_count"`
	OutstandingInvoiceAmountCents int64                    `json:"outstanding_invoice_amount_cents"`
	RecentProjects                []*Project               `json:"recent_projects"`
}

type DashboardModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func (m DashboardModel) GetDealershipDashboard(dealershipID int) (*DealershipDashboard, error) {
	dashboard := &DealershipDashboard{
		ProjectStatusCounts: []StatusCount{},
		RecentProjects:      []*Project{},
	}

	statusCounts, err := m.projectStatusCountsByDealership(dealershipID)
	if err != nil {
		return nil, fmt.Errorf("project status counts: %w", err)
	}
	dashboard.ProjectStatusCounts = statusCounts

	pendingApproval, err := m.pendingApprovalCountByDealership(dealershipID)
	if err != nil {
		return nil, fmt.Errorf("pending approval count: %w", err)
	}
	dashboard.PendingApprovalCount = pendingApproval

	invoiceCount, invoiceAmount, err := m.outstandingInvoicesByDealership(dealershipID)
	if err != nil {
		return nil, fmt.Errorf("outstanding invoices: %w", err)
	}
	dashboard.OutstandingInvoiceCount = invoiceCount
	dashboard.OutstandingInvoiceAmountCents = invoiceAmount

	recent, err := m.recentProjectsByDealership(dealershipID)
	if err != nil {
		return nil, fmt.Errorf("recent projects: %w", err)
	}
	dashboard.RecentProjects = recent

	return dashboard, nil
}

func (m DashboardModel) GetInternalDashboard() (*InternalDashboard, error) {
	dashboard := &InternalDashboard{
		ProjectStatusCounts:     []StatusCount{},
		ManufacturingStepCounts: []ManufacturingStepCount{},
		RecentProjects:          []*Project{},
	}

	statusCounts, err := m.projectStatusCountsGlobal()
	if err != nil {
		return nil, fmt.Errorf("project status counts: %w", err)
	}
	dashboard.ProjectStatusCounts = statusCounts

	stepCounts, err := m.manufacturingStepCountsGlobal()
	if err != nil {
		return nil, fmt.Errorf("manufacturing step counts: %w", err)
	}
	dashboard.ManufacturingStepCounts = stepCounts

	pendingProof, err := m.pendingProofCountGlobal()
	if err != nil {
		return nil, fmt.Errorf("pending proof count: %w", err)
	}
	dashboard.PendingProofCount = pendingProof

	invoiceCount, invoiceAmount, err := m.outstandingInvoicesGlobal()
	if err != nil {
		return nil, fmt.Errorf("outstanding invoices: %w", err)
	}
	dashboard.OutstandingInvoiceCount = invoiceCount
	dashboard.OutstandingInvoiceAmountCents = invoiceAmount

	recent, err := m.recentOrderedProjectsGlobal()
	if err != nil {
		return nil, fmt.Errorf("recent orders: %w", err)
	}
	dashboard.RecentProjects = recent

	return dashboard, nil
}

func (m DashboardModel) projectStatusCountsByDealership(dealershipID int) ([]StatusCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.STDB.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM projects
		WHERE dealership_id = $1
		GROUP BY status
	`, dealershipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStatusCounts(rows)
}

func (m DashboardModel) projectStatusCountsGlobal() ([]StatusCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.STDB.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM projects
		GROUP BY status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStatusCounts(rows)
}

func scanStatusCounts(rows *sql.Rows) ([]StatusCount, error) {
	counts := []StatusCount{}
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

func (m DashboardModel) manufacturingStepCountsGlobal() ([]ManufacturingStepCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.STDB.QueryContext(ctx, `
		SELECT manufacturing_step, COUNT(*) FROM inlays
		WHERE manufacturing_step IS NOT NULL
		GROUP BY manufacturing_step
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := []ManufacturingStepCount{}
	for rows.Next() {
		var sc ManufacturingStepCount
		if err := rows.Scan(&sc.Step, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

func (m DashboardModel) pendingApprovalCountByDealership(dealershipID int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int64
	err := m.STDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM inlay_proofs
		JOIN inlays ON inlays.id = inlay_proofs.inlay_id
		JOIN projects ON projects.id = inlays.project_id
		WHERE inlay_proofs.status = $1 AND projects.dealership_id = $2
	`, string(ProofStatuses.Pending), dealershipID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m DashboardModel) pendingProofCountGlobal() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int64
	err := m.STDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM inlay_proofs WHERE status = $1
	`, string(ProofStatuses.Pending)).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m DashboardModel) outstandingInvoicesByDealership(dealershipID int) (int64, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int64
	err := m.STDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM invoices
		JOIN projects ON projects.id = invoices.project_id
		WHERE invoices.status = $1 AND projects.dealership_id = $2
	`, string(InvoiceStatuses.Sent), dealershipID).Scan(&count)
	if err != nil {
		return 0, 0, err
	}

	var amount sql.NullInt64
	err = m.STDB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(order_snapshots.price_cents), 0) FROM order_snapshots
		JOIN invoices ON invoices.project_id = order_snapshots.project_id
		JOIN projects ON projects.id = invoices.project_id
		WHERE invoices.status = $1 AND projects.dealership_id = $2
	`, string(InvoiceStatuses.Sent), dealershipID).Scan(&amount)
	if err != nil {
		return 0, 0, err
	}

	return count, amount.Int64, nil
}

func (m DashboardModel) outstandingInvoicesGlobal() (int64, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int64
	err := m.STDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM invoices WHERE status = $1
	`, string(InvoiceStatuses.Sent)).Scan(&count)
	if err != nil {
		return 0, 0, err
	}

	var amount sql.NullInt64
	err = m.STDB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(order_snapshots.price_cents), 0) FROM order_snapshots
		JOIN invoices ON invoices.project_id = order_snapshots.project_id
		WHERE invoices.status = $1
	`, string(InvoiceStatuses.Sent)).Scan(&amount)
	if err != nil {
		return 0, 0, err
	}

	return count, amount.Int64, nil
}

func (m DashboardModel) recentProjectsByDealership(dealershipID int) ([]*Project, error) {
	query := postgres.SELECT(
		table.Projects.AllColumns,
	).FROM(
		table.Projects,
	).WHERE(
		table.Projects.DealershipID.EQ(postgres.Int(int64(dealershipID))),
	).ORDER_BY(
		table.Projects.UpdatedAt.DESC(),
	).LIMIT(5)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Projects
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, len(dest))
	for i, d := range dest {
		projects[i] = projectFromGen(d)
	}
	return projects, nil
}

func (m DashboardModel) recentOrderedProjectsGlobal() ([]*Project, error) {
	query := postgres.SELECT(
		table.Projects.AllColumns,
	).FROM(
		table.Projects,
	).WHERE(
		table.Projects.Status.EQ(postgres.String(string(ProjectStatuses.Ordered))),
	).ORDER_BY(
		table.Projects.OrderedAt.DESC(),
	).LIMIT(5)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Projects
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, len(dest))
	for i, d := range dest {
		projects[i] = projectFromGen(d)
	}
	return projects, nil
}
