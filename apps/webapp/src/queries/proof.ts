import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, InlayProof } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getProofsByInlay(
  inlayUuid: string,
): Promise<GET<InlayProof>[]> {
  const res = await api.get(`/inlay/${inlayUuid}/proofs`);
  return res.data;
}

export function getProofsByInlayOpts(inlayUuid: string) {
  return queryOptions({
    queryKey: ["inlay", inlayUuid, "proofs"],
    queryFn: () => getProofsByInlay(inlayUuid),
  });
}

export async function getProof(uuid: string): Promise<GET<InlayProof>> {
  const res = await api.get(`/proof/${uuid}`);
  return res.data;
}

export function getProofOpts(uuid: string) {
  return queryOptions({
    queryKey: ["proof", uuid],
    queryFn: () => getProof(uuid),
  });
}

export interface CreateProofRequest {
  design_asset_url: string;
  width: number;
  height: number;
  price_group_id?: number;
  price_cents?: number;
  scale_factor?: number;
  color_overrides?: Record<string, unknown>;
}

export async function postProof(params: {
  inlayUuid: string;
  body: CreateProofRequest;
}): Promise<GET<InlayProof>> {
  const res = await api.post(
    `/inlay/${params.inlayUuid}/proofs`,
    params.body,
  );
  return res.data;
}

export function postProofOpts() {
  return mutationOptions({
    mutationFn: postProof,
  });
}

export async function postApproveProof(
  proofUuid: string,
): Promise<GET<InlayProof>> {
  const res = await api.post(`/proof/${proofUuid}/approve`);
  return res.data;
}

export function postApproveProofOpts() {
  return mutationOptions({
    mutationFn: postApproveProof,
  });
}

export async function postDeclineProof(params: {
  proofUuid: string;
  body: { decline_reason: string };
}): Promise<GET<InlayProof>> {
  const res = await api.post(
    `/proof/${params.proofUuid}/decline`,
    params.body,
  );
  return res.data;
}

export function postDeclineProofOpts() {
  return mutationOptions({
    mutationFn: postDeclineProof,
  });
}
