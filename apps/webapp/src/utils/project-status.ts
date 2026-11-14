import type { ProjectStatus } from "@glassact/data";
import type { BadgeVariant } from "@glassact/ui";

export const PROJECT_STATUS_LABELS: Record<ProjectStatus, string> = {
  draft: "Draft",
  ordered: "Ordered",
  "in-production": "In Production",
  shipped: "Shipped",
  invoiced: "Invoiced",
  completed: "Completed",
  cancelled: "Cancelled",
};

export const PROJECT_STATUS_TONES: Record<ProjectStatus, BadgeVariant> = {
  draft: "draft",
  ordered: "ordered",
  "in-production": "production",
  shipped: "shipped",
  invoiced: "invoiced",
  completed: "completed",
  cancelled: "cancelled",
};

// Statuses where the project's inlays are moving through the shop floor, so
// per-inlay manufacturing progress is worth showing.
const MANUFACTURING_STATUSES: ProjectStatus[] = [
  "ordered",
  "in-production",
  "shipped",
];

export function isManufacturingStatus(status: ProjectStatus): boolean {
  return MANUFACTURING_STATUSES.includes(status);
}
