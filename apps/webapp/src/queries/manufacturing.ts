import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type {
  GET,
  InlayMilestone,
  InlayUpdate,
  InlayUpdateType,
  InlayWithInfo,
  ManufacturingStep,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export type KanbanInlay = GET<InlayWithInfo> & {
  project_name: string;
  dealership_name: string;
};

export async function getKanbanInlays(): Promise<KanbanInlay[]> {
  const res = await api.get("/inlays");
  return res.data;
}

export function getKanbanInlaysOpts() {
  return queryOptions({
    queryKey: ["kanban-inlays"],
    queryFn: () => getKanbanInlays(),
  });
}

export async function patchInlayStep(params: {
  uuid: string;
  step: ManufacturingStep;
}): Promise<GET<InlayWithInfo>> {
  const res = await api.patch(`/inlay/${params.uuid}/step`, {
    step: params.step,
  });
  return res.data;
}

export function patchInlayStepOpts() {
  return mutationOptions({
    mutationFn: patchInlayStep,
  });
}

export async function getInlayMilestones(
  inlayUuid: string,
): Promise<GET<InlayMilestone>[]> {
  const res = await api.get(`/inlay/${inlayUuid}/milestones`);
  return res.data;
}

export function getInlayMilestonesOpts(inlayUuid: string) {
  return queryOptions({
    queryKey: ["inlay", inlayUuid, "milestones"],
    queryFn: () => getInlayMilestones(inlayUuid),
  });
}

export async function getInlayUpdates(
  inlayUuid: string,
): Promise<GET<InlayUpdate>[]> {
  const res = await api.get(`/inlay/${inlayUuid}/updates`);
  return res.data;
}

export function getInlayUpdatesOpts(inlayUuid: string) {
  return queryOptions({
    queryKey: ["inlay", inlayUuid, "updates"],
    queryFn: () => getInlayUpdates(inlayUuid),
  });
}

export interface PostInlayUpdateRequest {
  update_type: InlayUpdateType;
  message: string;
}

export async function postInlayUpdate(params: {
  inlayUuid: string;
  body: PostInlayUpdateRequest;
}): Promise<GET<InlayUpdate>> {
  const res = await api.post(
    `/inlay/${params.inlayUuid}/updates`,
    params.body,
  );
  return res.data;
}

export function postInlayUpdateOpts() {
  return mutationOptions({
    mutationFn: postInlayUpdate,
  });
}
