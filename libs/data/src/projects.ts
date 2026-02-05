import { StandardTable } from "./helpers";

export type ProjectStatus =
  | "draft"
  | "designing"
  | "pending-approval"
  | "approved"
  | "ordered"
  | "in-production"
  | "shipped"
  | "delivered"
  | "invoiced"
  | "completed"
  | "cancelled";

export type Project = StandardTable<{
  dealership_id: number;
  name: string;
  status: ProjectStatus;
  ordered_at: string | null;
  ordered_by: number | null;
}>;
