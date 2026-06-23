import { StandardTable } from "./helpers";

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
