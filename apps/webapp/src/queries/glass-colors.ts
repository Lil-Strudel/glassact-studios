import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GlassColor, GET } from "@glassact/data";

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
