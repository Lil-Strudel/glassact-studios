import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { BakeRequest, BakeResult } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

// Fetched same-origin (not from the S3 redirect) so it can be read as text and
// rendered inline without CORS issues.
export async function getCatalogSvgText(uuid: string): Promise<string> {
  const res = await api.get(`/catalog/${uuid}/svg`, { responseType: "text" });
  return res.data as string;
}

export function getCatalogSvgTextOpts(uuid: string) {
  return queryOptions({
    queryKey: ["catalog", uuid, "svg"],
    queryFn: () => getCatalogSvgText(uuid),
    staleTime: 1000 * 60 * 30,
  });
}

export async function postBake(params: {
  uuid: string;
  body: BakeRequest;
}): Promise<BakeResult> {
  const res = await api.post(`/catalog/${params.uuid}/bake`, params.body);
  return res.data;
}

export function postBakeOpts() {
  return mutationOptions({
    mutationFn: postBake,
  });
}
