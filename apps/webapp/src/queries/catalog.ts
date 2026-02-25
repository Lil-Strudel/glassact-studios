import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { CatalogItem, GET, POST, PATCH } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getCatalogList(params?: {
  search?: string;
  category?: string;
  isActive?: boolean;
  limit?: number;
  offset?: number;
}): Promise<{
  items: GET<CatalogItem>[];
  total: number;
  limit: number;
  offset: number;
}> {
  const queryParams = new URLSearchParams();
  if (params?.search) queryParams.append("search", params.search);
  if (params?.category) queryParams.append("category", params.category);
  if (params?.isActive !== undefined)
    queryParams.append("is_active", String(params.isActive));
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const res = await api.get(
    `/catalog${queryParams.toString() ? "?" + queryParams.toString() : ""}`,
  );
  return res.data;
}

export function getCatalogListOpts(params?: {
  search?: string;
  category?: string;
  isActive?: boolean;
  limit?: number;
  offset?: number;
}) {
  return () =>
    queryOptions({
      queryKey: ["catalog", params],
      queryFn: () => getCatalogList(params),
    });
}

export async function getCatalogItem(uuid: string): Promise<GET<CatalogItem>> {
  const res = await api.get(`/catalog/${uuid}`);
  return res.data;
}

export function getCatalogItemOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["catalog", uuid],
      queryFn: () => getCatalogItem(uuid),
    });
}

export async function postCatalog(
  body: POST<CatalogItem>,
): Promise<GET<CatalogItem>> {
  const res = await api.post("/catalog", body);
  return res.data;
}

export function postCatalogOpts() {
  return mutationOptions({
    mutationFn: postCatalog,
  });
}

export async function patchCatalog(params: {
  uuid: string;
  body: PATCH<CatalogItem>;
}): Promise<GET<CatalogItem>> {
  const res = await api.patch(`/catalog/${params.uuid}`, params.body);
  return res.data;
}

export function patchCatalogOpts() {
  return mutationOptions({
    mutationFn: patchCatalog,
  });
}

export async function deleteCatalog(
  uuid: string,
): Promise<{ success: boolean }> {
  const res = await api.delete(`/catalog/${uuid}`);
  return res.data;
}

export function deleteCatalogOpts() {
  return mutationOptions({
    mutationFn: deleteCatalog,
  });
}

export async function getCatalogTags(uuid: string): Promise<string[]> {
  const res = await api.get(`/catalog/${uuid}/tags`);
  return res.data;
}

export function getCatalogTagsOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["catalog", uuid, "tags"],
      queryFn: () => getCatalogTags(uuid),
    });
}

export async function postCatalogTag(params: {
  uuid: string;
  tag: string;
}): Promise<string[]> {
  const res = await api.post(`/catalog/${params.uuid}/tags`, {
    tag: params.tag,
  });
  return res.data;
}

export function postCatalogTagOpts() {
  return mutationOptions({
    mutationFn: postCatalogTag,
  });
}

export async function deleteCatalogTag(params: {
  uuid: string;
  tag: string;
}): Promise<string[]> {
  const res = await api.delete(
    `/catalog/${params.uuid}/tags/${encodeURIComponent(params.tag)}`,
  );
  return res.data;
}

export function deleteCatalogTagOpts() {
  return mutationOptions({
    mutationFn: deleteCatalogTag,
  });
}
