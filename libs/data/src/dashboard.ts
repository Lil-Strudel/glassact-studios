import type { Project } from "./projects";
import type { GET } from "./helpers";

export type StatusCount = {
  status: string;
  count: number;
};

export type ManufacturingStepCount = {
  step: string;
  count: number;
};

export type DealershipDashboard = {
  project_status_counts: StatusCount[];
  pending_approval_count: number;
  outstanding_invoice_count: number;
  outstanding_invoice_amount_cents: number;
  recent_projects: GET<Project>[];
};

export type InternalDashboard = {
  project_status_counts: StatusCount[];
  manufacturing_step_counts: ManufacturingStepCount[];
  active_blocker_count: number;
  hard_blocker_count: number;
  pending_proof_count: number;
  outstanding_invoice_count: number;
  outstanding_invoice_amount_cents: number;
  recent_projects: GET<Project>[];
};
