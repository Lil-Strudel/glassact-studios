import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { InlayWithInfo, PATCH, Inlay } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getInlaysByProject(
  projectUuid: string,
): Promise<InlayWithInfo[]> {
  const res = await api.get(`/project/${projectUuid}/inlays`);
  return res.data;
}

export function getInlaysByProjectOpts(projectUuid: string) {
  return () =>
    queryOptions({
      queryKey: ["project", projectUuid, "inlays"],
      queryFn: () => getInlaysByProject(projectUuid),
    });
}

export async function getInlay(uuid: string): Promise<InlayWithInfo> {
  const res = await api.get(`/inlay/${uuid}`);
  return res.data;
}

export function getInlayOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["inlay", uuid],
      queryFn: () => getInlay(uuid),
    });
}

export interface PostCatalogInlayRequest {
  name: string;
  catalog_item_id: number;
  customization_notes?: string;
}

export async function postCatalogInlay(params: {
  projectUuid: string;
  body: PostCatalogInlayRequest;
}): Promise<InlayWithInfo> {
  const res = await api.post(
    `/project/${params.projectUuid}/inlays/catalog`,
    params.body,
  );
  return res.data;
}

export function postCatalogInlayOpts() {
  return mutationOptions({
    mutationFn: postCatalogInlay,
  });
}

export interface PostCustomInlayRequest {
  name: string;
  description: string;
  requested_width?: number;
  requested_height?: number;
}

export async function postCustomInlay(params: {
  projectUuid: string;
  body: PostCustomInlayRequest;
}): Promise<InlayWithInfo> {
  const res = await api.post(
    `/project/${params.projectUuid}/inlays/custom`,
    params.body,
  );
  return res.data;
}

export function postCustomInlayOpts() {
  return mutationOptions({
    mutationFn: postCustomInlay,
  });
}

export async function patchInlay(params: {
  uuid: string;
  body: { name?: string };
}): Promise<InlayWithInfo> {
  const res = await api.patch(`/inlay/${params.uuid}`, params.body);
  return res.data;
}

export function patchInlayOpts() {
  return mutationOptions({
    mutationFn: patchInlay,
  });
}

export async function deleteInlay(
  uuid: string,
): Promise<{ success: boolean }> {
  const res = await api.delete(`/inlay/${uuid}`);
  return res.data;
}

export function deleteInlayOpts() {
  return mutationOptions({
    mutationFn: deleteInlay,
  });
}
