import { GET, StandardTable } from "./helpers";

export type ProjectStatus =
  | "draft"
  | "ordered"
  | "in-production"
  | "shipped"
  | "delivered"
  | "invoiced"
  | "completed"
  | "cancelled";

export const PROJECT_STATUSES: ProjectStatus[] = [
  "draft",
  "ordered",
  "in-production",
  "shipped",
  "delivered",
  "invoiced",
  "completed",
  "cancelled",
];

export type Project = StandardTable<{
  dealership_id: number;
  name: string;
  internal_reference: string | null;
  status: ProjectStatus;
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

// The project list entry. `action_summary` is present only for internal users
// and only when the project has outstanding internal action.
export type ProjectListItem = GET<Project> & {
  action_summary?: ProjectActionSummary;
};
