import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { CatalogItem, GET } from "@glassact/data";

// Browse catalog with filtering

export async function browseCatalog(params?: {
  category?: string;
  tags?: string[];
  search?: string;
  limit?: number;
  offset?: number;
}): Promise<{ items: GET<CatalogItem>[]; total: number; limit: number; offset: number }> {
  const queryParams = new URLSearchParams();
  if (params?.category) queryParams.append("category", params.category);
  if (params?.tags?.length) queryParams.append("tags", params.tags.join(","));
  if (params?.search) queryParams.append("search", params.search);
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const res = await api.get(
    `/catalog/browse${queryParams.toString() ? "?" + queryParams.toString() : ""}`
  );
  return res.data;
}

export function browseCatalogOpts(params?: {
  category?: string;
  tags?: string[];
  search?: string;
  limit?: number;
  offset?: number;
}) {
  return () =>
    queryOptions({
      queryKey: ["catalog-browse", params],
      queryFn: () => browseCatalog(params),
    });
}

// Get all categories

export async function getCatalogCategories(): Promise<string[]> {
  const res = await api.get("/catalog/categories");
  return res.data;
}

export function getCatalogCategoriesOpts() {
  return () =>
    queryOptions({
      queryKey: ["catalog-categories"],
      queryFn: getCatalogCategories,
    });
}

// Get all tags

export async function getCatalogAllTags(): Promise<string[]> {
  const res = await api.get("/catalog/tags");
  return res.data;
}

export function getCatalogAllTagsOpts() {
  return () =>
    queryOptions({
      queryKey: ["catalog-all-tags"],
      queryFn: getCatalogAllTags,
    });
}
