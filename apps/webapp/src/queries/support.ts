import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type {
  GET,
  POST,
  PATCH,
  PriceGroup,
  SupportArticle,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getSupportArticles(): Promise<GET<SupportArticle>[]> {
  const res = await api.get("/support/articles");
  return res.data;
}

export function getSupportArticlesOpts() {
  return queryOptions({
    queryKey: ["support"],
    queryFn: getSupportArticles,
  });
}

export async function getSupportArticle(
  uuid: string,
): Promise<GET<SupportArticle>> {
  const res = await api.get(`/support/articles/${uuid}`);
  return res.data;
}

export function getSupportArticleOpts(uuid: string) {
  return queryOptions({
    queryKey: ["support", uuid],
    queryFn: () => getSupportArticle(uuid),
  });
}

export async function getSupportPriceGroups(): Promise<GET<PriceGroup>[]> {
  const res = await api.get("/support/price-groups");
  return res.data;
}

export function getSupportPriceGroupsOpts() {
  return queryOptions({
    queryKey: ["support", "price-groups"],
    queryFn: getSupportPriceGroups,
  });
}

export async function postSupportArticle(
  body: POST<SupportArticle>,
): Promise<GET<SupportArticle>> {
  const res = await api.post("/support/articles", body);
  return res.data;
}

export function postSupportArticleOpts() {
  return mutationOptions({
    mutationFn: postSupportArticle,
  });
}

export async function patchSupportArticle(params: {
  uuid: string;
  body: PATCH<SupportArticle>;
}): Promise<GET<SupportArticle>> {
  const res = await api.patch(`/support/articles/${params.uuid}`, params.body);
  return res.data;
}

export function patchSupportArticleOpts() {
  return mutationOptions({
    mutationFn: patchSupportArticle,
  });
}

export async function deleteSupportArticle(
  uuid: string,
): Promise<{ success: boolean }> {
  const res = await api.delete(`/support/articles/${uuid}`);
  return res.data;
}

export function deleteSupportArticleOpts() {
  return mutationOptions({
    mutationFn: deleteSupportArticle,
  });
}
