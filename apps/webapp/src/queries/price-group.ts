import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { PriceGroup, GET, POST, PATCH } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getPriceGroups(params?: {
  limit?: number;
  offset?: number;
}): Promise<{
  items: GET<PriceGroup>[];
  total: number;
  limit: number;
  offset: number;
}> {
  const queryParams = new URLSearchParams();
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const res = await api.get(
    `/price-groups${queryParams.toString() ? "?" + queryParams.toString() : ""}`,
  );
  return res.data;
}

export function getPriceGroupsOpts(params?: {
  limit?: number;
  offset?: number;
}) {
  return queryOptions({
    queryKey: ["price-groups", params],
    queryFn: () => getPriceGroups(params),
  });
}

export async function getPriceGroup(uuid: string): Promise<GET<PriceGroup>> {
  const res = await api.get(`/price-groups/${uuid}`);
  return res.data;
}

export function getPriceGroupOpts(uuid: string) {
  return queryOptions({
    queryKey: ["price-groups", uuid],
    queryFn: () => getPriceGroup(uuid),
  });
}

export async function postPriceGroup(
  body: POST<PriceGroup>,
): Promise<GET<PriceGroup>> {
  const res = await api.post("/price-groups", body);
  return res.data;
}

export function postPriceGroupOpts() {
  return mutationOptions({
    mutationFn: postPriceGroup,
  });
}

export async function patchPriceGroup(
  uuid: string,
  body: PATCH<PriceGroup>,
): Promise<GET<PriceGroup>> {
  const res = await api.patch(`/price-groups/${uuid}`, body);
  return res.data;
}

export function patchPriceGroupOpts(uuid: string) {
  return mutationOptions({
    mutationFn: (body: PATCH<PriceGroup>) => patchPriceGroup(uuid, body),
  });
}

export async function deletePriceGroup(
  uuid: string,
): Promise<{ success: boolean }> {
  const res = await api.delete(`/price-groups/${uuid}`);
  return res.data;
}

export function deletePriceGroupOpts(uuid: string) {
  return mutationOptions({
    mutationFn: () => deletePriceGroup(uuid),
  });
}
