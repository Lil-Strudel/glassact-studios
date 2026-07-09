import { GET, StandardTable } from "./helpers";

export type ProjectStatus =
  | "draft"
  | "ordered"
  | "in-production"
  | "shipped"
  | "invoiced"
  | "completed"
  | "cancelled";

export const PROJECT_STATUSES: ProjectStatus[] = [
  "draft",
  "ordered",
  "in-production",
  "shipped",
  "invoiced",
  "completed",
  "cancelled",
];

export type Project = StandardTable<{
  dealership_id: number;
  name: string;
  internal_reference: string | null;
  status: ProjectStatus;
  tracking_number: string | null;
  ordered_at: string | null;
  ordered_by: number | null;
}>;

// Per-project counts of outstanding internal actions, attached to the project
// list response for internal users only.
export type ProjectActionSummary = {
  needs_internal_approval: number;
  needs_proof: number;
  awaiting_reply: number;
};

// The project list entry. `action_summary` and `dealership_name` are present
// only for internal users; `action_summary` only when the project has
// outstanding internal action.
export type ProjectListItem = GET<Project> & {
  dealership_name?: string;
  action_summary?: ProjectActionSummary;
};

// The single-project detail response. Adds the owning dealership's name.
// `awaiting_payment` is a soft, informational signal: the owning dealership
// requires payment before shipping and there is an unpaid invoice on a project
// that has not yet shipped. It never blocks internal staff from shipping.
export type ProjectDetail = GET<Project> & {
  dealership_name?: string;
  awaiting_payment?: boolean;
};
