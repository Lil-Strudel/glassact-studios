import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GlassColor, GET, POST, PATCH } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getGlassColors(): Promise<GET<GlassColor>[]> {
  const res = await api.get("/glass-colors");
  return res.data;
}

export function getGlassColorsOpts() {
  return queryOptions({
    queryKey: ["glass-colors"],
    queryFn: getGlassColors,
    staleTime: 1000 * 60 * 30,
  });
}

export async function getGlassColorsAdmin(params?: {
  limit?: number;
  offset?: number;
}): Promise<{
  items: GET<GlassColor>[];
  total: number;
  limit: number;
  offset: number;
}> {
  const queryParams = new URLSearchParams();
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const res = await api.get(
    `/glass-colors/all${queryParams.toString() ? "?" + queryParams.toString() : ""}`,
  );
  return res.data;
}

export function getGlassColorsAdminOpts(params?: {
  limit?: number;
  offset?: number;
}) {
  return queryOptions({
    queryKey: ["glass-colors", "admin", params],
    queryFn: () => getGlassColorsAdmin(params),
  });
}

export async function postGlassColor(
  body: POST<GlassColor>,
): Promise<GET<GlassColor>> {
  const res = await api.post("/glass-colors", body);
  return res.data;
}

export function postGlassColorOpts() {
  return mutationOptions({
    mutationFn: postGlassColor,
  });
}

export async function patchGlassColor(
  uuid: string,
  body: PATCH<GlassColor>,
): Promise<GET<GlassColor>> {
  const res = await api.patch(`/glass-colors/${uuid}`, body);
  return res.data;
}

export function patchGlassColorOpts(uuid: string) {
  return mutationOptions({
    mutationFn: (body: PATCH<GlassColor>) => patchGlassColor(uuid, body),
  });
}

export async function deleteGlassColor(
  uuid: string,
): Promise<{ success: boolean }> {
  const res = await api.delete(`/glass-colors/${uuid}`);
  return res.data;
}

export function deleteGlassColorOpts(uuid: string) {
  return mutationOptions({
    mutationFn: () => deleteGlassColor(uuid),
  });
}
