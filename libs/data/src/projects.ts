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
  // One installation kit covers every inlay on the project.
  // `installation_kit` is the draft-time choice; `installation_kit_price_cents`
  // is null until the order is placed, then locks the charge.
  installation_kit: boolean;
  installation_kit_price_cents: number | null;
}>;

// Per-project counts of outstanding internal actions, attached to the project
// list response for internal users only.
// `awaiting_reply` is a yes/no: chat is one thread per project, so there is
// nothing to count.
export type ProjectActionSummary = {
  needs_internal_approval: number;
  needs_proof: number;
  awaiting_reply: boolean;
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
// `is_watching` is the requesting user's own subscription state; `watcher_count`
// counts every active watcher on both sides of the project.
export type ProjectDetail = GET<Project> & {
  dealership_name?: string;
  awaiting_payment?: boolean;
  is_watching: boolean;
  watcher_count: number;
};
