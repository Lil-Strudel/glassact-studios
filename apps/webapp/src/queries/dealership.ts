import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Dealership, GET } from "@glassact/data";

export async function getDealerships(): Promise<GET<Dealership>[]> {
  const res = await api.get("/dealership");
  return res.data;
}

export function getDealershipsOpts() {
  return queryOptions({
    queryKey: ["dealership"],
    queryFn: getDealerships,
  });
}
