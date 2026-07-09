import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { Grout, GET, POST, PATCH } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getGrouts(): Promise<GET<Grout>[]> {
  const res = await api.get("/grouts");
  return res.data;
}

export function getGroutsOpts() {
  return queryOptions({
    queryKey: ["grouts"],
    queryFn: getGrouts,
    staleTime: 1000 * 60 * 30,
  });
}

export async function getGroutsAdmin(params?: {
  limit?: number;
  offset?: number;
}): Promise<{
  items: GET<Grout>[];
  total: number;
  limit: number;
  offset: number;
}> {
  const queryParams = new URLSearchParams();
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const res = await api.get(
    `/grouts/all${queryParams.toString() ? "?" + queryParams.toString() : ""}`,
  );
  return res.data;
}

export function getGroutsAdminOpts(params?: { limit?: number; offset?: number }) {
  return queryOptions({
    queryKey: ["grouts", "admin", params],
    queryFn: () => getGroutsAdmin(params),
  });
}

export async function postGrout(body: POST<Grout>): Promise<GET<Grout>> {
  const res = await api.post("/grouts", body);
  return res.data;
}

export function postGroutOpts() {
  return mutationOptions({
    mutationFn: postGrout,
  });
}

export async function patchGrout(
  uuid: string,
  body: PATCH<Grout>,
): Promise<GET<Grout>> {
  const res = await api.patch(`/grouts/${uuid}`, body);
  return res.data;
}

export function patchGroutOpts(uuid: string) {
  return mutationOptions({
    mutationFn: (body: PATCH<Grout>) => patchGrout(uuid, body),
  });
}

export async function deleteGrout(uuid: string): Promise<{ success: boolean }> {
  const res = await api.delete(`/grouts/${uuid}`);
  return res.data;
}

export function deleteGroutOpts(uuid: string) {
  return mutationOptions({
    mutationFn: () => deleteGrout(uuid),
  });
}
