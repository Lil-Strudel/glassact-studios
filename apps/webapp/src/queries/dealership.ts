import { queryOptions, SolidMutationOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Dealership, GET, POST } from "@glassact/data";

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

export async function postDealership(
  body: POST<Dealership>,
): Promise<GET<Dealership>> {
  const res = await api.post("/dealership", body);
  return res.data;
}

export function postDealershipsOpts(): SolidMutationOptions<
  GET<Dealership>,
  Error,
  POST<Dealership>,
  unknown
> {
  return {
    mutationFn: postDealership,
  };
}
