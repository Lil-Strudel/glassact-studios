import type { InlayDetail, ProjectStatus } from "@glassact/data";
import type { BadgeVariant } from "@glassact/ui";
import { isManufacturingStatus } from "../../utils/project-status";

// Where an inlay is in its life. This — not proof state — decides what the inlay
// page leads with, because most inlays never see a real proof conversation:
// stock catalog items are ready on arrival and customized ones only need a
// one-shot internal pricing approval.
export type InlayPhase =
  | "configuring" // draft, not ready, nothing pending — a custom inlay waiting on a designer
  | "awaiting-approval" // a pending proof needs somebody to act
  | "ready" // draft and orderable
  | "in-production" // ordered through shipped
  | "complete" // invoiced or completed
  | "cancelled";

export function deriveInlayPhase(
  inlay: InlayDetail,
  projectStatus: ProjectStatus,
): InlayPhase {
  if (projectStatus === "cancelled") return "cancelled";
  if (projectStatus === "invoiced" || projectStatus === "completed") {
    return "complete";
  }
  if (isManufacturingStatus(projectStatus)) return "in-production";

  // Draft from here on.
  if (inlay.latest_proof?.status === "pending") return "awaiting-approval";
  if (inlay.is_ready) return "ready";
  return "configuring";
}

export const INLAY_PHASE_LABELS: Record<InlayPhase, string> = {
  configuring: "Awaiting proof",
  "awaiting-approval": "Awaiting approval",
  ready: "Ready to order",
  "in-production": "In production",
  complete: "Complete",
  cancelled: "Cancelled",
};

export function inlayPhaseBadgeVariant(phase: InlayPhase): BadgeVariant {
  switch (phase) {
    case "ready":
    case "complete":
      return "success";
    case "awaiting-approval":
      return "gilt";
    case "in-production":
      return "production";
    case "cancelled":
      return "cancelled";
    default:
      return "outline";
  }
}
