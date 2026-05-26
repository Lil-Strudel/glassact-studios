import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { Grout, GET } from "@glassact/data";

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
