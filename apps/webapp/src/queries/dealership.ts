import { queryOptions, SolidMutationOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Dealership, GET, POST } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

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

export async function getDealership(uuid: string): Promise<GET<Dealership>> {
  const res = await api.get(`/dealership/${uuid}`);
  return res.data;
}

export function getDealershipOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["dealership", uuid],
      queryFn: () => getDealership(uuid),
    });
}

export async function postDealership(
  body: POST<Dealership>,
): Promise<GET<Dealership>> {
  const res = await api.post("/dealership", body);
  return res.data;
}

export function postDealershipsOpts() {
  return mutationOptions({
    mutationFn: postDealership,
  });
}
