import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, InlayBlocker, InlayWithInfo, ManufacturingStep } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export type KanbanInlay = GET<InlayWithInfo> & {
  project_name: string;
  dealership_name: string;
  has_hard_blocker: boolean;
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

export async function getBlockersByInlay(
  inlayUuid: string,
): Promise<GET<InlayBlocker>[]> {
  const res = await api.get(`/inlay/${inlayUuid}/blockers`);
  return res.data;
}

export function getBlockersByInlayOpts(inlayUuid: string) {
  return queryOptions({
    queryKey: ["inlay", inlayUuid, "blockers"],
    queryFn: () => getBlockersByInlay(inlayUuid),
  });
}

export interface PostBlockerRequest {
  blocker_type: "soft" | "hard";
  reason: string;
  step_blocked: string;
}

export async function postBlocker(params: {
  inlayUuid: string;
  body: PostBlockerRequest;
}): Promise<GET<InlayBlocker>> {
  const res = await api.post(`/inlay/${params.inlayUuid}/blockers`, params.body);
  return res.data;
}

export function postBlockerOpts() {
  return mutationOptions({
    mutationFn: postBlocker,
  });
}

export interface PostResolveBlockerRequest {
  resolution_notes?: string;
}

export async function postResolveBlocker(params: {
  blockerUuid: string;
  body: PostResolveBlockerRequest;
}): Promise<GET<InlayBlocker>> {
  const res = await api.post(
    `/blocker/${params.blockerUuid}/resolve`,
    params.body,
  );
  return res.data;
}

export function postResolveBlockerOpts() {
  return mutationOptions({
    mutationFn: postResolveBlocker,
  });
}
