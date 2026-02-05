import { StandardTable } from "./helpers";
import { ManufacturingStep } from "./inlays";

export type MilestoneEventType = "entered" | "exited" | "reverted";

export type InlayMilestone = StandardTable<{
  inlay_id: number;
  step: ManufacturingStep;
  event_type: MilestoneEventType;
  performed_by: number;
  event_time: string;
}>;
