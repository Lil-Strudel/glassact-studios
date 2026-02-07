import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { CatalogItem, GET, POST, PATCH } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

// Admin list queries

export async function getCatalogList(params?: {
  search?: string;
  category?: string;
  isActive?: boolean;
  limit?: number;
  offset?: number;
}): Promise<{ items: GET<CatalogItem>[]; total: number; limit: number; offset: number }> {
  const queryParams = new URLSearchParams();
  if (params?.search) queryParams.append("search", params.search);
  if (params?.category) queryParams.append("category", params.category);
  if (params?.isActive !== undefined) queryParams.append("is_active", String(params.isActive));
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const res = await api.get(
    `/catalog${queryParams.toString() ? "?" + queryParams.toString() : ""}`
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

// Single item

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

// Create

export async function postCatalog(body: POST<CatalogItem>): Promise<GET<CatalogItem>> {
  const res = await api.post("/catalog", body);
  return res.data;
}

export function postCatalogOpts() {
  return mutationOptions({
    mutationFn: postCatalog,
  });
}

// Update

export async function patchCatalog(
  uuid: string,
  body: PATCH<CatalogItem>
): Promise<GET<CatalogItem>> {
  const res = await api.patch(`/catalog/${uuid}`, body);
  return res.data;
}

export function patchCatalogOpts(uuid: string) {
  return mutationOptions({
    mutationFn: (body: PATCH<CatalogItem>) => patchCatalog(uuid, body),
  });
}

// Delete

export async function deleteCatalog(uuid: string): Promise<{ success: boolean }> {
  const res = await api.delete(`/catalog/${uuid}`);
  return res.data;
}

export function deleteCatalogOpts(uuid: string) {
  return mutationOptions({
    mutationFn: () => deleteCatalog(uuid),
  });
}

// Tags

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

export async function postCatalogTag(uuid: string, tag: string): Promise<string[]> {
  const res = await api.post(`/catalog/${uuid}/tags`, { tag });
  return res.data;
}

export function postCatalogTagOpts(uuid: string) {
  return mutationOptions({
    mutationFn: (tag: string) => postCatalogTag(uuid, tag),
  });
}

export async function deleteCatalogTag(uuid: string, tag: string): Promise<string[]> {
  const res = await api.delete(`/catalog/${uuid}/tags/${encodeURIComponent(tag)}`);
  return res.data;
}

export function deleteCatalogTagOpts(uuid: string, tag: string) {
  return mutationOptions({
    mutationFn: () => deleteCatalogTag(uuid, tag),
  });
}
